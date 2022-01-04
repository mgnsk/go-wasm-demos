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
	w io.Writer
	r io.Reader

	localWriter io.WriteCloser
	localReader io.Reader
	localDone   *MessagePort

	remoteWriter *MessagePort
	remoteReader *MessagePort
	remoteDone   *MessagePort

	call string
}

// NewCall creates a new wrpc call.
func NewCall(w io.Writer, r io.Reader, name string) *Call {
	c := &Call{
		w:    w,
		r:    r,
		call: name,
	}

	c.localDone, c.remoteDone = Pipe()

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
func NewCallFromJS(data js.Value) *Call {
	var r *MessagePort
	if reader := data.Get("reader"); !reader.IsUndefined() {
		r = NewMessagePort(reader)
	}

	return &Call{
		remoteWriter: NewMessagePort(data.Get("writer")),
		remoteReader: r,
		remoteDone:   NewMessagePort(data.Get("done")),
		call:         data.Get("call").String(),
	}
}

// ExecuteLocal executes the call locally.
// It blocks until the call returns.
func (c *Call) ExecuteLocal() {
	call, ok := calls[c.call]
	if !ok {
		panic(fmt.Errorf("call '%s' not found", c.call))
	}
	defer c.remoteDone.Close()
	defer c.remoteWriter.Close()
	call(c.remoteWriter, c.remoteReader)
}

// ExecuteRemote executes the remote call.
// It blocks until the call returns.
func (c *Call) ExecuteRemote() {
	if c.localReader != nil {
		go mustCopyAll(c.w, c.localReader)
	}

	if c.localWriter != nil {
		go mustCopyAll(c.localWriter, c.r)
	}

	if _, err := c.localDone.ReadMessage(); err != io.EOF {
		panic(err)
	}
}

// JSMessage returns the JS message payload.
func (c *Call) JSMessage() (map[string]interface{}, []interface{}) {
	messages := map[string]interface{}{
		"call":   c.call,
		"writer": c.remoteWriter.JSValue(),
		"done":   c.remoteDone.JSValue(),
	}
	transferables := []interface{}{c.remoteWriter.JSValue(), c.remoteDone.JSValue()}
	if c.remoteReader != nil {
		messages["reader"] = c.remoteReader.JSValue()
		if rr := c.remoteReader; rr != c.remoteWriter {
			// Don't sent duplicate conn.
			transferables = append(transferables, rr.JSValue())
		}
	}
	return messages, transferables
}

func mustCopyAll(dst io.Writer, src io.Reader) {
	if n, err := io.Copy(dst, src); err != nil {
		panic(err)
	} else if n == 0 {
		panic("copyAndClose: zero copy")
	}
	if c, ok := dst.(io.Closer); ok {
		if err := c.Close(); err != nil {
			panic(err)
		}
	}
}
