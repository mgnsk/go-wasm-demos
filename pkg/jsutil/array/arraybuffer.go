// +build js,wasm

package array

import (
	"errors"
	"fmt"
	"syscall/js"

	"github.com/mgnsk/jsutil"
)

// TODO documentation is a bit rough.

// Buffer is a JS ArrayBuffer.
type Buffer js.Value

func createBuffer(slice interface{}, byteLength int) (Buffer, error) {
	if slice == nil {
		return Buffer(js.Null()), fmt.Errorf("createBuffer: slice must not be ni")
	}

	uint8Array := NewBuffer(byteLength).Uint8Array()
	if n := js.CopyBytesToJS(uint8Array.JSValue(), jsutil.SliceToByteSlice(slice)); n != byteLength {
		return Buffer{}, fmt.Errorf("createBuffer: copied: %d, expected: %d", n, byteLength)
	}
	return Buffer(uint8Array.Buffer()), nil
}

// CreateBufferFromSlice copies and creates a read only Buffer buffer.
func CreateBufferFromSlice(s interface{}) (Buffer, error) {
	switch s := s.(type) {
	case []int8:
		return createBuffer(s, len(s))
	case []int16:
		return createBuffer(s, len(s)*2)
	case []int32:
		return createBuffer(s, len(s)*4)
	case []int64:
		return createBuffer(s, len(s)*8)
	case []uint8:
		return createBuffer(s, len(s))
	case []uint16:
		return createBuffer(s, len(s)*2)
	case []uint32:
		return createBuffer(s, len(s)*4)
	case []uint64:
		return createBuffer(s, len(s)*8)
	case []float32:
		return createBuffer(s, len(s)*4)
	case []float64:
		return createBuffer(s, len(s)*8)
	default:
		return Buffer(js.Null()), errors.New("CreateBufferFromSlice: invalid type")
	}
}

// NewBuffer creates a new JS byte buffer.
func NewBuffer(size int) Buffer {
	return Buffer(
		js.Global().Get("ArrayBuffer").New(size),
	)
}

// JSValue returns JS value for a.
func (a Buffer) JSValue() js.Value {
	return js.Value(a)
}

// Int8Array view over the array.
func (a Buffer) Int8Array() TypedArray {
	return TypedArray(
		js.Global().Get("Int8Array").New(a.JSValue(), 0, a.ByteLength()),
	)
}

// Int16Array view over the array.
func (a Buffer) Int16Array() TypedArray {
	return TypedArray(
		js.Global().Get("Int16Array").New(a.JSValue(), 0, a.ByteLength()/2),
	)
}

// Int32Array view over the array.
func (a Buffer) Int32Array() TypedArray {
	return TypedArray(
		js.Global().Get("Int32Array").New(a.JSValue(), 0, a.ByteLength()/4),
	)
}

// BigInt64Array view over the array.
func (a Buffer) BigInt64Array() TypedArray {
	return TypedArray(
		js.Global().Get("BigInt64Array").New(a.JSValue(), 0, a.ByteLength()/8),
	)
}

// Uint8Array view over the array buffer.
func (a Buffer) Uint8Array() TypedArray {
	return TypedArray(
		js.Global().Get("Uint8Array").New(a.JSValue(), 0, a.ByteLength()),
	)
}

// Uint16Array view over the array buffer.
func (a Buffer) Uint16Array() TypedArray {
	return TypedArray(
		js.Global().Get("Uint16Array").New(a.JSValue(), 0, a.ByteLength()/2),
	)
}

// Uint32Array view over the array buffer.
func (a Buffer) Uint32Array() TypedArray {
	return TypedArray(
		js.Global().Get("Uint32Array").New(a.JSValue(), 0, a.ByteLength()/4),
	)
}

// BigUint64Array view over the array.
func (a Buffer) BigUint64Array() TypedArray {
	return TypedArray(
		js.Global().Get("BigUint64Array").New(a.JSValue(), 0, a.ByteLength()/8),
	)
}

// Float32Array view over the array buffer.
func (a Buffer) Float32Array() TypedArray {
	return TypedArray(
		js.Global().Get("Float32Array").New(a.JSValue(), 0, a.ByteLength()/4),
	)
}

// Float64Array view over the array buffer.
func (a Buffer) Float64Array() TypedArray {
	return TypedArray(
		js.Global().Get("Float64Array").New(a.JSValue(), 0, a.ByteLength()/8),
	)
}

// ByteLength returns the length of byte array.
func (a Buffer) ByteLength() int {
	return a.JSValue().Get("byteLength").Int()
}

// CopyBytes copies out bytes from js array buffer.
func (a Buffer) CopyBytes() ([]byte, error) {
	length := a.ByteLength()
	if length == 0 {
		return nil, errors.New("CopyBytes: 0 copy")
	}
	b := make([]byte, length)
	if n := js.CopyBytesToGo(b, a.Uint8Array().JSValue()); n != length {
		return nil, fmt.Errorf("CopyBytes: copied: %d, expected: %d", n, length)
	}
	return b, nil
}

// Write bytes into array.
func (a Buffer) Write(p []byte) (n int, err error) {
	if n := js.CopyBytesToJS(a.Uint8Array().JSValue(), p); n < len(p) {
		return 0, fmt.Errorf("Write: copied: %d, expected: %d", n, len(p))
	}
	return n, nil
}
