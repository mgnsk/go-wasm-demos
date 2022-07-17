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
	port := NewMessagePort(js.Global())
	defer port.Close()

	// Notify the caller to start sending calls. We have established
	// an event listener for the worker port.
	if err := port.WriteMessage(map[string]any{}, nil); err != nil {
		return fmt.Errorf("server: error sending worker init ACK: %w", err)
	}

	for {
		data, err := port.ReadMessage()
		if err != nil {
			return fmt.Errorf("server: error reading from port: %w", err)
		}
		switch {
		case !data.Get("call").IsUndefined():
			if err := s.call(data); err != nil {
				return err
			}
		default:
			jsutil.ConsoleLog("server: invalid message", data)
		}
	}
}

func (s *Server) call(data js.Value) error {
	name := data.Get("call").String()
	r := NewMessagePort(data.Get("r"))
	w := NewMessagePort(data.Get("w"))
	defer w.Close()

	f, ok := s.funcs[name]
	if !ok {
		return fmt.Errorf("server: WorkerFunc '%s' not found", name)
	}

	if err := f(w, r); err != nil {
		return w.WriteError(err)
	}

	return nil
}
