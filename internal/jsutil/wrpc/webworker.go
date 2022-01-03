//go:build js && wasm
// +build js,wasm

package wrpc

import (
	"context"
	"fmt"
	"syscall/js"
	"time"
)

// Worker is a browser thread.
type Worker struct {
	worker js.Value
	port   ReadWriter
}

// Terminate the webworker.
func (w *Worker) Terminate() {
	w.worker.Call("terminate")
}

// Call sends a remote call to be executed on the worker. Call returns when
// the the worker receives the call.
func (w *Worker) Call(ctx context.Context, call Call) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	messages, transferables := call.JSMessage()

	if err := w.port.Write(ctx, messages, transferables); err != nil {
		return fmt.Errorf("error sending call: %w", err)
	}
	if _, err := w.port.Read(ctx); err != nil {
		return fmt.Errorf("error waiting for call to be received: %w", err)
	}

	return nil
}

// NewWorker spawns a worker with timeout.
func NewWorkerWithTimeout(url string, timeout time.Duration) (*Worker, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return NewWorker(ctx, url)
}

// NewWorker spawns a worker.
func NewWorker(ctx context.Context, url string) (*Worker, error) {
	worker := js.Global().Get("Worker").New(url)

	newWorker := &Worker{
		worker: worker,
		port:   NewMessagePort(worker),
	}

	// Wait for the worker to be ready.
	if _, err := newWorker.port.Read(ctx); err != nil {
		return nil, fmt.Errorf("error waiting for worker to become ready: %w", err)
	}

	return newWorker, nil
}
