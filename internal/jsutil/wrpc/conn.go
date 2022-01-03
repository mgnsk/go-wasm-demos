package wrpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"syscall/js"
	"time"

	"github.com/mgnsk/go-wasm-demos/internal/jsutil/array"
)

// Conn implements net.Conn interface around MessagePort.
type Conn struct {
	port        *MessagePort
	readCtx     context.Context
	readCancel  context.CancelFunc
	writeCtx    context.Context
	writeCancel context.CancelFunc
}

var _ net.Conn = &Conn{}

// Pipe returns a synchronous duplex MessagePort conn pipe.
func ConnPipe() (*Conn, *Conn) {
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
	ctx := context.Background()
	if c.readCtx != nil {
		ctx = c.readCtx
	}

	msg, err := c.port.Read(ctx)
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

	ctx := context.Background()
	if c.writeCtx != nil {
		ctx = c.writeCtx
	}

	if err := c.port.Write(ctx, messages, transferables); err != nil {
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

// LocalAddr returns the local network address.
func (c *Conn) LocalAddr() net.Addr {
	return nil
}

// RemoteAddr returns the remote network address.
func (c *Conn) RemoteAddr() net.Addr {
	return nil
}

// Close the conn.
func (c *Conn) Close() error {
	if c.readCancel != nil {
		c.readCancel()
	}
	if c.writeCancel != nil {
		c.writeCancel()
	}
	c.port.Close()
	return nil
}

// SetDeadline sets the deadline for all future operations.
func (c *Conn) SetDeadline(t time.Time) error {
	if t.IsZero() {
		c.readCtx, c.readCancel = nil, nil
		c.writeCtx, c.writeCancel = nil, nil
	} else {
		c.readCtx, c.readCancel = context.WithDeadline(context.Background(), t)
		c.writeCtx, c.writeCancel = context.WithDeadline(context.Background(), t)
	}
	return nil
}

// SetDeadline sets the read deadline for all future operations.
func (c *Conn) SetReadDeadline(t time.Time) error {
	if t.IsZero() {
		c.readCtx, c.readCancel = nil, nil
	} else {
		c.readCtx, c.readCancel = context.WithDeadline(context.Background(), t)
	}
	return nil
}

// SetDeadline sets the write deadline for all future operations.
func (c *Conn) SetWriteDeadline(t time.Time) error {
	if t.IsZero() {
		c.writeCtx, c.writeCancel = nil, nil
	} else {
		c.writeCtx, c.writeCancel = context.WithDeadline(context.Background(), t)
	}
	return nil
}
