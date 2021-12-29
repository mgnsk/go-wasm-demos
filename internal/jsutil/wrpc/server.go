//go:build js && wasm
// +build js,wasm

package wrpc

import (
	"context"
	"sync/atomic"
	"syscall/js"

	"github.com/mgnsk/go-wasm-demos/internal/jsutil"
)

// Server handles MessagePort calls on a webworker.
type Server struct {
	callCount uint64
	sched     *Scheduler
}

// HandleMessages handles messages from port.
func (s *Server) HandleMessages(ctx context.Context, port Conn) {
	for {
		data, err := port.ReadRaw(ctx)
		if err != nil {
			panic(err)
		}

		switch {
		case !data.Get("start_scheduler").IsUndefined():
			target := NewMessagePort(data.Get("port"))
			// Announce that we have set up listeners for target port.
			if err := port.WriteRaw(ctx, map[string]interface{}{}, nil); err != nil {
				panic(err)
			}
			go s.HandleMessages(ctx, target)
			go func() {
				if err := s.sched.Run(ctx, target); err != nil {
					panic(err)
				}
			}()

		case !data.Get("rc").IsUndefined():
			call := NewCallFromJS(
				data.Get("rc"),
				data.Get("input"),
				data.Get("output"),
			)

			// Currently allow 1 concurrent call per worker.
			// TODO configure this on runtime.
			// It can happen if multiple ports are scheduling into this one.
			if atomic.LoadUint64(&s.callCount)+1 > 1 {
				jsutil.ConsoleLog("Rescheduling...")
				// Reschedule until we have a free worker.
				go func() {
					if err := s.sched.Call(context.TODO(), call); err != nil {
						panic(err)
					}
				}()
			} else {
				atomic.AddUint64(&s.callCount, 1)
				go func() {
					defer atomic.AddUint64(&s.callCount, ^uint64(0))
					call.Execute()
				}()
			}
		default:
			panic("invalid message type")
		}
	}
}

// Run the webworker MessagePort server.
func (s *Server) Run(ctx context.Context) {
	if !jsutil.IsWorker {
		panic("Must have webworker environment")
	}

	port := NewMessagePort(js.Global())
	if err := port.WriteRaw(ctx, map[string]interface{}{}, nil); err != nil {
		panic(err)
	}

	go s.HandleMessages(ctx, port)

	jsutil.ConsoleLog("Worker started")

	<-ctx.Done()
	panic(ctx.Err())
}

// NewServer creates a new server instance.
func NewServer() *Server {
	return &Server{
		sched: defaultScheduler,
	}
}
