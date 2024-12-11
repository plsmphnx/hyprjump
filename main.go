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

type IPC string

type Workspace struct {
	ID        int `json:"id"`
	MonitorID int `json:"monitorID"`
	Windows   int `json:"windows"`
	i         int
}

func main() {
	defer func() {
		if e := recover(); e != nil {
			fmt.Fprintln(os.Stderr, e)
		}
	}()

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

	ipc := IPC(fmt.Sprintf("%s/hypr/%s/.socket.sock",
		os.Getenv("XDG_RUNTIME_DIR"),
		os.Getenv("HYPRLAND_INSTANCE_SIGNATURE"),
	))

	var active Workspace
	var workspaces []Workspace
	info := ipc.Call("j/activeworkspace", "j/workspaces")
	check(json.Unmarshal(info[0], &active))
	check(json.Unmarshal(info[1], &workspaces))
	slices.SortFunc(workspaces, func(a, b Workspace) int { return a.ID - b.ID })

	var monitor []Workspace
	for i, ws := range workspaces {
		ws.i = i
		if ws.MonitorID == active.MonitorID {
			if ws.ID == active.ID {
				active.i = len(monitor)
			}
			monitor = append(monitor, ws)
		}
	}

	id := max(1, min(math.MaxInt32, func() int {
		if prev {
			if active.i == 0 || free > 0 {
				tgt := monitor[0]
				if free < 0 || tgt.Windows == 0 {
					return tgt.ID
				}
				for tgt.i >= 0 && workspaces[tgt.i].ID == tgt.ID {
					tgt.i--
					tgt.ID--
				}
				return tgt.ID
			}
			return monitor[active.i-1].ID
		} else {
			if active.i == len(monitor)-1 || free > 0 {
				tgt := monitor[len(monitor)-1]
				if free < 0 || tgt.Windows == 0 {
					return tgt.ID
				}
				for tgt.i < len(workspaces) && workspaces[tgt.i].ID == tgt.ID {
					tgt.i++
					tgt.ID++
				}
				return tgt.ID
			}
			return monitor[active.i+1].ID
		}
	}()))

	for i, d := range disp {
		d = strings.Replace(d, "@", fmt.Sprint(id), -1)
		disp[i] = "/dispatch " + strings.TrimSpace(d)
	}
	for _, res := range ipc.Call(disp...) {
		if string(res) != "ok" {
			fmt.Fprintln(os.Stderr, string(res))
		}
	}
}

func (ipc IPC) Call(cmds ...string) [][]byte {
	conn := must(net.DialUnix("unix", nil, &net.UnixAddr{Name: string(ipc)}))
	defer func() { check(conn.Close()) }()

	cmd := cmds[0]
	if len(cmds) > 1 {
		cmd = "[[BATCH]]" + strings.Join(cmds, ";")
	}

	io.WriteString(conn, cmd)
	return bytes.Split(must(io.ReadAll(conn)), []byte{'\n', '\n', '\n'})
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
