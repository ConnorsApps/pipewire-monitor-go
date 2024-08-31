package pwmonitor

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"time"

	// Use jsonv2 for pure streaming
	json_v2 "github.com/go-json-experiment/json"
)

var ContextCancelledError = errors.New("context cancelled")

// Monitor listens to pipewire events and sends them to the output channel
// Provide a filter function to remove events you're not interested in
func Monitor(ctx context.Context, output chan []*Event, filter ...func(*Event) bool) error {
	cmdErr := make(chan error)
	r, w := io.Pipe()
	defer r.Close()

	go func() {
		defer w.Close()
		cmd := exec.CommandContext(ctx, "pw-dump", "--monitor", "--no-colors")

		cmd.Stdout = w
		cmdErr <- cmd.Run()
	}()

	scan := bufio.NewScanner(r)

	for {
		select {
		case err := <-cmdErr:
			return fmt.Errorf("pw-dump --monitor: %w", err)

		case <-ctx.Done():
			return ContextCancelledError

		default:
			chunkReader, chunkWriter := io.Pipe()
			go func() {
				defer chunkWriter.Close()

				for scan.Scan() {
					out := scan.Bytes()
					chunkWriter.Write(out)
					// Reads until the end of the JSON array
					if len(out) == 1 && string(out) == "]" {
						return
					}
				}
			}()

			events := make([]*Event, 0, 10)

			if err := json_v2.UnmarshalRead(chunkReader, &events); err != nil {
				return fmt.Errorf("unmarshal event: %w", err)
			}

			var filtered = make([]*Event, 0, 10)

		EVENT_LOOP:
			for _, e := range events {
				for _, f := range filter {
					if !f(e) {
						continue EVENT_LOOP
					}
				}

				// Add handy timestamp
				e.CapturedAt = time.Now()
				filtered = append(filtered, e)
			}

			output <- filtered
		}
	}
}
