package main

import (
	"github.com/hayeah/go-forever"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	// minUptime     = kingpin.Flag("minUptime", "Minimum uptime for a script to not be considered 'spinning'").Duration()
	spinSleepTime = kingpin.Flag("spinSleepTime", "Interval between restarts if a child is spinning").Default("3s").Duration()
	// dropRestartFile = kingpin.Flag("dropRestartFile", "Touch the restart.txt to restart child.").Short('r').Bool()
	restartFile = kingpin.Flag("restartFile", "Touch this file to restart child.").Default("restart.txt").String()
	cmd         = kingpin.Arg("cmd", "command to run").Required().String()
	args        = kingpin.Arg("args", "arguments").Strings()
)

func main() {
	kingpin.Version(forever.VERSION)
	kingpin.Parse()

	// kingpin.
	// log.Println(minUptime, spinSleepTime, *cmd, *args)
	forever.Start(*cmd, *args, &forever.Options{
		RestartFile:   *restartFile,
		SpinSleepTime: *spinSleepTime,
	})
}
