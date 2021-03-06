# Install

```
go install github.com/hayeah/go-forever/forever
```

# Example

To run a process continuously:

```
forever ruby my-server.rb arg1 arg2 arg3
```

To restart:

```
touch restart.txt
```

# Go Forever

A CLI tool that runs a process continuously (i.e. forever). Similar to [forever.js](https://github.com/foreverjs/forever).

```
$ forever --help

usage: forever [<flags>] <cmd> [<args>...]

Flags:
  --help                       Show context-sensitive help (also try --help-long and --help-man).
  --spinSleepTime=3s           Interval between restarts if a child is spinning
  --restartFile="restart.txt"  Touch this file to restart child.
  --version                    Show application version.

Args:
  <cmd>     command to run
  [<args>]  arguments
```

