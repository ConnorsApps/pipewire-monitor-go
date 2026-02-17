# Pipewire Monitor Golang
A wrapper for watching [Pipewire](https://docs.pipewire.org/) events using the CLI `pw-dump`

```sh
pw-dump ---monitor --no-colors
```

## Example

```golang
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	pwmonitor "github.com/ConnorsApps/pipewire-monitor-go"
)

// Only watch for nodes or removal events
func filter(e *pwmonitor.Event) bool {
	return e.Type == pwmonitor.EventNode || e.IsRemovalEvent()
}

func main() {
	var (
		ctx, cancel = context.WithCancel(context.Background())
		eventsChan  = make(chan []*pwmonitor.Event)
		sigChan     = make(chan os.Signal, 1)
		errChan     = make(chan error)
	)
	defer cancel()

	// Setup signal handling for graceful shutdown
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		errChan <- pwmonitor.Monitor(ctx, eventsChan, filter)
	}()

	for {
		select {
		case err := <-errChan:
			fmt.Println("pwmonitor.Monitor closed with an error", err)
			return
		case <-sigChan:
			fmt.Println("Received shutdown signal, cleaning up...")
			cancel()
			return
		case events := <-eventsChan:
			for _, e := range events {
				fmt.Println(e.Type, "id:", e.ID)
			}
		}
	}
}
```
