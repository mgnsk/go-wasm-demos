// +build js,wasm

package wrpc

import (
	"syscall/js"
	"unsafe"
)

// Call is a remote call that can be scheduled to a worker.
type Call struct {
	// RemoteCall will be run in a remote webworker.
	RemoteCall RemoteCall
	// InputReader is a port where the worker can read its input data from.
	Input *MessagePort
	// ResultPort is the port where the result gets written into.
	Output *MessagePort
}

// Execute the call locally.
func (c Call) exec(cb func()) {
	defer cb()
	c.RemoteCall(c.Input, c.Output)
}

// getJSCall returns js messages along with transferables that can be sent over a MessagePort.
func (c Call) getJS() (messages map[string]interface{}, transferables []interface{}) {
	rc := *(*uintptr)(unsafe.Pointer(&c.RemoteCall))
	messages = map[string]interface{}{
		"rc":     int(rc),
		"output": c.Output,
	}
	transferables = []interface{}{
		c.Output,
	}
	if c.Input != nil {
		messages["input"] = c.Input
		transferables = append(transferables, c.Input)
	}
	return
}

// newCallFromJS constructs a call from javascript arguments.
func newCallFromJS(rc, input, output js.Value) Call {
	rcPtr := uintptr(rc.Int())
	remoteCall := *(*RemoteCall)(unsafe.Pointer(&rcPtr))

	var inputPort *MessagePort
	if input.Truthy() {
		inputPort = NewMessagePort(input)
	}

	return Call{
		RemoteCall: remoteCall,
		Input:      inputPort,
		Output:     NewMessagePort(output),
	}
}
