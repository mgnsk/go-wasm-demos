//go:build js && wasm
// +build js,wasm

package wrpc

import (
	"context"
	"fmt"
	"sync/atomic"
	"syscall/js"
	"time"

	"github.com/mgnsk/go-wasm-demos/internal/jsutil"
)

// Server handles MessagePort calls on a webworker.
type Server struct {
	callCount uint64
	sched     *Scheduler
}

func ack(ctx context.Context, port Writer) error {
	return port.Write(ctx, map[string]interface{}{"ack": true}, nil)
}

// HandleMessages handles messages from port.
func (s *Server) HandleMessages(ctx context.Context, port ReadWriter) {
	for {
		data, err := port.Read(ctx)
		if err != nil {
			panic(err)
		}

		switch {
		case !data.Get("start_scheduler").IsUndefined():
			target := NewMessagePort(data.Get("port"))
			if err := ack(ctx, port); err != nil {
				panic(err)
			}
			s.sched.Register(target)
			go s.HandleMessages(ctx, target)
		case !data.Get("rc").IsUndefined():
			call := NewCallFromJS(data)
			if err := ack(ctx, port); err != nil {
				panic(err)
			}

			go func() {
				// Currently allow 1 concurrent call per worker.
				// TODO configure this on runtime.
				// It can happen if multiple ports are scheduling into this one.
				if atomic.AddUint64(&s.callCount, 1) > 1 {
					jsutil.ConsoleLog("Rescheduling...")
					// Reschedule until we have a free worker.
					ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
					defer cancel()

					if err := s.sched.Call(ctx, call); err != nil {
						panic(fmt.Errorf("error rescheduling: %w", err))
					}
				} else {
					defer atomic.AddUint64(&s.callCount, ^uint64(0))
					call.Execute()
				}
			}()
		case !data.Get("ack").IsUndefined():
		default:
			panic("Server: invalid message")
		}
	}
}

// Run the webworker MessagePort server.
func (s *Server) Run(ctx context.Context) {
	if !jsutil.IsWorker {
		panic("Must have webworker environment")
	}

	port := NewMessagePort(js.Global())
	if err := port.Write(ctx, map[string]interface{}{}, nil); err != nil {
		panic(err)
	}

	go s.HandleMessages(ctx, port)

	jsutil.ConsoleLog("Server started")

	<-ctx.Done()
	panic(ctx.Err())
}

// NewServer creates a new server instance.
func NewServer() *Server {
	return &Server{
		sched: defaultScheduler,
	}
}
