# Pipewire Monitor Golang
A wrapper for watching [Pipewire](https://docs.pipewire.org/) events using the CLI `pw-dump`

```sh
pw-dump ---monitor --no-colors
```

## Example

```golang
import (
	"context"
	"fmt"

	pwmonitor "github.com/ConnorsApps/pipewire-monitor-go"
)

// Only watch for nodes or removal events
func filter(e *pwmonitor.Event) bool {
	return e.Type == pwmonitor.EventNode || e.IsRemovalEvent()
}

func main() {
	var (
		ctx        = context.Background()
		eventsChan = make(chan []*pwmonitor.Event)
	)
	go func() {
		panic(pwmonitor.Monitor(ctx, eventsChan, filter))
	}()

	for {
		events := <-eventsChan
		for _, e := range events {
			fmt.Println(e.Type, "id:", e.ID)
		}
	}
}
```
