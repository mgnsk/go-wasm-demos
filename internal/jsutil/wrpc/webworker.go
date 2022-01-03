//go:build js && wasm
// +build js,wasm

package wrpc

import (
	"fmt"
	"syscall/js"
)

// Worker is a browser thread.
type Worker struct {
	worker js.Value
	port   *MessagePort
}

// Terminate the webworker.
func (w *Worker) Terminate() {
	w.worker.Call("terminate")
}

// Call sends a remote call to be executed on the worker. Call returns when
// the the worker receives the call.
func (w *Worker) Call(call Call) error {
	messages, transferables := call.JSMessage()

	if err := w.port.WriteMessage(messages, transferables); err != nil {
		return fmt.Errorf("error sending call: %w", err)
	}
	if _, err := w.port.ReadMessage(); err != nil {
		return fmt.Errorf("error waiting for call to be received: %w", err)
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

	return newWorker, nil
}
