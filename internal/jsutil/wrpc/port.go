package wrpc

import (
	"context"
	"io"
	"runtime"
	"sync"
	"syscall/js"
)

// Reader is a port reader interface.
type Reader interface {
	Read(context.Context) (js.Value, error)
}

// Writer is a port writer interface.
type Writer interface {
	Write(context.Context, map[string]interface{}, []interface{}) error
}

// ReadWriter is a port read-writer interface.
type ReadWriter interface {
	Reader
	Writer
}

// ReadWriteCloser is a port reader, writer and closer interface.
type ReadWriteCloser interface {
	ReadWriter
	io.Closer
}

// MessagePort is a synchronous JS MessagePort wrapper.
type MessagePort struct {
	value    js.Value
	messages chan js.Value
	ack      chan struct{}
	done     chan struct{}
	once     sync.Once
	err      error
}

// Pipe returns a synchronous duplex MessagePort pipe.
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
		ack:      make(chan struct{}),
		done:     make(chan struct{}),
	}

	onError := js.FuncOf(p.onError)
	onMessageError := js.FuncOf(p.onError)
	onMessage := js.FuncOf(p.onMessage)

	value.Set("onerror", onError)
	value.Set("onmessageerror", onMessageError)
	value.Set("onmessage", onMessage)

	runtime.SetFinalizer(p, func(port interface{}) {
		port.(*MessagePort).value.Call("close")
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

// Read reads a single message or error from the port.
func (p *MessagePort) Read(ctx context.Context) (js.Value, error) {
	select {
	case <-ctx.Done():
		return js.Value{}, ctx.Err()
	case <-p.done:
		return js.Value{}, p.err
	case msg := <-p.messages:
		p.value.Call("postMessage", map[string]interface{}{"__ack": true})
		return msg, nil
	}
}

// Write is a blocking postMessage call.
func (p *MessagePort) Write(ctx context.Context, messages map[string]interface{}, transferables []interface{}) error {
	p.value.Call("postMessage", messages, transferables)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.done:
		return p.err
	case <-p.ack:
		return nil
	}
}

// Close the port. All pending reads and writes are unblocked and return io.ErrClosedPipe.
func (p *MessagePort) Close() {
	p.once.Do(func() {
		p.err = io.ErrClosedPipe
		close(p.done)
		p.value.Call("postMessage", map[string]interface{}{"__eof": true})
	})
}

func (p *MessagePort) onError(_ js.Value, args []js.Value) interface{} {
	go func() {
		p.once.Do(func() {
			p.err = js.Error{Value: args[0]}
			close(p.done)
		})
	}()
	return nil
}

func (p *MessagePort) onMessage(_ js.Value, args []js.Value) interface{} {
	go func() {
		data := args[0].Get("data")
		switch {
		case !data.Get("__eof").IsUndefined():
			p.once.Do(func() {
				p.err = io.EOF
				close(p.done)
			})
		case !data.Get("__ack").IsUndefined():
			select {
			case <-p.done:
			case p.ack <- struct{}{}:
			}
		default:
			select {
			case <-p.done:
			case p.messages <- data:
			}
		}
	}()
	return nil
}
