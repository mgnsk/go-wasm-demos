package wrpcnet

import (
	"container/list"
	"errors"
	"io"
	"runtime"
	"syscall/js"

	"github.com/mgnsk/go-wasm-demos/pkg/array"
	"github.com/mgnsk/go-wasm-demos/pkg/jsutil"
)

// MessagePort is a synchronous JS MessagePort wrapper.
type MessagePort struct {
	Value    js.Value
	messages *list.List
	notify   chan struct{}
	ack      chan struct{}
	done     chan struct{}
	err      error
}

// Pipe returns a synchronous duplex MessagePort pipe.
func Pipe() (*MessagePort, *MessagePort) {
	ch := js.Global().Get("MessageChannel").New()
	p1 := NewMessagePort(ch.Get("port1"))
	p2 := NewMessagePort(ch.Get("port2"))
	return p1, p2
}

// NewMessagePort creates a synchronous JS MessagePort wrapper.
func NewMessagePort(value js.Value) *MessagePort {
	p := &MessagePort{
		Value:    value,
		messages: list.New(),
		notify:   make(chan struct{}),
		ack:      make(chan struct{}),
		done:     make(chan struct{}),
	}

	onError := js.FuncOf(p.onError)
	onMessageError := js.FuncOf(p.onError)
	onMessage := js.FuncOf(p.onMessage)

	value.Set("onerror", onError)
	value.Set("onmessageerror", onMessageError)
	value.Set("onmessage", onMessage)

	runtime.SetFinalizer(p, func(any) {
		onError.Release()
		onMessageError.Release()
		onMessage.Release()
	})

	return p
}

// ReadMessage reads a single message or error from the port.
func (p *MessagePort) ReadMessage() (js.Value, error) {
	select {
	case <-p.done:
		return js.Value{}, p.err
	case <-p.notify:
		p.Value.Call("postMessage", map[string]any{"__ack": true})

		msg := p.messages.Remove(p.messages.Front()).(js.Value)
		jsutil.ConsoleLog("readMessage", msg)

		return msg, nil
	}
}

// WriteMessage writes a messages into the port.
// It blocks until the remote side reads the message.
func (p *MessagePort) WriteMessage(messages map[string]any, transferables []any) error {
	p.Value.Call("postMessage", messages, transferables)
	select {
	case <-p.done:
		// jsutil.ConsoleLog("postMessage error", p.err.Error())
		return p.err
	case <-p.ack:
		jsutil.ConsoleLog("postMessage ack")
		return nil
	}
}

// Read a byte array message from the port.
func (p *MessagePort) Read(b []byte) (n int, err error) {
	msg, err := p.ReadMessage()
	if err != nil {
		return 0, err
	}

	ab := msg.Get("arr")
	if ab.IsUndefined() {
		panic("expected an ArrayBuffer message")
	}

	arr := array.NewUint8Array(ab)
	if arr.Len() > len(b) {
		p.messages.PushFront(msg)

		return 0, io.ErrShortBuffer
	}

	return arr.CopyBytesToGo(b), nil
}

// Write a byte array message into the port.
func (p *MessagePort) Write(b []byte) (n int, err error) {
	ab := array.NewFromSlice(b).ArrayBuffer()
	messages := map[string]any{"arr": ab}
	transferables := []any{ab}

	if err := p.WriteMessage(messages, transferables); err != nil {
		return 0, err
	}

	return len(b), nil
}

// Close the port. All pending reads and writes are unblocked and return io.ErrClosedPipe.
func (p *MessagePort) Close() error {
	p.err = io.ErrClosedPipe
	close(p.done)
	p.Value.Call("postMessage", map[string]any{"__eof": true})
	p.Value.Call("close")
	return nil
}

// CloseWithError writes an error message into the port and closes the port.
// All pending reads and writes are unblocked and return io.ErrClosedPipe.
func (p *MessagePort) CloseWithError(err error) {
	p.err = io.ErrClosedPipe
	close(p.done)
	p.Value.Call("postMessage", map[string]any{"__err": err.Error()})
	p.Value.Call("close")
}

func (p *MessagePort) onError(_ js.Value, args []js.Value) any {
	if p.err == nil {
		p.err = js.Error{Value: args[0]}
		close(p.done)
	}
	return nil
}

func (p *MessagePort) onMessage(this js.Value, args []js.Value) any {
	data := args[0].Get("data")
	eof := data.Get("__eof")
	err := data.Get("__err")
	ack := data.Get("__ack")

	switch {
	case !eof.IsUndefined():
		if p.err == nil {
			p.err = io.EOF
			close(p.done)
		}

	case !err.IsUndefined():
		if p.err == nil {
			p.err = errors.New(err.String())
			close(p.done)
		}

	case !ack.IsUndefined():
		go func() {
			select {
			case <-p.done:
			case p.ack <- struct{}{}:
			}
		}()

	default:
		p.messages.PushBack(data)

		go func() {
			select {
			case <-p.done:
			case p.notify <- struct{}{}:
			}
		}()
	}

	return nil
}
