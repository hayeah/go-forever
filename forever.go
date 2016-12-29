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
)

// Options customizes a forever superisor
type Options struct {
	MinUptime int
}

// Start a process that runs continuously
func Start(name string, args []string) {
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

	var waitForever chan interface{}
	waitForever <- struct{}{}
}
