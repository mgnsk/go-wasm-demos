//go:build js && wasm
// +build js,wasm

package wrpc

import (
	"fmt"
	"io"
	"syscall/js"
)

// Call is a remote call that can be scheduled to a worker.
type Call struct {
	w io.WriteCloser
	r io.Reader

	localWriter io.WriteCloser
	localReader io.Reader

	remoteWriter *MessagePort
	remoteReader *MessagePort

	call string
}

// NewCall creates a new wrpc call.
func NewCall(w io.WriteCloser, r io.Reader, name string) Call {
	c := Call{
		w:    w,
		r:    r,
		call: name,
	}

	if p, ok := w.(*MessagePort); ok {
		c.remoteWriter = p
	} else {
		c.remoteWriter, c.localReader = Pipe()
	}

	if r != nil {
		if conn, ok := r.(*MessagePort); ok {
			c.remoteReader = conn
		} else {
			c.remoteReader, c.localWriter = Pipe()
		}
	}

	return c
}

// NewCallFromJS constructs a call from JS message.
func NewCallFromJS(data js.Value) Call {
	w := NewMessagePort(data.Get("writer"))

	var r *MessagePort
	if reader := data.Get("reader"); !reader.IsUndefined() {
		r = NewMessagePort(reader)
	}

	return Call{
		remoteWriter: w,
		remoteReader: r,
		call:         data.Get("call").String(),
	}
}

// ExecuteLocal executes the call locally.
func (c Call) ExecuteLocal() {
	call, ok := calls[c.call]
	if !ok {
		panic(fmt.Errorf("call '%s' not found", c.call))
	}
	call(c.remoteWriter, c.remoteReader)
}

// ExecuteRemote executes the remote call.
func (c Call) ExecuteRemote() {
	if c.localReader != nil {
		go mustCopy(c.w, c.localReader)
	}
	if c.localWriter != nil {
		go mustCopy(c.localWriter, c.r)
	}
}

// JSMessage returns the JS message payload.
func (c Call) JSMessage() (map[string]interface{}, []interface{}) {
	messages := map[string]interface{}{
		"call":   c.call,
		"writer": c.remoteWriter.JSValue(),
	}
	transferables := []interface{}{c.remoteWriter.JSValue()}
	if c.remoteReader != nil {
		messages["reader"] = c.remoteReader.JSValue()
		if rr := c.remoteReader; rr != c.remoteWriter {
			// Don't sent duplicate conn.
			transferables = append(transferables, rr.JSValue())
		}
	}
	return messages, transferables
}

func mustCopy(dst io.WriteCloser, src io.Reader) {
	defer dst.Close()
	if n, err := io.Copy(dst, src); err != nil {
		panic(err)
	} else if n == 0 {
		panic("copyAndClose: zero copy")
	}
}
