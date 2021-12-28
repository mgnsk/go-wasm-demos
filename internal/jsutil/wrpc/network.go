//go:build js && wasm
// +build js,wasm

package wrpc

import (
	"context"
	"time"

	"github.com/joomcode/errorx"
)

// TODO chrome needs high timeout, too slow for wasm
// as firefox just blazes. Needs testing.
const ackTimeout = 3 * time.Second

// WorkerRunner spawns webworkers.
type WorkerRunner struct {
	workers []*Worker
}

// Spawn a webworker.
func (r *WorkerRunner) Spawn(ctx context.Context) *Worker {
	newWorker, err := createWorkerFromSource(IndexJS)
	if err != nil {
		errorx.Panic(errorx.Decorate(err, "error creating worker"))
	}

	// Connect new worker with all running workers.
	for _, w := range r.workers {
		port1, port2 := Pipe()
		newWorker.StartRemoteScheduler(port1)
		w.StartRemoteScheduler(port2)
	}

	r.workers = append(r.workers, newWorker)

	// Connect our main thread and new worker.
	mainPort1, mainPort2 := Pipe()
	newWorker.StartRemoteScheduler(mainPort2)

	go func() {
		// Start scheduling to new worker.
		if err := GlobalScheduler.Run(ctx, mainPort1); err != nil {
			panic(err)
		}
	}()

	return newWorker
}

func factorial(n int) int {
	if n >= 1 {
		return n * factorial(n-1)
	}
	return 1
}

// A formula for the number of possible combinations of r objects from a set of n objects.
// C(n, r) = n! / (r!(n-r)!)
func cnr(n int, r int) int {
	cnr := factorial(n) / (factorial(r) * factorial(n-r))
	return cnr
}
