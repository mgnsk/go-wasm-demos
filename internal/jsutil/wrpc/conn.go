package wrpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"syscall/js"

	"github.com/mgnsk/go-wasm-demos/internal/jsutil/array"
)

// Conn wraps MessagePort with io.ReadWriteCloser interface.
type Conn struct {
	port *MessagePort
}

func connPipe() (*Conn, *Conn) {
	p1, p2 := Pipe()
	return NewConn(p1), NewConn(p2)
}

// NewConn creates a new conn.
func NewConn(port *MessagePort) *Conn {
	return &Conn{
		port: port,
	}
}

// JSValue returns the underlying JS MessagePort value.
// TODO: this is currently needed for Call.
func (c *Conn) JSValue() js.Value {
	return c.port.JSValue()
}

// Read a byte array message from the conn.
func (c *Conn) Read(b []byte) (n int, err error) {
	msg, err := c.port.Read(context.Background())
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return 0, io.ErrClosedPipe
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return 0, os.ErrDeadlineExceeded
		}
		return 0, err
	}

	arr := msg.Get("arr")
	if arr.IsUndefined() {
		return 0, fmt.Errorf("expected an ArrayBuffer message")
	}

	return copy(b, array.ArrayBuffer(arr).Bytes()), nil
}

// Write a byte array message into the conn.
func (c *Conn) Write(b []byte) (n int, err error) {
	arr := array.NewArrayBufferFromSlice(b).JSValue()
	messages := map[string]interface{}{"arr": arr}
	transferables := []interface{}{arr}

	if err := c.port.Write(context.Background(), messages, transferables); err != nil {
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

// Close the conn.
func (c *Conn) Close() error {
	return c.port.Close()
}
