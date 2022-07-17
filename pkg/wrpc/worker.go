package wrpc

import (
	"fmt"
	"runtime"
	"syscall/js"
)

// Worker is a Web Worker wrapper.
type Worker struct {
	worker js.Value
	port   *MessagePort
}

// Close the worker.
func (wk *Worker) Close() {
	wk.worker.Call("terminate")
}

// Call synchronously executes a remote call on the worker.
func (wk *Worker) Call(w, r *MessagePort, name string) error {
	messages := map[string]any{
		"call": name,
		"w":    w.value,
		"r":    r.value,
	}

	var transferables []any

	if r == w {
		transferables = []any{w.value}
	} else {
		transferables = []any{w.value, r.value}
	}

	if err := wk.port.WriteMessage(messages, transferables); err != nil {
		return fmt.Errorf("error calling worker: %w", err)
	}

	return nil
}

// NewWorker spawns a worker.
func NewWorker(url string) (*Worker, error) {
	worker := js.Global().Get("Worker").New(url)

	newWorker := &Worker{
		worker: worker,
		port:   NewMessagePort(worker),
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
