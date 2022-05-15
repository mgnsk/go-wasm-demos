//go:build js && wasm
// +build js,wasm

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
func (w *Worker) Close() {
	w.worker.Call("terminate")
}

// Execute synchronously executes a remote call on the worker.
func (w *Worker) Execute(output, input *MessagePort, name string) error {
	messages := map[string]interface{}{
		"call":   name,
		"output": output.JSValue(),
		"input":  input.JSValue(),
	}

	transferables := []interface{}{output.JSValue()}
	if input != output {
		// Don't sent duplicate conn.
		transferables = append(transferables, input.JSValue())
	}

	if err := w.port.WriteMessage(messages, transferables); err != nil {
		return fmt.Errorf("error sending call: %w", err)
	}

	if err := w.port.WriteMessage(map[string]interface{}{"__ping": true}, nil); err != nil {
		return fmt.Errorf("error pinging worker: %w", err)
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

	runtime.SetFinalizer(newWorker, func(w interface{}) {
		w.(*Worker).Close()
	})

	return newWorker, nil
}
