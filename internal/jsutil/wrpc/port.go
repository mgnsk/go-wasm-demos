package wrpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"syscall/js"
	"time"

	"github.com/mgnsk/go-wasm-demos/internal/jsutil/array"
)

// RawReader is a port reader interface.
type RawReader interface {
	ReadRaw(context.Context) (js.Value, error)
}

// RawWriter is a port writer interface.
type RawWriter interface {
	WriteRaw(context.Context, map[string]interface{}, []interface{}) error
}

// RawWriteCloser is a port writer and closer interface.
type RawWriteCloser interface {
	RawWriter
	io.Closer
}

// Conn is an interface to MessagePort.
type Conn interface {
	RawReader
	RawWriteCloser
	net.Conn
}

var _ Conn = &MessagePort{}

// MessagePort is a synchronous JS MessagePort wrapper.
type MessagePort struct {
	value    js.Value
	messages chan js.Value
	errs     chan error
	ack      chan struct{}
	done     chan struct{}

	readCtx     context.Context
	readCancel  context.CancelFunc
	writeCtx    context.Context
	writeCancel context.CancelFunc
}

// Pipe returns a synchronous duplex Conn pipe.
func Pipe() (*MessagePort, *MessagePort) {
	ch := js.Global().Get("MessageChannel").New()
	p1 := NewMessagePort(ch.Get("port1"))
	p2 := NewMessagePort(ch.Get("port2"))
	return p1, p2
}

// NewMessagePort wraps a JS value into MessagePort.
func NewMessagePort(value js.Value) *MessagePort {
	p := &MessagePort{
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

// JSValue returns the underlying MessagePort value.
func (p *MessagePort) JSValue() js.Value {
	return p.value
}

// ReadMessage reads a single message or error from the port.
func (p *MessagePort) ReadRaw(ctx context.Context) (js.Value, error) {
	select {
	case <-ctx.Done():
		return js.Value{}, ctx.Err()
	case err := <-p.errs:
		return js.Value{}, err
	case msg := <-p.messages:
		return msg, nil
	}
}

// PostMessage is a blocking postMessage call.
func (p *MessagePort) WriteRaw(ctx context.Context, messages map[string]interface{}, transferables []interface{}) error {
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

// Read a byte array message from the port.
func (p *MessagePort) Read(b []byte) (n int, err error) {
	ctx := context.Background()
	if p.readCtx != nil {
		ctx = p.readCtx
	}

	data, err := p.ReadRaw(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return 0, io.ErrClosedPipe
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return 0, os.ErrDeadlineExceeded
		}
		return 0, err
	}

	arr := data.Get("arr")
	if arr.IsUndefined() {
		return 0, fmt.Errorf("invalid message")
	}

	return array.Buffer(arr).Read(b)
}

// Write a byte array message into the port.
func (p *MessagePort) Write(b []byte) (n int, err error) {
	arr, err := array.CreateBufferFromSlice(b)
	if err != nil {
		return 0, err
	}

	messages := map[string]interface{}{"arr": arr}
	transferables := []interface{}{arr}

	ctx := context.Background()
	if p.writeCtx != nil {
		ctx = p.writeCtx
	}

	if err := p.WriteRaw(ctx, messages, transferables); err != nil {
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

// Close the port.
func (p *MessagePort) Close() error {
	close(p.done)
	if p.readCancel != nil {
		p.readCancel()
	}
	if p.writeCancel != nil {
		p.writeCancel()
	}
	p.value.Call("close")
	return nil
	// TODO notify close on other side?
}

// LocalAddr returns the local network address.
func (p *MessagePort) LocalAddr() net.Addr {
	return nil
}

// RemoteAddr returns the remote network address.
func (p *MessagePort) RemoteAddr() net.Addr {
	return nil
}

// SetDeadline sets the deadline for all future operations.
func (p *MessagePort) SetDeadline(t time.Time) error {
	if t.IsZero() {
		p.readCtx, p.readCancel = nil, nil
		p.writeCtx, p.writeCancel = nil, nil
	} else {
		p.readCtx, p.readCancel = context.WithDeadline(context.Background(), t)
		p.writeCtx, p.writeCancel = context.WithDeadline(context.Background(), t)
	}
	return nil
}

// SetDeadline sets the read deadline for all future operations.
func (p *MessagePort) SetReadDeadline(t time.Time) error {
	if t.IsZero() {
		p.readCtx, p.readCancel = nil, nil
	} else {
		p.readCtx, p.readCancel = context.WithDeadline(context.Background(), t)
	}
	return nil
}

// SetDeadline sets the write deadline for all future operations.
func (p *MessagePort) SetWriteDeadline(t time.Time) error {
	if t.IsZero() {
		p.writeCtx, p.writeCancel = nil, nil
	} else {
		p.writeCtx, p.writeCancel = context.WithDeadline(context.Background(), t)
	}
	return nil
}

func (p *MessagePort) onError(this js.Value, args []js.Value) interface{} {
	go func() {
		select {
		case <-p.done:
		case p.errs <- js.Error{args[0]}:
		}
	}()
	return nil
}

func (p *MessagePort) onMessage(this js.Value, args []js.Value) interface{} {
	go func() {
		data := args[0].Get("data")

		if !data.Get("__ack").IsUndefined() {
			p.ack <- struct{}{}
			return
		}

		defer p.value.Call("postMessage", map[string]interface{}{"__ack": true})

		select {
		case <-p.done:
		case p.messages <- data:
		}
	}()

	return nil
}
