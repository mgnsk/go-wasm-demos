package wrpc

import (
	"fmt"
	"runtime"
	"syscall/js"

	"github.com/mgnsk/go-wasm-demos/pkg/wrpcnet"
)

// Worker is a Web Worker wrapper.
type Worker struct {
	worker js.Value
	port   *wrpcnet.MessagePort
}

// Close the worker.
func (wk *Worker) Close() {
	wk.worker.Call("terminate")
}

// Call synchronously executes a remote call on the worker.
func (wk *Worker) Call(w, r *wrpcnet.MessagePort, name string) error {
	messages := map[string]any{
		"call": name,
		"w":    w.Value,
		"r":    r.Value,
	}

	var transferables []any

	if r == w {
		transferables = []any{w.Value}
	} else {
		transferables = []any{w.Value, r.Value}
	}

	return wk.port.WriteMessage(messages, transferables)
}

// NewWorker spawns a worker.
func NewWorker(url string) (*Worker, error) {
	worker := js.Global().Get("Worker").New(url)

	newWorker := &Worker{
		worker: worker,
		port:   wrpcnet.NewMessagePort(worker),
	}

	// Wait for the worker to be ready.
	if _, err := newWorker.port.ReadMessage(); err != nil {
		return nil, fmt.Errorf("error waiting for worker to become ready: %w", err)
	}

	runtime.SetFinalizer(newWorker, func(w any) {
		w.(*Worker).Close()
	})

	return newWorker, nil
}
