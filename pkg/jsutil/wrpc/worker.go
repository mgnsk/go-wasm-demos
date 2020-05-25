// +build js,wasm

package wrpc

import (
	"context"
	"syscall/js"

	"github.com/mgnsk/jsutil"
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

	// Wait for the first message to receive the messagePort on that
	// RPC calls from main thread are sent to.
	onmessage := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer ack(js.Global())

		data := args[0].Get("data")

		// Add the main thread port.
		mainPort := data.Get("main_port")
		if mainPort != js.Undefined() {
			// Set up the main port that receives commands from main thread.
			NewMessagePort(mainPort)

			return nil
			// Not scheduling to main thread from the worker.
		}

		// Start the scheduler to specified port.
		startScheduler := data.Get("start_scheduler")
		if jsutil.IsWorker && startScheduler != js.Undefined() {
			networkPort := data.Get("port")
			np := NewMessagePort(networkPort)

			// Start scheduling to the port until the port gets closed.
			go func() {
				if err := GlobalScheduler.RunScheduler(np.ctx, np); err != nil {
					jsutil.Dump("Scheduling stopped:", err)
					panic(err)
					// TODO
				}
			}()

			return nil
		}

		return nil
	})
	js.Global().Set("onmessage", onmessage)

	// Notify main thread that worker started.
	ack(js.Global())

	select {}
}
