//go:build js && wasm
// +build js,wasm

package wrpc

import (
	"fmt"
	"syscall/js"
)

// Call is a remote call that can be executed on a worker.
type Call struct {
	output *MessagePort
	input  *MessagePort
	name   string
}

// NewCall creates a new wrpc call.
func NewCall(output, input *MessagePort, name string) Call {
	return Call{
		output: output,
		input:  input,
		name:   name,
	}
}

// NewCallFromJS creates a call from a JS message.
func NewCallFromJS(data js.Value) Call {
	var r *MessagePort
	if reader := data.Get("input"); !reader.IsUndefined() {
		r = NewMessagePort(reader)
	}

	return Call{
		output: NewMessagePort(data.Get("output")),
		input:  r,
		name:   data.Get("call").String(),
	}
}

// Execute the call locally.
func (c Call) Execute() {
	call, ok := calls[c.name]
	if !ok {
		panic(fmt.Errorf("call '%s' not found", c.name))
	}
	defer c.output.Close()
	call(c.output, c.input)
}

// JSMessage returns the JS message payload.
func (c Call) JSMessage() (map[string]interface{}, []interface{}) {
	messages := map[string]interface{}{
		"call":   c.name,
		"output": c.output.JSValue(),
	}
	transferables := []interface{}{c.output.JSValue()}
	if c.input != nil {
		messages["input"] = c.input.JSValue()
		if c.input != c.output {
			// Don't sent duplicate conn.
			transferables = append(transferables, c.input.JSValue())
		}
	}
	return messages, transferables
}
