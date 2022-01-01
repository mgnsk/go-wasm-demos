//go:build js && wasm
// +build js,wasm

package wrpc

import (
	"context"
	"fmt"
	"syscall/js"
	"unsafe"
)

// defaultScheduler is main scheduler to schedule to workers.
var defaultScheduler = NewScheduler()

// Call is a remote call that can be scheduled to a worker.
type Call struct {
	// writer is the port where the result gets written into.
	writer js.Value
	// reader is a port where the worker can read its reader data from.
	reader js.Value
	// rc is a static pointer to a remote call function.
	rc int
}

// NewCall creates a new wrpc call.
func NewCall(w *Conn, r *Conn, f RemoteCall) Call {
	c := Call{
		writer: w.JSValue(),
		rc:     int(*(*uintptr)(unsafe.Pointer(&f))),
	}
	if r != nil {
		c.reader = r.JSValue()
	}
	return c
}

// NewCallFromJS constructs a call from JS message.
func NewCallFromJS(data js.Value) Call {
	return Call{
		writer: data.Get("output"),
		reader: data.Get("input"),
		rc:     data.Get("rc").Int(),
	}
}

// Execute the call locally.
func (c Call) Execute() {
	rcPtr := uintptr(c.rc)
	f := *(*RemoteCall)(unsafe.Pointer(&rcPtr))

	w := NewConn(NewMessagePort(c.writer))

	var r *Conn
	if !c.reader.IsUndefined() {
		r = NewConn(NewMessagePort(c.reader))
		defer r.Close()
	}

	f(w, r)
}

// Scheduler schedules calls to ports.
type Scheduler struct {
	ports chan ReadWriter
}

// NewScheduler constructor.
func NewScheduler() *Scheduler {
	return &Scheduler{
		ports: make(chan ReadWriter, 65535),
	}
}

// Register a port to be scheduler calls to.
func (s *Scheduler) Register(port ReadWriter) {
	s.ports <- port
}

// Call sends the remote call to the first worker who receives it.
// It returns when worker receives the call.
func (s *Scheduler) Call(ctx context.Context, call Call) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("scheduler: could not find a port: %w", ctx.Err())
	case port := <-s.ports:
		defer func() {
			s.ports <- port
		}()
		messages := map[string]interface{}{
			"rc":     call.rc,
			"output": call.writer,
		}
		transferables := []interface{}{call.writer}
		if !call.reader.IsUndefined() {
			messages["input"] = call.reader
			transferables = append(transferables, call.reader)
		}
		if err := port.Write(ctx, messages, transferables); err != nil {
			return fmt.Errorf("error sending call: %w", err)
		}
		if _, err := port.Read(ctx); err != nil {
			return fmt.Errorf("error reading call ACK: %w", err)
		}
		return nil
	}
}
