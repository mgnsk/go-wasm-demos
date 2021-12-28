//go:build js && wasm
// +build js,wasm

package wrpc

import (
	"context"
	"sync/atomic"
	"syscall/js"

	"github.com/mgnsk/go-wasm-demos/internal/jsutil"
)

var callCount uint64

func HandleMessages(ctx context.Context, r RawReader) {
	for {
		data, err := r.ReadRaw(ctx)
		if err != nil {
			panic(err)
		}

		jsutil.ConsoleLog("got message", data)

		switch true {
		// Start the scheduler to specified port.
		case !data.Get("start_scheduler").IsUndefined():
			target := NewMessagePort(data.Get("port"))
			go HandleMessages(ctx, target)
			go func() {
				if err := GlobalScheduler.Run(ctx, target); err != nil {
					panic(err)
				}
			}()

			if err := target.WriteRaw(ctx, map[string]interface{}{}); err != nil {
				panic(err)
			}

		case !data.Get("rc").IsUndefined():
			call := NewCallFromJS(
				data.Get("rc"),
				data.Get("input"),
				data.Get("output"),
			)

			// Currently allow 1 concurrent call per worker.
			// TODO configure this on runtime.
			// It can happen if multiple ports are scheduling into this one.
			if atomic.AddUint64(&callCount, 1) > 1 {
				jsutil.ConsoleLog("Rescheduling...")
				// Reschedule until we have a free worker.
				go func() {
					if err := GlobalScheduler.Call(context.TODO(), call); err != nil {
						panic(err)
					}
				}()
			} else {
				go call.Execute()
			}
		}
	}
}

// RunServer runs on the webworker side to start the server implementing the WebRPC.
func RunServer(ctx context.Context) {
	if !jsutil.IsWorker {
		panic("Must have webworker environment")
	}

	port := NewMessagePort(js.Global())
	go HandleMessages(ctx, port)

	if err := port.WriteRaw(ctx, map[string]interface{}{}); err != nil {
		panic(err)
	}

	jsutil.ConsoleLog("Worker started")

	<-ctx.Done()
	panic(ctx.Err())
}
