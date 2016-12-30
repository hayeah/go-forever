// Package forever impements similar features as https://github.com/foreverjs/forever
package forever

import (
	"context"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Options customizes a forever superisor
type Options struct {
	MinUptime   int
	RestartFile string
}

// monitoredProcess is supervised by the forever process
type monitoredProcess struct {
	Cmd *exec.Cmd
}

// // RunForever starts a process and runs it continuously.
// func (mp *monitoredProcess) RunForever() error {
// }

// // Restart sends SIGINT to supervised process, wait for exit, then spin up a process.
// func (mp *monitoredProcess) Restart() error {
// }

// // Restart sends SIGHUP to supervised process.
// func (mp *monitoredProcess) SoftRestart() error {
// }

// Start a process that runs continuously
func Start(name string, args []string, options *Options) {
	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, os.Interrupt)

	ctx := context.Background()
	cancelContext, cancel := context.WithCancel(ctx)

	var cmdGroup sync.WaitGroup
	var cmd *exec.Cmd
	go func() {
		for {
			cmd = exec.CommandContext(cancelContext, name, args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			cmdGroup.Add(1)
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
				cmdGroup.Done()
			}

			<-time.After(3 * time.Second)
			log.Println("Restarting process")
		}
		// cmd.ProcessState.Success
	}()

	go func() {
		select {
		case <-signalC:
			log.Println("Killing process")
			cancel()
			go func() {
				<-time.After(3 * time.Second)
				log.Println("Waiting for process to exit")
			}()
			cmdGroup.Wait()
			os.Exit(0)
		}
	}()

	if options.RestartFile != "" {
		rf := options.RestartFile
		_, err := os.Stat(rf)
		if os.IsNotExist(err) {
			f, err := os.Create(rf)
			if err != nil {
				log.Fatal("Failed to create restart file", err)
			}
			f.Close()
		}

		restartWatcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal("watch restart file:", err)
		}
		restartWatcher.Add(options.RestartFile)

		go func() {
			for event := range restartWatcher.Events {
				log.Println("restart watcher event", event)
				if event.Op == fsnotify.Chmod {
					err := cmd.Process.Signal(os.Interrupt)
					if err != nil {
						log.Println("Failed to signal process to restart:", err)
					}
				}
			}
		}()

	}

	var waitForever chan interface{}
	waitForever <- struct{}{}
}
