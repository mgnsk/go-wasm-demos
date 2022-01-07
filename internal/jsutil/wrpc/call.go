//go:build js && wasm
// +build js,wasm

package wrpc

import (
	"fmt"
	"syscall/js"
)

// Call is a remote call that can be executed on a worker.
type Call struct {
	w    *MessagePort
	r    *MessagePort
	call string
}

// NewCall creates a new wrpc call.
func NewCall(w, r *MessagePort, name string) *Call {
	return &Call{
		w:    w,
		r:    r,
		call: name,
	}
}

// NewCallFromJS creates a call from a JS message.
func NewCallFromJS(data js.Value) *Call {
	var r *MessagePort
	if reader := data.Get("reader"); !reader.IsUndefined() {
		r = NewMessagePort(reader)
	}

	return &Call{
		w:    NewMessagePort(data.Get("writer")),
		r:    r,
		call: data.Get("call").String(),
	}
}

// Execute the call locally.
func (c *Call) Execute() {
	call, ok := calls[c.call]
	if !ok {
		panic(fmt.Errorf("call '%s' not found", c.call))
	}
	defer c.w.Close()
	call(c.w, c.r)
}

// JSMessage returns the JS message payload.
func (c *Call) JSMessage() (map[string]interface{}, []interface{}) {
	messages := map[string]interface{}{
		"call":   c.call,
		"writer": c.w.JSValue(),
	}
	transferables := []interface{}{c.w.JSValue()}
	if c.r != nil {
		messages["reader"] = c.r.JSValue()
		if c.r != c.w {
			// Don't sent duplicate conn.
			transferables = append(transferables, c.r.JSValue())
		}
	}
	return messages, transferables
}
