//go:build js && wasm

package wrpc

import (
	"fmt"
	"syscall/js"

	"github.com/mgnsk/go-wasm-demos/pkg/jsutil"
)

// Server is runs on a worker.
type Server struct {
	funcs map[string]WorkerFunc
}

// NewServer creates a new worker server.
func NewServer() *Server {
	return &Server{
		funcs: map[string]WorkerFunc{},
	}
}

// WithFunc registers a function to handle.
func (s *Server) WithFunc(name string, f WorkerFunc) *Server {
	s.funcs[name] = f
	return s
}

// ListenAndServe runs the server on worker.
func (s *Server) ListenAndServe() error {
	if !jsutil.IsWorker() {
		panic("server: must run in webworker environment")
	}

	port := NewMessagePort(js.Global())
	defer port.Close()

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
			s.execute(data)
		case !data.Get("__ping").IsUndefined():
		default:
			jsutil.ConsoleLog("server: invalid message", data)
		}
	}
}

func (s *Server) execute(data js.Value) {
	name := data.Get("call").String()
	r := NewMessagePort(data.Get("r"))
	w := NewMessagePort(data.Get("w"))
	defer w.Close()

	f, ok := s.funcs[name]
	if !ok {
		jsutil.ConsoleLog("server: WorkerFunc '%s' not found", name)
		return
	}

	f(w, r)
}
