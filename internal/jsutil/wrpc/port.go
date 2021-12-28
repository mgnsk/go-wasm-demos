package wrpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"syscall/js"

	"github.com/mgnsk/go-wasm-demos/internal/jsutil/array"
)

// RawReader is a port reader interface.
type RawReader interface {
	ReadRaw(context.Context) (js.Value, error)
}

// RawWriter is a port writer interface.
type RawWriter interface {
	WriteRaw(context.Context, map[string]interface{}, ...interface{}) error
}

// RawWriteCloser is a port writer and closer interface.
type RawWriteCloser interface {
	RawWriter
	io.Closer
}

// Port is an interface to MessagePort.
type Port interface {
	RawReader
	RawWriteCloser
	io.Reader
	io.WriteCloser
	js.Wrapper
}

var _ Port = &messagePort{}

// messagePort enables blocking duplex communication with any js object
// implementing the onmessage event and postMessage method.
type messagePort struct {
	value    js.Value
	messages chan js.Value
	errs     chan error
	ack      chan struct{}
	done     chan struct{}
}

// Pipe returns a duplex Port pipe.
func Pipe() (Port, Port) {
	ch := js.Global().Get("MessageChannel").New()
	p1 := NewMessagePort(ch.Get("port1"))
	p2 := NewMessagePort(ch.Get("port2"))
	return p1, p2
}

// NewMessagePort wraps a JS value into MessagePort.
func NewMessagePort(value js.Value) Port {
	p := &messagePort{
		value:    value,
		messages: make(chan js.Value),
		errs:     make(chan error),
		ack:      make(chan struct{}),
		done:     make(chan struct{}),
	}

	onError := js.FuncOf(p.onError)
	onMessageError := js.FuncOf(p.onError)
	onMessage := js.FuncOf(p.onMessage)

	value.Set("onerror", onError)
	value.Set("onmessageerror", onMessageError)
	value.Set("onmessage", onMessage)

	runtime.SetFinalizer(p, func(interface{}) {
		onError.Release()
		onMessageError.Release()
		onMessage.Release()
	})

	return p
}

// JSValue returns the JS MessagePort value.
func (p *messagePort) JSValue() js.Value {
	return p.value
}

// Close the port.
func (p *messagePort) Close() error {
	close(p.done)
	p.value.Call("close")
	return nil
	// TODO notify close on other side?
}

// ReadMessage reads a single message or error from the port.
func (p *messagePort) ReadRaw(ctx context.Context) (js.Value, error) {
	select {
	case <-ctx.Done():
		return js.Value{}, ctx.Err()
	case err := <-p.errs:
		return js.Value{}, err
	case msg := <-p.messages:
		return msg, nil
	}
}

// Read a byte array message from the port.
func (p *messagePort) Read(b []byte) (n int, err error) {
	data, err := p.ReadRaw(context.TODO())
	if err != nil {
		return 0, err
	}

	arr := data.Get("arr")
	if arr.IsUndefined() {
		return 0, fmt.Errorf("invalid message")
	}

	return array.Buffer(arr).Read(b)
}

// Write a byte array message into the port.
func (p *messagePort) Write(b []byte) (n int, err error) {
	// Since we don't use a pipe on the write side,
	// we have to rely on manual signaling.
	select {
	case <-p.done:
		return 0, io.ErrClosedPipe
	default:
	}

	arr, err := array.CreateBufferFromSlice(b)
	if err != nil {
		return 0, err
	}

	messages := map[string]interface{}{"arr": arr}
	transferables := []interface{}{arr}

	if err := p.WriteRaw(context.TODO(), messages, transferables); err != nil {
		if errors.Is(err, context.Canceled) {
			return 0, io.ErrClosedPipe
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return 0, os.ErrDeadlineExceeded
		}
		return 0, err
	}

	return len(b), nil
}

// PostMessage is a blocking postMessage call.
func (p *messagePort) WriteRaw(ctx context.Context, messages map[string]interface{}, transferables ...interface{}) error {
	p.value.Call("postMessage", messages, transferables)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-p.errs:
		return err
	case <-p.ack:
		return nil
	}
}

func (p *messagePort) onError(this js.Value, args []js.Value) interface{} {
	go func() {
		select {
		case <-p.done:
		case p.errs <- fmt.Errorf("%v", args):
		}
	}()
	return nil
}

func (p *messagePort) onMessage(this js.Value, args []js.Value) interface{} {
	go func() {
		data := args[0].Get("data")

		if !data.Get("__ack").IsUndefined() {
			p.ack <- struct{}{}
			return
		}

		select {
		case <-p.done:
		case p.messages <- data:
			// Post the ACK once someone read the message.
			p.value.Call("postMessage", map[string]interface{}{"__ack": true})
		}
	}()

	return nil
}
