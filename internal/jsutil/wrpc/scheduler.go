//go:build js && wasm
// +build js,wasm

package wrpc

import (
	"context"
	"syscall/js"
	"unsafe"

	"github.com/mgnsk/go-wasm-demos/internal/jsutil"
)

// defaultScheduler is main scheduler to schedule to workers.
var defaultScheduler = NewScheduler()

// Call is a remote call that can be scheduled to a worker.
type Call struct {
	// rc will be run in a remote webworker.
	rc RemoteCall
	// reader is a port where the worker can read its reader data from.
	reader Conn
	// writer is the port where the result gets written into.
	writer Conn
}

// NewCallFromJS constructs a call from javascript arguments.
func NewCallFromJS(rc, input, output js.Value) Call {
	rcPtr := uintptr(rc.Int())
	remoteCall := *(*RemoteCall)(unsafe.Pointer(&rcPtr))

	var inputPort Conn
	if !input.IsUndefined() {
		inputPort = NewMessagePort(input)
	}

	return Call{
		rc:     remoteCall,
		reader: inputPort,
		writer: NewMessagePort(output),
	}
}

// GetJS returns js messages along with transferables that can be sent over a MessagePort.
func (c Call) GetJS() (messages map[string]interface{}, transferables []interface{}) {
	rc := *(*uintptr)(unsafe.Pointer(&c.rc))
	messages = map[string]interface{}{
		"rc":     int(rc),
		"output": c.writer,
	}
	transferables = []interface{}{
		c.writer,
	}
	if c.reader != nil {
		messages["input"] = c.reader
		transferables = append(transferables, c.reader)
	}
	return
}

// Execute the call locally.
func (c Call) Execute() {
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
func (s *Scheduler) Run(ctx context.Context, w RawWriter) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case call := <-s.queue:
			messages, transferables := call.GetJS()
			jsutil.ConsoleLog(messages, transferables)
			if err := w.WriteRaw(ctx, messages, transferables); err != nil {
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
