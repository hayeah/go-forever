package main

import (
	"github.com/hayeah/go-forever"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	minUptime       = kingpin.Flag("minUptime", "Minimum uptime for a script to not be considered 'spinning'").Duration()
	spinSleepTime   = kingpin.Flag("spinSleepTime", "Interval between restarts if a child is spinning").Duration()
	dropRestartFile = kingpin.Flag("dropRestartFile", "Touch the restart.txt to restart child.").Short('r').Bool()
	restartFile     = kingpin.Flag("restartFile", "Touch the file to restart child.").String()
	cmd             = kingpin.Arg("cmd", "command to run").Required().String()
	args            = kingpin.Arg("args", "arguments").Strings()
)

func main() {
	kingpin.Version("0.0.1")
	kingpin.Parse()

	if *dropRestartFile && *restartFile == "" {
		*restartFile = "restart.txt"
	}
	// kingpin.
	// log.Println(minUptime, spinSleepTime, *cmd, *args)
	forever.Start(*cmd, *args, &forever.Options{
		RestartFile: *restartFile,
	})
}
