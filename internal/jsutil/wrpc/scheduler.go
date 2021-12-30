//go:build js && wasm
// +build js,wasm

package wrpc

import (
	"context"
	"syscall/js"
	"unsafe"
)

// defaultScheduler is main scheduler to schedule to workers.
var defaultScheduler = NewScheduler()

// Call is a remote call that can be scheduled to a worker.
type Call struct {
	// rc will be run in a remote webworker.
	rc RemoteCall
	// reader is a port where the worker can read its reader data from.
	reader *Conn
	// writer is the port where the result gets written into.
	writer *Conn
}

// NewCallFromJS constructs a call from javascript arguments.
func NewCallFromJS(rc, input, output js.Value) Call {
	rcPtr := uintptr(rc.Int())
	remoteCall := *(*RemoteCall)(unsafe.Pointer(&rcPtr))

	var inputConn *Conn
	if !input.IsUndefined() {
		inputConn = NewConn(NewMessagePort(input))
	}

	return Call{
		rc:     remoteCall,
		reader: inputConn,
		writer: NewConn(NewMessagePort(output)),
	}
}

// GetJS returns js messages along with transferables that can be sent over a MessagePort.
func (c Call) GetJS() (messages map[string]interface{}, transferables []interface{}) {
	rc := int(*(*uintptr)(unsafe.Pointer(&c.rc)))
	w := c.writer.JSValue()

	messages = map[string]interface{}{
		"rc":     rc,
		"output": w,
	}
	transferables = []interface{}{w}

	if c.reader != nil {
		r := c.reader.JSValue()
		messages["input"] = r
		transferables = append(transferables, r)
	}

	return
}

// Execute the call locally.
func (c Call) Execute() {
	if c.reader != nil {
		defer c.reader.Close()
	}
	c.rc(c.writer, c.reader)
}

// Scheduler schedules calls to ports.
type Scheduler struct {
	queue chan Call
}

// NewScheduler constructor.
func NewScheduler() *Scheduler {
	return &Scheduler{
		queue: make(chan Call),
	}
}

// Run starts a scheduler to schedule calls to port.
func (s *Scheduler) Run(ctx context.Context, w Writer) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case call := <-s.queue:
			messages, transferables := call.GetJS()
			if err := w.Write(ctx, messages, transferables); err != nil {
				return err
			}
		}
	}
}

// Call sends the remote call on the first worker who receives it.
func (s *Scheduler) Call(ctx context.Context, call Call) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.queue <- call:
	}
	return nil
}
