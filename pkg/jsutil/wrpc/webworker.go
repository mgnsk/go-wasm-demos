// +build js,wasm

package wrpc

import (
	"syscall/js"
	"time"

	"github.com/joomcode/errorx"
	"github.com/mgnsk/jsutil"
)

// IndexJS boots up webworker go main.
var IndexJS []byte

// CreateTimeout specifies timeout for waiting for webworker hello.
var CreateTimeout = 10 * time.Second

// Worker is a browser thread that communicates through net.Conn interface.
type Worker struct {
	worker                js.Value
	ack                   chan struct{}
	port                  *MessagePort
	remoteListenerStarted chan struct{}
}

// CreateWorkerFromSource creates a Worker from js source.
// The worker is terminated when context is canceled.
func CreateWorkerFromSource(indexJS []byte) (*Worker, error) {
	url := jsutil.CreateURLObject(string(indexJS), "application/javascript")
	worker := js.Global().Get("Worker").New(url)

	w := &Worker{
		worker:                worker,
		ack:                   make(chan struct{}),
		remoteListenerStarted: make(chan struct{}),
	}

	onmessage := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			w.ack <- struct{}{}
		}()
		return nil
	})
	worker.Set("onmessage", onmessage)
	//defer onmessage.Release()

	// Wait for the ACK signal.
	select {
	case <-w.ack:
	case <-time.After(CreateTimeout):
		worker.Call("terminate")
		return nil, errorx.TimeoutElapsed.New("ACK timeout: waited for worker to be ready in %s", CreateTimeout)
	}

	// As the first message, we are sending a MessagePort
	// on there to continue further communication.
	messageChannel := js.Global().Get("MessageChannel").New()

	// Create our side of port.
	w.port = NewMessagePort(messageChannel.Get("port1"))

	// Send port2 and transfer the ownership to the worker.
	port2 := messageChannel.Get("port2")
	message := map[string]interface{}{
		"main_port": port2,
	}
	transfer := []interface{}{
		port2,
	}
	worker.Call("postMessage", message, transfer)

	// Wait for the worker to acknowledge it received the port.
	select {
	case <-w.ack:
	case <-time.After(CreateTimeout):
		worker.Call("terminate")
		return nil, errorx.TimeoutElapsed.New("ACK timeout: waited for port received ack")
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
	messages := map[string]interface{}{
		"start_scheduler": true,
		"port":            to,
	}
	transferables := []interface{}{to}
	w.JSValue().Call("postMessage", messages, transferables)
}

// ACK channel.
func (w *Worker) ACK() <-chan struct{} {
	return w.ack
}

// MessagePort returns the worker's port.
func (w *Worker) MessagePort() *MessagePort {
	return w.port
}

// // RemoteListenerReady channel is closed when the remote listener accept is called.
// func (w *Worker) RemoteListenerReady() <-chan struct{} {
// 	return w.remoteListenerStarted
// }

// Terminate the webworker.
func (w *Worker) Terminate() {
	w.worker.Call("terminate")
}
