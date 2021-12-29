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

// Worker is a browser thread that communicates through net.Conn interface.
type Worker struct {
	worker js.Value
	Port   Conn
}

// StartRemoteScheduler sends target port to worker and
// starts a scheduler remotely so that worker schedules
// wrpc calls to target.
func (w *Worker) StartRemoteScheduler(ctx context.Context, target js.Value) error {
	if err := w.Port.WriteRaw(
		ctx,
		map[string]interface{}{
			"start_scheduler": true,
			"port":            target,
		},
		[]interface{}{target},
	); err != nil {
		return err
	}
	// Note: we cannot read from target, the ownership has been moved.
	if _, err := w.Port.ReadRaw(ctx); err != nil {
		return err
	}
	return nil
}

// Terminate the webworker.
func (w *Worker) Terminate() {
	w.worker.Call("terminate")
}

// WorkerRunner spawns webworkers.
type WorkerRunner struct {
	sched   *Scheduler
	workers []*Worker
}

// NewWorkerRunner creates a new webworker runner.
func NewWorkerRunner() *WorkerRunner {
	return &WorkerRunner{sched: defaultScheduler}
}

// SpawnWithTimeout spawns a worker with timeout.
func (r *WorkerRunner) SpawnWithTimeout(indexJS string, timeout time.Duration) (*Worker, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return r.Spawn(ctx, indexJS)
}

// Spawn a webworker.
func (r *WorkerRunner) Spawn(ctx context.Context, indexJS string) (*Worker, error) {
	url := jsutil.CreateURLObject(indexJS, "application/javascript")
	worker := js.Global().Get("Worker").New(url)

	newWorker := &Worker{
		worker: worker,
	}

	// Create our side of port.
	newWorker.Port = NewMessagePort(worker)

	// Wait for the worker to be ready.
	if _, err := newWorker.Port.ReadRaw(ctx); err != nil {
		return nil, fmt.Errorf("error waiting for worker to become ready: %w", err)
	}

	// Connect new worker with all running workers.
	for _, w := range r.workers {
		port1, port2 := Pipe()
		if err := newWorker.StartRemoteScheduler(ctx, port1.JSValue()); err != nil {
			return nil, fmt.Errorf("error starting remote scheduler: %w", err)
		}
		if err := w.StartRemoteScheduler(ctx, port2.JSValue()); err != nil {
			return nil, fmt.Errorf("error starting remote scheduler: %w", err)
		}
	}

	r.workers = append(r.workers, newWorker)

	go func() {
		if err := r.sched.Run(context.TODO(), newWorker.Port); err != nil {
			panic(err)
		}
	}()

	return newWorker, nil
}
