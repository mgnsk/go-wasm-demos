//go:build js && wasm
// +build js,wasm

package wrpc

import (
	"io"
	"syscall/js"
	"unsafe"
)

// Call is a remote call that can be scheduled to a worker.
type Call struct {
	w io.WriteCloser
	r io.Reader

	localWriter io.WriteCloser
	localReader io.Reader

	remoteWriter *Conn
	remoteReader *Conn

	rc int
}

// NewCall creates a new wrpc call.
func NewCall(w io.WriteCloser, r io.Reader, f RemoteCall) Call {
	c := Call{
		w:  w,
		r:  r,
		rc: int(*(*uintptr)(unsafe.Pointer(&f))),
	}

	if conn, ok := w.(*Conn); ok {
		c.remoteWriter = conn
	} else {
		c.remoteWriter, c.localReader = ConnPipe()
	}

	if r != nil {
		if conn, ok := r.(*Conn); ok {
			c.remoteReader = conn
		} else {
			c.remoteReader, c.localWriter = ConnPipe()
		}
	}

	return c
}

// NewCallFromJS constructs a call from JS message.
func NewCallFromJS(data js.Value) Call {
	writer := data.Get("output")
	reader := data.Get("input")
	rc := data.Get("rc").Int()

	w := NewConn(NewMessagePort(writer))

	var r *Conn
	if !reader.IsUndefined() {
		r = NewConn(NewMessagePort(reader))
	}

	return Call{
		remoteWriter: w,
		remoteReader: r,
		rc:           rc,
	}
}

// ExecuteLocal executes the call locally.
func (c Call) ExecuteLocal() {
	rcPtr := uintptr(c.rc)
	f := *(*RemoteCall)(unsafe.Pointer(&rcPtr))
	f(c.remoteWriter, c.remoteReader)
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
		"rc":     c.rc,
		"output": c.remoteWriter.JSValue(),
	}
	transferables := []interface{}{c.remoteWriter.JSValue()}
	if c.remoteReader != nil {
		messages["input"] = c.remoteReader.JSValue()
		transferables = append(transferables, c.remoteReader.JSValue())
	}
	return messages, transferables
}
