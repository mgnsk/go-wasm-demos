//go:build js && wasm
// +build js,wasm

package array

import (
	"bytes"
	"fmt"
	"io"
	"syscall/js"
)

type reader struct {
	ab  js.Value
	buf *bytes.Buffer
	eof bool
}

func (r *reader) Read(b []byte) (int, error) {
	if r.eof {
		return 0, io.EOF
	}

	if r.buf != nil {
		return r.buf.Read(b)
	}

	if length := r.ab.Get("byteLength").Int(); len(b) < length {
		buf := make([]byte, length)

		view := js.Global().Get("Uint8Array").New(r.ab)
		n := js.CopyBytesToGo(buf, view)
		if n != length {
			return n, io.ErrUnexpectedEOF
		}

		r.buf = bytes.NewBuffer(buf)

		return r.buf.Read(b)
	}

	view := js.Global().Get("Uint8Array").New(r.ab)
	n := js.CopyBytesToGo(b, view)

	r.eof = true

	return n, nil
}

// NewReader returns a buffered io.Reader for ArrayBuffer.
func NewReader(ab js.Value) io.Reader {
	return &reader{
		ab: ab,
	}
}

// Type if a type of array.
type Type string

func (t Type) String() string {
	return string(t)
}

// Array types.
const (
	Int8Array      Type = "Int8Array"
	Int16Array     Type = "Int16Array"
	Int32Array     Type = "Int32Array"
	BigInt64Array  Type = "BigInt64Array"
	Uint8Array     Type = "Uint8Array"
	Uint16Array    Type = "Uint16Array"
	Uint32Array    Type = "Uint32Array"
	BigUint64Array Type = "BigUint64Array"
	Float32Array   Type = "Float32Array"
	Float64Array   Type = "Float64Array"
)

// TypedArray is a JS TypedArray.
type TypedArray js.Value

// New wraps an ArrayBuffer with TypedArray.
func New(typ Type, buf js.Value) TypedArray {
	switch typ {
	case Int8Array:
	case Int16Array:
	case Int32Array:
	case BigInt64Array:
	case Uint8Array:
	case Uint16Array:
	case Uint32Array:
	case BigUint64Array:
	case Float32Array:
	case Float64Array:
	default:
		panic(fmt.Errorf("New: invalid type '%s'", typ))
	}
	return TypedArray(js.Global().Get(typ.String()).New(buf))
}

// NewFromSlice creates a new read-only TypedArray.
func NewFromSlice(v interface{}) TypedArray {
	b := Encode(v)
	buf := js.Global().Get("ArrayBuffer").New(len(b))
	view := New(Uint8Array, buf)

	if n := js.CopyBytesToJS(view.JSValue(), b); n != len(b) {
		panic(fmt.Errorf("NewArrayBufferFromSlice: copied: %d, expected: %d", n, len(b)))
	}

	switch v.(type) {
	case []int8:
		return New(Int8Array, buf)
	case []int16:
		return New(Int16Array, buf)
	case []int32:
		return New(Int32Array, buf)
	case []int64:
		return New(BigInt64Array, buf)
	case []uint8:
		return view
	case []uint16:
		return New(Uint16Array, buf)
	case []uint32:
		return New(Uint32Array, buf)
	case []uint64:
		return New(BigUint64Array, buf)
	case []float32:
		return New(Float32Array, buf)
	case []float64:
		return New(Float64Array, buf)
	default:
		panic(fmt.Errorf("NewTypedArrayFromSlice: invalid type '%T'", v))
	}
}

// JSValue returns the underlying JS value.
func (a TypedArray) JSValue() js.Value {
	return js.Value(a)
}

// Buffer returns the underlying ArrayBuffer.
func (a TypedArray) Buffer() js.Value {
	return a.JSValue().Get("buffer")
}

// Bytes copies bytes from the underlying ArrayBuffer.
func (a TypedArray) Bytes() []byte {
	b := make([]byte, a.ByteLength())
	if _, err := NewReader(a.Buffer()).Read(b); err != nil {
		panic(err)
	}
	return b
}

// Len returns the length of the array.
func (a TypedArray) Len() int {
	return a.JSValue().Get("length").Int()
}

// ByteLength returns the byte length of the array.
func (a TypedArray) ByteLength() int {
	return a.JSValue().Get("byteLength").Int()
}

// Type returns the type of buffer.
func (a TypedArray) Type() Type {
	return Type(a.JSValue().Get("constructor").Get("name").String())
}
