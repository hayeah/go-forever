// Package forever impements similar features as https://github.com/foreverjs/forever
package forever

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Options customizes a forever superisor
type Options struct {
	MinUptime   int
	RestartFile string
}

// NewMonitoredProcess creates a new monitored process
func NewMonitoredProcess(name string, args []string) (mp *MonitoredProcess) {
	mp = &MonitoredProcess{
		Name: name,
		Args: args,
	}

	return
}

// MonitoredProcess is supervised by the forever process
type MonitoredProcess struct {
	Name string
	Args []string

	// The currently running supervised process
	cmd *exec.Cmd
	// Indicate that supervisor should stop
	stopRestart bool
}

// RunForever starts a process and runs it continuously.
func (mp *MonitoredProcess) RunForever() {
	for {
		cmd := exec.Command(mp.Name, mp.Args...)
		mp.cmd = cmd

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// cmdGroup.Add(1)
		err := cmd.Start()
		if err != nil {
			log.Println("cmd failed to start:", cmd)
		} else {
			log.Printf("Process running: %d\n", cmd.Process.Pid)
			err = cmd.Wait()
			if err != nil {
				log.Printf("Wait(%d): %s", cmd.Process.Pid, err)
				// panic(err)
			}
			// cmdGroup.Done()
		}

		if mp.stopRestart {
			return
		}

		<-time.After(3 * time.Second)
		log.Println("Restarting process")
	}
}

// Restart sends SIGINT to supervised process, wait for exit, then spin up a process.
func (mp *MonitoredProcess) Restart() error {
	return mp.sendInterrupt()
}

// Stop kills the supervised process, causing RunForever to end
func (mp *MonitoredProcess) Stop() (ps *os.ProcessState, err error) {
	mp.stopRestart = true

	err = mp.sendInterrupt()
	if err != nil {
		return
	}

	return mp.cmd.Process.Wait()
}

func (mp *MonitoredProcess) sendInterrupt() error {
	return mp.cmd.Process.Signal(os.Interrupt)
}

// // Restart sends SIGHUP to supervised process.
// func (mp *monitoredProcess) SoftRestart() error {
// }

// Supervisor is the supervising process
type Supervisor struct {
	Child   *MonitoredProcess
	Options *Options
}

func (s *Supervisor) handleInterrupt() (ps *os.ProcessState, err error) {
	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, os.Interrupt)

	select {
	case <-signalC:
		return s.Child.Stop()
	}
}

func (s *Supervisor) watchRestartFile() (err error) {
	rf := s.Options.RestartFile
	if rf == "" {
		return
	}

	_, err = os.Stat(rf)
	if os.IsNotExist(err) {
		f, err := os.Create(rf)
		if err != nil {
			return err
			// log.Fatal("Failed to create restart file", err)
		}
		f.Close()
	}

	if err != nil {
		return
	}

	restartWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return
		// log.Fatal("watch restart file:", err)
	}
	restartWatcher.Add(rf)

	for event := range restartWatcher.Events {
		if event.Op == fsnotify.Chmod {
			err := s.Child.Restart()
			if err != nil {
				log.Println("Failed to signal process to restart:", err)
			}
		}
	}

	return
}

// Start a process that runs continuously
func Start(name string, args []string, options *Options) {
	mp := NewMonitoredProcess(name, args)

	s := &Supervisor{
		Child:   mp,
		Options: options,
	}

	go mp.RunForever()

	go func() {
		err := s.watchRestartFile()
		if err != nil {
			log.Fatal("Restart file", err)
		}
	}()

	go func() {
		s.handleInterrupt()
		os.Exit(0)
	}()

	var waitForever chan interface{}
	waitForever <- struct{}{}
}
