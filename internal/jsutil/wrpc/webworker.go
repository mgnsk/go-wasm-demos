//go:build js && wasm
// +build js,wasm

package wrpc

import (
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
	Port   *MessagePort
}

// createWorkerFromSource creates a Worker from js source.
// The worker is terminated when context is canceled.
func createWorkerFromSource(indexJS []byte) (*Worker, error) {
	url := jsutil.CreateURLObject(string(indexJS), "application/javascript")
	worker := js.Global().Get("Worker").New(url)

	w := &Worker{
		worker: worker,
	}

	// Create our side of port.
	w.Port = NewMessagePort(worker)

	// Wait for the ACK signal.
	select {
	case <-w.Port.ack:
	case <-time.After(CreateTimeout):
		worker.Call("terminate")
		panic("Worker: ACK timeout")
	}

	return w, nil
}

// JSValue returns the underlying js value.
func (w *Worker) JSValue() js.Value {
	return w.worker
}

// StartRemoteScheduler starts a scheduler on the remote end
// that schedules to 'to'.
func (w *Worker) StartRemoteScheduler(to *MessagePort) {
	w.worker.Call(
		"postMessage",
		map[string]interface{}{
			"start_scheduler": true,
			"port":            to,
		},
		[]interface{}{to},
	)

	select {
	case <-w.Port.ack:
	case <-time.After(CreateTimeout):
		w.worker.Call("terminate")
		panic("Worker: ACK timeout")
	}
}

// Terminate the webworker.
func (w *Worker) Terminate() {
	w.worker.Call("terminate")
}
