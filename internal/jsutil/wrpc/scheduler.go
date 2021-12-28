//go:build js && wasm
// +build js,wasm

package wrpc

import (
	"context"
	"sync"
	"syscall/js"
	"unsafe"
)

// GlobalScheduler is main scheduler to schedule to workers.
var GlobalScheduler = NewScheduler()

// Call is a remote call that can be scheduled to a worker.
type Call struct {
	// rc will be run in a remote webworker.
	rc RemoteCall
	// input is a port where the worker can read its input data from.
	input *MessagePort
	// output is the port where the result gets written into.
	output *MessagePort
}

// NewCallFromJS constructs a call from javascript arguments.
func NewCallFromJS(rc, input, output js.Value) Call {
	rcPtr := uintptr(rc.Int())
	remoteCall := *(*RemoteCall)(unsafe.Pointer(&rcPtr))

	var inputPort *MessagePort
	if input.Truthy() {
		inputPort = NewMessagePort(input)
	}

	return Call{
		rc:     remoteCall,
		input:  inputPort,
		output: NewMessagePort(output),
	}
}

// GetJS returns js messages along with transferables that can be sent over a MessagePort.
func (c Call) GetJS() (messages map[string]interface{}, transferables []interface{}) {
	rc := *(*uintptr)(unsafe.Pointer(&c.rc))
	messages = map[string]interface{}{
		"rc":     int(rc),
		"output": c.output,
	}
	transferables = []interface{}{
		c.output,
	}
	if c.input != nil {
		messages["input"] = c.input
		transferables = append(transferables, c.input)
	}
	return
}

// Execute the call locally.
func (c Call) Execute() {
	// if c.input != nil {
	// 	ack(c.input.JSValue())
	// }
	defer ack(c.output.JSValue())
	c.rc(c.input, c.output)
}

// Scheduler schedules calls to ports.
type Scheduler struct {
	queue chan *message
}

type message struct {
	wg   sync.WaitGroup
	call Call
}

// NewScheduler constructor.
func NewScheduler() *Scheduler {
	return &Scheduler{
		queue: make(chan *message),
	}
}

// Run starts a scheduler to schedule calls to port.
// Runs sync on a single port.
func (s *Scheduler) Run(ctx context.Context, port *MessagePort) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-s.queue:
			messages, transferables := msg.call.GetJS()
			port.JSValue().Call("postMessage", messages, transferables)
			// TODO
			// <-port.ack
			msg.wg.Wait()
		}
	}
}

// Call sends the remote call on the first worker who receives it.
// Workers are single-threaded.
func (s *Scheduler) Call(ctx context.Context, call Call) error {
	msg := &message{
		call: call,
	}

	msg.wg.Add(1)
	defer msg.wg.Done()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.queue <- msg:
		select {
		case <-call.output.ack:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}
