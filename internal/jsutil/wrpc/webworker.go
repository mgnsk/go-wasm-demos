//go:build js && wasm
// +build js,wasm

package wrpc

import (
	"context"
	"fmt"
	"syscall/js"
	"time"

	"github.com/mgnsk/go-wasm-demos/internal/jsutil"
)

// IndexJS boots up webworker go main.
var IndexJS []byte

// CreateTimeout specifies timeout for waiting for webworker hello.
var CreateTimeout = 3 * time.Second

// Worker is a browser thread that communicates through net.Conn interface.
type Worker struct {
	worker js.Value
	Port   Port
}

// NewWorkerFromSource creates a Worker from js source.
// The worker is terminated when context is canceled.
func NewWorkerFromSource(indexJS []byte) (*Worker, error) {
	url := jsutil.CreateURLObject(string(indexJS), "application/javascript")
	worker := js.Global().Get("Worker").New(url)

	w := &Worker{
		worker: worker,
	}

	// Create our side of port.
	w.Port = NewMessagePort(worker)

	ctx, cancel := context.WithTimeout(context.Background(), CreateTimeout)
	defer cancel()

	if _, err := w.Port.ReadRaw(ctx); err != nil {
		return nil, err
	}

	return w, nil
}

// JSValue returns the underlying js value.
func (w *Worker) JSValue() js.Value {
	return w.worker
}

// StartRemoteScheduler sends target port to worker and
// starts a scheduler remotely so that worker schedules
// wrpc calls to target.
func (w *Worker) StartRemoteScheduler(ctx context.Context, target Port) error {
	if err := w.Port.WriteRaw(
		ctx,
		map[string]interface{}{
			"start_scheduler": true,
			"port":            target,
		},
		target,
	); err != nil {
		return err
	}
	// Read the ACK.
	if _, err := target.ReadRaw(ctx); err != nil {
		return err
	}
	return nil
}

// Terminate the webworker.
func (w *Worker) Terminate() {
	w.worker.Call("terminate")
}

// TODO chrome needs high timeout, too slow for wasm
// as firefox just blazes. Needs testing.
const ackTimeout = 3 * time.Second

// WorkerRunner spawns webworkers.
type WorkerRunner struct {
	workers []*Worker
}

// Spawn a webworker.
func (r *WorkerRunner) Spawn(ctx context.Context) (*Worker, error) {
	newWorker, err := NewWorkerFromSource(IndexJS)
	if err != nil {
		return nil, err
	}

	// // Wait until the worker is ready.
	if _, err := newWorker.Port.ReadRaw(ctx); err != nil {
		return nil, fmt.Errorf("error waiting for worker to be ready: %w", err)
	}

	// Connect new worker with all running workers.
	for _, w := range r.workers {
		port1, port2 := Pipe()
		if err := newWorker.StartRemoteScheduler(ctx, port1); err != nil {
			return nil, fmt.Errorf("error starting remote scheduler: %w", err)
		}
		if err := w.StartRemoteScheduler(ctx, port2); err != nil {
			return nil, fmt.Errorf("error starting remote scheduler: %w", err)
		}
	}

	r.workers = append(r.workers, newWorker)

	go func() {
		if err := GlobalScheduler.Run(ctx, newWorker.Port); err != nil {
			panic(err)
		}
	}()

	return newWorker, nil
}
