// Package forever impements similar features as https://github.com/foreverjs/forever
package forever

import (
	"os"
	"os/exec"
	"os/signal"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/fsnotify/fsnotify"
)

// VERSION is the semantic version string
const VERSION = "0.0.2"

// Options customizes a forever superisor
type Options struct {
	// MinUptime     int
	SpinSleepTime time.Duration
	RestartFile   string
}

// Supervisor is the supervising process
type Supervisor struct {
	Options *Options

	// The currently running supervised process
	child *exec.Cmd

	// Child process id
	childPid int

	// Indicate that supervisor should stop
	stopRestart bool

	// Indicate that this is a requested restart
	restartRequested bool
}

// Supervise starts a process and runs it continuously.
func (s *Supervisor) Supervise(bin string, args []string) {
	// clog := log.WithFields(log.Fields{
	// 	"bin": bin,
	// })

	log.WithFields(log.Fields{
		"bin":  bin,
		"args": args,
	}).Info("Supervising child")
	for {
		if s.stopRestart {
			return
		}

		child := exec.Command(bin, args...)
		s.child = child

		child.Stdout = os.Stdout
		child.Stderr = os.Stderr

		err := child.Start()
		if err != nil {
			log.WithField("err", err).Info("Failed to start child")
		} else {
			pid := child.Process.Pid
			s.childPid = pid
			log.WithField("pid", pid).Info("Child running")
			err = child.Wait()
			log.WithFields(log.Fields{
				"pid": pid,
				"err": err,
			}).Info("Child terminated")
			s.childPid = 0
		}

		if s.stopRestart {
			return
		}

		// If is user requested restart, don't wait for spin sleep time
		if s.restartRequested {
			// reset the flag
			s.restartRequested = false
		} else {
			interval := s.Options.SpinSleepTime
			log.WithField("interval", interval).Info("Waiting to restart")
			<-time.After(interval)
		}
	}
}

// Restart sends SIGINT to supervised process, wait for exit, then spin up a process.
func (s *Supervisor) Restart() error {
	s.restartRequested = true
	return s.interruptChild()
}

// Stop kills the supervised process, causing RunForever to end
func (s *Supervisor) Stop() (ps *os.ProcessState, err error) {
	log.Info("Stopping superisor")

	s.stopRestart = true

	err = s.interruptChild()
	if err != nil {
		return
	}

	return s.child.Process.Wait()
}

func (s *Supervisor) interruptChild() error {
	if s.childPid == 0 {
		log.Info("No running child to interrupt")
		return nil
	}

	log.WithFields(log.Fields{
		"pid": s.childPid,
	}).Info("Interrupt child")

	return s.child.Process.Signal(os.Interrupt)
}

func (s *Supervisor) handleInterrupt() (ps *os.ProcessState, err error) {
	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, os.Interrupt)

	select {
	case <-signalC:
		return s.Stop()
	}
}

func (s *Supervisor) watchRestartFile() (err error) {
	rf := s.Options.RestartFile

	_, err = os.Stat(rf)
	if os.IsNotExist(err) {
		f, err := os.Create(rf)
		if err != nil {
			return err
		}
		f.Close()
	}

	if err != nil {
		return
	}

	restartWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	restartWatcher.Add(rf)

	for event := range restartWatcher.Events {
		if event.Op == fsnotify.Chmod {
			err := s.Restart()
			if err != nil {
				log.Println("Failed to signal process to restart:", err)
			}
		}
	}

	return
}

// Start a process that runs continuously
func Start(bin string, args []string, options *Options) {
	s := &Supervisor{
		Options: options,
	}

	go func() {
		err := s.watchRestartFile()
		if err != nil {
			log.Fatal("Restart file", err)
		}
	}()

	go func() {
		// This handler can cause forever loop to stop
		s.handleInterrupt()
	}()

	// forever loop
	s.Supervise(bin, args)

	os.Exit(0)

	// var waitForever chan interface{}
	// waitForever <- struct{}{}
}
