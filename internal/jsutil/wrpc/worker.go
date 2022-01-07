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

// Call a remote call on the worker. Call returns as soon as the worker
// receives the call. Since the worker is single-threaded, Ping can be
// used to wait for the current call.
func (w *Worker) Call(call *Call) error {
	messages, transferables := call.JSMessage()

	if err := w.port.WriteMessage(messages, transferables); err != nil {
		return fmt.Errorf("error sending call: %w", err)
	}
	if _, err := w.port.ReadMessage(); err != nil {
		return fmt.Errorf("error waiting for call to be received: %w", err)
	}

	return nil
}

// Ping the worker. If the worker is busy, it blocks until the worker responds.
func (w *Worker) Ping() error {
	if err := w.port.WriteMessage(map[string]interface{}{"ping": true}, nil); err != nil {
		return fmt.Errorf("error pinging worker: %w", err)
	}
	if _, err := w.port.ReadMessage(); err != nil {
		return fmt.Errorf("error waiting for worker ping reply: %w", err)
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
