// +build js,wasm

package wrpc

import (
	"context"
	"io"
	"runtime"
	"sync/atomic"
	"syscall/js"

	"github.com/joomcode/errorx"
	"github.com/mgnsk/jsutil"
	"github.com/mgnsk/jsutil/array"
)

// MessagePort enables duplex communication with any js object
// implementing the onmessage event and postMessage method.
type MessagePort struct {
	// JS MessagePort object.
	value js.Value

	// A writer where onmessage event handler writes to.
	recvWriter *io.PipeWriter
	// A reader from where messages written to recvWriter can be read from.
	recvReader *io.PipeReader

	// remoteReady is closed when the remote end starts listening.
	remoteReady chan struct{}

	ack chan struct{}

	// isEOF when true, indicates that remote side closed its port.
	isEOF bool
	// isClosed indicates that the port was closed from this side.
	isClosed bool

	// Context that is canceled when port is closed.
	ctx    context.Context
	cancel context.CancelFunc
}

// Pipe returns a message channel pipe connection between ports.
func Pipe() (*MessagePort, *MessagePort) {
	ch := js.Global().Get("MessageChannel").New()
	return NewMessagePort(ch.Get("port1")), NewMessagePort(ch.Get("port2"))
}

// NewMessagePort constructor.
func NewMessagePort(value js.Value) *MessagePort {
	recvReader, recvWriter := io.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	port := &MessagePort{
		value:       value,
		recvReader:  recvReader,
		recvWriter:  recvWriter,
		remoteReady: make(chan struct{}),
		ack:         make(chan struct{}),
		ctx:         ctx,
		cancel:      cancel,
	}

	onerror, onmessage, onmessageerror := port.getEventHandlers()

	port.value.Set("onerror", onerror)
	port.value.Set("onmessage", onmessage)
	port.value.Set("onmessageerror", onmessageerror)

	port.notifyReady()

	// Clean up when port is not used anymore.
	runtime.SetFinalizer(port, func(p *MessagePort) {
		onerror.Release()
		onmessage.Release()
		onmessageerror.Release()
	})

	return port
}

// setEventHandlers sets the handlers for incoming messages and error handling.
func (port *MessagePort) getEventHandlers() (onerror, onmessage, onmessageerror js.Func) {
	onerror = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		jsutil.ConsoleLog("MessagePort: onerror:", args[0])
		return nil
	})

	onmessage = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// TODO assert that args are valid.
		data := args[0].Get("data")

		if data.Get("ready") != js.Undefined() {
			go func() {
				port.remoteReady <- struct{}{}
			}()
			return nil
		}

		if data.Get("ack") != js.Undefined() {
			go func() {
				port.ack <- struct{}{}
			}()
			return nil
		}

		// Handle port close from other side and start emitting EOF.
		EOF := data.Get("EOF")
		if EOF != js.Undefined() {
			// Set the EOF flag for Write. It does not use the pipe.
			port.isEOF = true
			port.cancel()
			// Close only writer. reader will get an EOF.
			port.recvWriter.Close()
			port.value.Call("close")
			return nil
		}

		// Remote call.
		rc := data.Get("rc")
		if jsutil.IsWorker && rc != js.Undefined() {

			call := newCallFromJS(
				data.Get("rc"),
				data.Get("input"),
				data.Get("output"),
			)

			// Currently allow 1 concurrent call per worker.
			// TODO configure this on runtime.
			// It can happen if multiple ports are scheduling into this one.
			if atomic.AddUint64(&CallCount, 1) > 1 {
				jsutil.ConsoleLog("Rescheduling...")
				// Reschedule until we have a free worker.
				go GlobalScheduler.Call(context.TODO(), call)
				return nil
			}

			// Notify the caller to send input now as we have set up event handlers.
			if call.Input != nil {
				ack(call.Input.JSValue())
			}

			go call.exec(func() {
				//	atomic.AddUint64(&CallCount, ^uint64(0))
				// Ack call output when call finished.
				ack(call.Output.JSValue())
			})
			return nil
		}

		// ArrayBuffer data message.
		arr := data.Get("arr")
		if arr != js.Undefined() {
			go func() {
				// Ack enables blocking write calls on the other side.
				defer ack(port.value)

				recvBytes, err := array.Buffer(arr).CopyBytes()
				if err != nil {
					errorx.Panic(errorx.Decorate(err, "copyBytes: error"))
				} else if len(recvBytes) == 0 {
					errorx.Panic(errorx.InternalError.New("copyBytes: 0 bytes"))
				}

				if n, err := port.recvWriter.Write(recvBytes); err == io.ErrClosedPipe {
					// This side of the port was closed. Notify other side.
					port.notifyEOF()
				} else if err == io.EOF {
					// Other side of port was closed. Close call was already handled.
				} else if err != nil {
					errorx.Panic(errorx.Decorate(err, "recvWriter: write error"))
				} else if n == 0 {
					errorx.Panic(errorx.InternalError.New("recvWriter: 0 bytes"))
				}
			}()
			return nil
		}

		return nil
	})

	onmessageerror = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		jsutil.ConsoleLog("MessagePort: onmessageerror:", args[0])
		return nil
	})

	return onerror, onmessage, onmessageerror
}

// Read from port.
func (port *MessagePort) Read(p []byte) (n int, err error) {
	return port.recvReader.Read(p)
}

// Write to port.
func (port *MessagePort) Write(p []byte) (n int, err error) {
	// Since we don't use a pipe on the write side,
	// we have to rely on manual signaling.
	if port.isEOF {
		return 0, io.EOF
	} else if port.isClosed {
		return 0, io.ErrClosedPipe
	}

	arr, err := array.CreateBufferFromSlice(p)
	if err != nil {
		return 0, err
	}

	messages := map[string]interface{}{"arr": arr}
	transferables := []interface{}{arr}
	port.PostMessage(messages, transferables)
	<-port.ack
	return len(p), nil
}

// Close the port.
func (port *MessagePort) Close() error {
	if port.isEOF {
		return io.EOF
	} else if port.isClosed {
		return io.ErrClosedPipe
	}

	// Let port.Write know we are closed.
	port.isClosed = true
	// Stop schedulers to this port.
	port.cancel()
	// Notify remote end of EOF.
	port.notifyEOF()
	port.recvReader.Close()
	port.recvWriter.Close()
	port.value.Call("close")
	return nil
}

// JSValue returns the underlying js value.
func (port *MessagePort) JSValue() js.Value {
	if port == nil {
		return js.Null()
	}
	return port.value
}

func (port *MessagePort) notifyEOF() {
	// Notify the remote side to emit an EOF from now on.
	port.PostMessage(map[string]interface{}{
		"EOF": true,
	})
}

func (port *MessagePort) notifyReady() {
	port.PostMessage(map[string]interface{}{
		"ready": true,
	})
}

// PostMessage sends a raw js message to remote end.
func (port *MessagePort) PostMessage(args ...interface{}) {
	port.value.Call("postMessage", args...)
}

// RemoteReady returns a channel that is closed when the remote end starts listening.
func (port *MessagePort) RemoteReady() <-chan struct{} {
	return port.remoteReady
}
