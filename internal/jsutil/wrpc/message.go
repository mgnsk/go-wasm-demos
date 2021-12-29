package wrpc

import (
	"fmt"
	"syscall/js"

	"github.com/mgnsk/go-wasm-demos/internal/jsutil/array"
)

// MessageReader reads an array buffer data message.
type MessageReader struct {
	value js.Value
}

// NewMessageReader creates a new message reader.
func NewMessageReader(value js.Value) MessageReader {
	return MessageReader{value}
}

// Read the message if it is an array buffer message.
func (r MessageReader) Read(b []byte) (n int, err error) {
	arr := r.value.Get("arr")
	if arr.IsUndefined() {
		return 0, fmt.Errorf("not an array buffer message")
	}
	return array.Buffer(arr).Read(b)
}
