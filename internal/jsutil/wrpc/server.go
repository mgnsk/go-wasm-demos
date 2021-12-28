//go:build js && wasm
// +build js,wasm

package wrpc

import (
	"context"
	"syscall/js"

	"github.com/mgnsk/go-wasm-demos/internal/jsutil"
)

func ack(value js.Value) {
	value.Call("postMessage", map[string]interface{}{
		"ack": true,
	})
}

// RunServer runs on the webworker side to start the server implementing the WebRPC.
func RunServer(ctx context.Context) {
	if !jsutil.IsWorker {
		panic("Must have webworker environment")
	}

	jsutil.ConsoleLog("Worker started")

	onmessage := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer ack(js.Global())

		data := args[0].Get("data")

		// Start the scheduler to specified port.
		// This may be called by any thread.
		startScheduler := data.Get("start_scheduler")
		if jsutil.IsWorker && !startScheduler.IsUndefined() {
			jsutil.ConsoleLog("received port")

			networkPort := data.Get("port")
			np := NewMessagePort(networkPort)

			// Start scheduling to the port until the port gets closed.
			go func() {
				if err := GlobalScheduler.Run(ctx, np); err != nil {
					panic(err)
				}
			}()

			return nil
		}

		return nil
	})
	js.Global().Set("onmessage", onmessage)

	// Notify main thread that worker started.
	ack(js.Global())

	<-ctx.Done()
	panic(ctx.Err())
}
