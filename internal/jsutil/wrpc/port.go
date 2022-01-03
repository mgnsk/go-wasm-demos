package wrpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"syscall/js"

	"github.com/mgnsk/go-wasm-demos/internal/jsutil/array"
)

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
func (p *MessagePort) ReadMessage() (js.Value, error) {
	select {
	case <-p.done:
		return js.Value{}, p.err
	case msg := <-p.messages:
		p.value.Call("postMessage", map[string]interface{}{"__ack": true})
		return msg, nil
	}
}

// Write is a blocking postMessage call.
func (p *MessagePort) WriteMessage(messages map[string]interface{}, transferables []interface{}) error {
	p.value.Call("postMessage", messages, transferables)
	select {
	case <-p.done:
		return p.err
	case <-p.ack:
		return nil
	}
}

// Read a byte array message from the conn.
func (p *MessagePort) Read(b []byte) (n int, err error) {
	msg, err := p.ReadMessage()
	if err != nil {
		return 0, err
	}

	arr := msg.Get("arr")
	if arr.IsUndefined() {
		return 0, fmt.Errorf("expected an ArrayBuffer message")
	}

	return copy(b, array.ArrayBuffer(arr).Bytes()), nil
}

// Write a byte array message into the conn.
func (p *MessagePort) Write(b []byte) (n int, err error) {
	arr := array.NewArrayBufferFromSlice(b).JSValue()
	messages := map[string]interface{}{"arr": arr}
	transferables := []interface{}{arr}

	if err := p.WriteMessage(messages, transferables); err != nil {
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

// Close the port. All pending reads and writes are unblocked and return io.ErrClosedPipe.
func (p *MessagePort) Close() error {
	p.once.Do(func() {
		p.err = io.ErrClosedPipe
		close(p.done)
		p.value.Call("postMessage", map[string]interface{}{"__eof": true})
	})
	if p.err == io.ErrClosedPipe {
		return p.err
	}
	return nil
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
