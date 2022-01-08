//go:build js && wasm
// +build js,wasm

package wrpc

import (
	"fmt"
	"syscall/js"

	"github.com/mgnsk/go-wasm-demos/internal/jsutil"
)

// ListenAndServe runs the wrpc server on worker.
// The worker is single-threaded, the server blocks
// while a call is executing.
func ListenAndServe() error {
	if !jsutil.IsWorker() {
		panic("server: must run in webworker environment")
	}

	port := NewMessagePort(js.Global())
	// Notify the caller to start sending calls. We have established
	// and event listener for the worker port.
	if err := port.WriteMessage(map[string]interface{}{}, nil); err != nil {
		return fmt.Errorf("server: error sending worker init ACK: %w", err)
	}

	for {
		data, err := port.ReadMessage()
		if err != nil {
			return fmt.Errorf("server: error reading from port: %w", err)
		}
		switch {
		case !data.Get("call").IsUndefined():
			NewCallFromJS(data).Execute()
		case !data.Get("__ping").IsUndefined():
		default:
			jsutil.ConsoleLog("server: invalid message", data)
		}
	}
}
