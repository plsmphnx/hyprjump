# hyprjump

This is a simple tool for [Hyprland](https://hyprland.org) to jump over empty
workspaces that exist between occupied ones, only creating a new workspace after
progressing beyond the first or last workspace on the current monitor. It is
conceptually similar to [hyprnome](https://github.com/donovanglover/hyprnome),
but simplified and generalized in a way that fits my personal workflow. It is
written in zero-dependency Go.

# Usage

```
hyprjump [next/prev] [used/free] [dispatchers...]
```

-   `next`/`prev` - This sets the direction in which workspaces are selected.
    Defaults to `next`.
-   `used`/`free` - This forces the selection to remain on an occupied workspace
    (`used`) or jump to the next empty workspace (`free`). Defaults to neither.
-   `dispatchers` - Any argument that isn't one of the above keywords will be
    treated as a dispatcher to send to Hyprland, in order. If the dispatcher has
    arguments, any `@` will be replaced with the selected workspace ID; with no
    arguments, the workspace ID will be added as a single argument. If none are
    provided, this defaults to `workspace @`.

Note that the argument parser is fairly unopinionated for simplicity; if
multiple incompatible keywords are given, the last one will take effect.

# Examples

-   Go to the next occupied workspace: `hyprjump next used`
-   Move the current window to the previous workspace:
    `hyprjump movetoworkspace prev`
-   Open `foot` on a new workspace: `hyprjump free "exec [workspace @] foot"`
