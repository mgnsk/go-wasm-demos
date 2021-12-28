//go:build js && wasm
// +build js,wasm

package wrpc

import (
	"context"
	"io"
	"net"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall/js"
	"time"

	"github.com/joomcode/errorx"
	"github.com/mgnsk/go-wasm-demos/internal/jsutil"
	"github.com/mgnsk/go-wasm-demos/internal/jsutil/array"
)

// callCount specifies how many calls are currently processing.
var callCount uint64 = 0

// MessagePort enables duplex communication with any js object
// implementing the onmessage event and postMessage method.
type MessagePort struct {
	// JS MessagePort object.
	port js.Value

	// A writer where onmessage event handler writes to.
	recvWriter net.Conn
	// A reader from where messages written to recvWriter can be read from.
	recvReader net.Conn

	// remoteReady is closed when the remote end starts listening.
	remoteReady chan struct{}

	// readyOnce sets port to be ready.
	readyOnce sync.Once

	ack chan struct{}

	err error

	onError        js.Func
	onMessageError js.Func
	onMessage      js.Func

	writeCtx    context.Context
	writeCancel context.CancelFunc
}

// Pipe returns a message channel pipe connection between ports.
func Pipe() (*MessagePort, *MessagePort) {
	ch := js.Global().Get("MessageChannel").New()
	port1 := NewMessagePort(ch.Get("port1"))
	port2 := NewMessagePort(ch.Get("port2"))

	return port1, port2
}

// NewMessagePort constructor.
func NewMessagePort(port js.Value) *MessagePort {
	recvReader, recvWriter := net.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	p := &MessagePort{
		port:        port,
		recvReader:  recvReader,
		recvWriter:  recvWriter,
		remoteReady: make(chan struct{}),
		ack:         make(chan struct{}),
		writeCtx:    ctx,
		writeCancel: cancel,
	}

	p.onError = js.FuncOf(p.onErrorHandler)
	p.onMessageError = js.FuncOf(p.onMessageErrorHandler)
	p.onMessage = js.FuncOf(p.onMessageHandler)

	p.port.Set("onerror", p.onError)
	p.port.Set("onmessageerror", p.onMessageError)
	p.port.Set("onmessage", p.onMessage)

	runtime.SetFinalizer(p, func(v interface{}) {
		port := v.(*MessagePort)
		port.onError.Release()
		port.onMessageError.Release()
		port.onMessage.Release()
	})

	return p
}

// Read from port.
func (p *MessagePort) Read(b []byte) (n int, err error) {
	return p.recvReader.Read(b)
}

// Write to port.
func (p *MessagePort) Write(b []byte) (n int, err error) {
	// Since we don't use a pipe on the write side,
	// we have to rely on manual signaling.
	if p.err != nil {
		return 0, p.err
	}

	arr, err := array.CreateBufferFromSlice(b)
	if err != nil {
		return 0, err
	}

	messages := map[string]interface{}{"arr": arr}
	transferables := []interface{}{arr}

	p.port.Call("postMessage", messages, transferables)

	select {
	case <-p.writeCtx.Done():
		if p.err != nil {
			return 0, p.err
		}
		if p.writeCtx.Err() == context.DeadlineExceeded {
			return 0, os.ErrDeadlineExceeded
		}
	case <-p.ack:
	}

	return len(b), nil
}

// Close the port.
func (p *MessagePort) Close() error {
	if p.err != nil {
		return p.err
	}

	p.recvReader.Close()
	p.recvWriter.Close()

	p.err = io.ErrClosedPipe

	p.writeCancel()

	// Notify remote end of EOF.
	p.notifyEOF()

	p.port.Call("close")

	return nil
}

// LocalAddr returns the local port addr
func (p *MessagePort) LocalAddr() net.Addr {
	return nil
}

// RemoteAddr returns the remote port addr.
func (p *MessagePort) RemoteAddr() net.Addr {
	return nil
}

// SetDeadline for read and write operations.
func (p *MessagePort) SetDeadline(t time.Time) error {
	if p.err != nil {
		return p.err
	}

	p.recvReader.SetReadDeadline(t)
	p.setWriteDeadline(t)

	return nil
}

// SetReadDeadline for read operations.
func (p *MessagePort) SetReadDeadline(t time.Time) error {
	if p.err != nil {
		return p.err
	}

	p.recvReader.SetReadDeadline(t)

	return nil
}

// SetWriteDeadline for write operations.
func (p *MessagePort) SetWriteDeadline(t time.Time) error {
	if p.err != nil {
		return p.err
	}

	p.setWriteDeadline(t)

	return nil
}

// JSValue returns the underlying js value.
func (p *MessagePort) JSValue() js.Value {
	if p == nil {
		return js.Null()
	}
	return p.port
}

func (p *MessagePort) setWriteDeadline(t time.Time) {
	if t.IsZero() {
		p.writeCtx, p.writeCancel = nil, nil
	} else {
		p.writeCtx, p.writeCancel = context.WithDeadline(context.Background(), t)
	}
}

func (p *MessagePort) notifyEOF() {
	// Notify the remote side to emit an EOF from now on.
	p.port.Call("postMessage", map[string]interface{}{
		"EOF": true,
	})
}

func (p *MessagePort) onErrorHandler(this js.Value, args []js.Value) interface{} {
	jsutil.ConsoleLog("MessagePort: onerror:", args[0])
	return nil
}

func (p *MessagePort) onMessageErrorHandler(this js.Value, args []js.Value) interface{} {
	jsutil.ConsoleLog("MessagePort: onmessageerror:", args[0])
	return nil
}

func (p *MessagePort) onMessageHandler(this js.Value, args []js.Value) interface{} {
	// TODO assert that args are valid.
	data := args[0].Get("data")

	if !data.Get("ack").IsUndefined() {
		go func() {
			p.ack <- struct{}{}
		}()
		return nil
	}

	// Handle port close from other side and start emitting EOF.
	if !data.Get("EOF").IsUndefined() {
		if p.writeCancel != nil {
			p.writeCancel()
		}

		// Close only writer. reader will get an EOF.
		p.recvWriter.Close()

		p.port.Call("close")

		// TODO
		// close(p.ack)

		return nil
	}

	// Remote call.
	rc := data.Get("rc")
	if jsutil.IsWorker && !rc.IsUndefined() {
		call := NewCallFromJS(
			data.Get("rc"),
			data.Get("input"),
			data.Get("output"),
		)

		// Currently allow 1 concurrent call per worker.
		// TODO configure this on runtime.
		// It can happen if multiple ports are scheduling into this one.
		if atomic.AddUint64(&callCount, 1) > 1 {
			jsutil.ConsoleLog("Rescheduling...")
			// Reschedule until we have a free worker.
			go GlobalScheduler.Call(context.TODO(), call)
			return nil
		}

		go call.Execute()

		return nil
	}

	// ArrayBuffer data message.
	arr := data.Get("arr")
	if !arr.IsUndefined() {
		go func() {
			// Ack enables blocking write calls on the other side.
			defer ack(p.port)

			recvBytes, err := array.Buffer(arr).CopyBytes()
			if err != nil {
				errorx.Panic(errorx.Decorate(err, "copyBytes: error"))
			} else if len(recvBytes) == 0 {
				errorx.Panic(errorx.InternalError.New("copyBytes: 0 bytes"))
			}

			if n, err := p.recvWriter.Write(recvBytes); err == io.ErrClosedPipe {
				// This side of the port was closed. Notify other side.
				p.notifyEOF()
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
}
