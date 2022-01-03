//go:build js && wasm
// +build js,wasm

package wrpc

import (
	"context"
	"fmt"
	"syscall/js"

	"github.com/mgnsk/go-wasm-demos/internal/jsutil"
)

// ListenAndServe runs the wrpc server on worker.
func ListenAndServe(ctx context.Context) error {
	if !jsutil.IsWorker() {
		panic("server: must run in webworker environment")
	}

	port := NewMessagePort(js.Global())
	if err := port.Write(ctx, map[string]interface{}{}, nil); err != nil {
		return fmt.Errorf("server: error sending init ACK: %w", err)
	}

	for {
		data, err := port.Read(ctx)
		if err != nil {
			return fmt.Errorf("server: error reading from port: %w", err)
		}
		switch {
		case !data.Get("call").IsUndefined():
			call := NewCallFromJS(data)
			if err := port.Write(ctx, map[string]interface{}{"received": true}, nil); err != nil {
				panic(err)
			}
			go call.ExecuteLocal()
		case !data.Get("ack").IsUndefined():
		default:
			jsutil.ConsoleLog(data)
			panic("Server: invalid message")
		}
	}
}
