package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"slices"
	"strings"
)

type Workspace struct {
	ID        int `json:"id"`
	MonitorID int `json:"monitorID"`
	Windows   int `json:"windows"`

	w int
	m int
}

func main() {
	defer exit()

	var disp []string
	var prev bool
	var free int8

	for _, arg := range os.Args[1:] {
		switch arg {
		case "next":
			prev = false
		case "prev":
			prev = true
		case "used":
			free = -1
		case "free":
			free = 1
		default:
			if strings.IndexByte(arg, ' ') < 0 {
				disp = append(disp, arg+" @")
			} else {
				disp = append(disp, arg)
			}
		}
	}

	if len(disp) == 0 {
		disp = []string{"workspace @"}
	}

	var active Workspace
	var workspaces []Workspace
	info := ipc("j/activeworkspace", "j/workspaces")
	check(json.Unmarshal(info[0], &active))
	check(json.Unmarshal(info[1], &workspaces))
	slices.SortFunc(workspaces, func(a, b Workspace) int { return a.ID - b.ID })

	var monitor []Workspace
	for i, ws := range workspaces {
		ws.w = i
		if ws.MonitorID == active.MonitorID {
			ws.m = len(monitor)
			monitor = append(monitor, ws)
		}
		if ws.ID == active.ID {
			active.w = ws.w
			active.m = ws.m
		}
	}

	id := max(1, min(math.MaxInt32, func() int {
		if prev {
			if active.m == 0 || free > 0 {
				tgt := monitor[0]
				if free < 0 || tgt.Windows == 0 {
					return tgt.ID
				}
				for tgt.w >= 0 && workspaces[tgt.w].ID == tgt.ID {
					tgt.w--
					tgt.ID--
				}
				return tgt.ID
			}
			return monitor[active.m-1].ID
		} else {
			if active.m == len(monitor)-1 || free > 0 {
				tgt := monitor[len(monitor)-1]
				if free < 0 || tgt.Windows == 0 {
					return tgt.ID
				}
				for tgt.w < len(workspaces) && workspaces[tgt.w].ID == tgt.ID {
					tgt.w++
					tgt.ID++
				}
				return tgt.ID
			}
			return monitor[active.m+1].ID
		}
	}()))

	for i, d := range disp {
		if strings.IndexByte(d, '@') < 0 {
			disp[i] = "/dispatch " + d
		} else {
			d = strings.Replace(d, "@", "%d", 1)
			disp[i] = fmt.Sprintf("/dispatch "+d, id)
		}
	}
	for _, res := range ipc(disp...) {
		if string(res) != "ok" {
			fmt.Fprintln(os.Stderr, string(res))
		}
	}
}

func ipc(cmds ...string) [][]byte {
	conn := must(net.DialUnix("unix", nil, &net.UnixAddr{
		Name: fmt.Sprintf("%s/hypr/%s/.socket.sock",
			os.Getenv("XDG_RUNTIME_DIR"),
			os.Getenv("HYPRLAND_INSTANCE_SIGNATURE"),
		),
	}))
	defer func() { check(conn.Close()) }()

	var cmd string
	if len(cmds) == 1 {
		cmd = cmds[0]
	} else {
		cmd = "[[BATCH]]" + strings.Join(cmds, ";")
	}

	io.WriteString(conn, cmd)
	return bytes.Split(must(io.ReadAll(conn)), []byte{'\n', '\n', '\n'})
}

func exit() {
	if e := recover(); e != nil {
		fmt.Fprintln(os.Stderr, e)
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func must[T any](t T, e error) T {
	check(e)
	return t
}
