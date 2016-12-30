# Go Forever

A CLI tool that runs a process continuously (i.e. forever). Similar to [forever.js](https://github.com/foreverjs/forever).

```
go run cli/forever.go --help
usage: forever [<flags>] <cmd> [<args>...]

Flags:
      --help                     Show context-sensitive help (also try --help-long and --help-man).
  -r, --dropRestartFile          Touch the restart.txt to restart child.
      --restartFile=RESTARTFILE  Touch the file to restart child.
      --version                  Show application version.

Args:
  <cmd>     command to run
  [<args>]  arguments
```

# Example

To run a process continuously:

```
forever --restartFile=restart.txt ruby my-server.rb arg1 arg2 arg3
```

To restart:

```
touch restart.txt
```