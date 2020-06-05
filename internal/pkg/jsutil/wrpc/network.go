// +build js,wasm

package wrpc

import (
	"context"
	"syscall/js"
	"time"

	"github.com/joomcode/errorx"
)

// TODO chrome needs high timeout, too slow for wasm
// as firefox just blazes. Needs testing.
const ackTimeout = 3 * time.Second

var (
	links   = 0
	workers []*Worker
)

// SpawnWorker spawns and connects a new webworker.
func SpawnWorker(ctx context.Context) *Worker {
	newWorker, err := CreateWorkerFromSource(IndexJS)
	if err != nil {
		errorx.Panic(errorx.Decorate(err, "error creating worker"))
	}

	linkDone := make(chan struct{})

	// Add links between this and previous workers.
	for _, w := range workers {
		existingWorker := w

		messageChannel := js.Global().Get("MessageChannel").New()

		port1 := NewMessagePort(messageChannel.Get("port1"))
		port2 := NewMessagePort(messageChannel.Get("port2"))

		// Connect the two workers by starting event listeners and schedulers
		// on both sides so they can communicate.
		newWorker.StartRemoteScheduler(port1)
		existingWorker.StartRemoteScheduler(port2)

		links++

		go func() {
			select {
			case <-newWorker.ACK():
			case <-time.After(ackTimeout):
				panic("ack timeout")
			}
			select {
			case <-existingWorker.ACK():
			case <-time.After(ackTimeout):
				panic("ack timeout")
			}
			linkDone <- struct{}{}
		}()
	}

	go func() {
		// Start scheduling to new worker.
		if err := GlobalScheduler.RunScheduler(ctx, newWorker.MessagePort()); err != nil {
			panic(err)
		}
	}()

	workers = append(workers, newWorker)

	if len(workers) > 1 {
		// Wait for all the new links we created.
		combinations := cnr(len(workers), 2)
		for i := 0; i < combinations-links; i++ {
			<-linkDone
		}
	}

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
