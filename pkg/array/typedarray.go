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

	length := r.ab.Get("byteLength").Int()
	view := js.Global().Get("Uint8Array").New(r.ab)

	if len(b) <= length {
		n := js.CopyBytesToGo(b, view)
		if n != length {
			return n, io.ErrUnexpectedEOF
		}

		r.eof = true

		return n, nil
	}

	buf := make([]byte, length)
	n := js.CopyBytesToGo(buf, view)
	if n != length {
		return n, io.ErrUnexpectedEOF
	}

	r.buf = bytes.NewBuffer(buf)

	return r.buf.Read(b)
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

// NewInt8Array creates a new Int8Array view over the buffer.
func NewInt8Array(buf js.Value) TypedArray {
	return TypedArray(js.Global().Get(Int8Array.String()).New(buf))
}

// NewInt16Array creates a new Int16Array view over the buffer.
func NewInt16Array(buf js.Value) TypedArray {
	return TypedArray(js.Global().Get(Int16Array.String()).New(buf))
}

// NewInt32Array creates a new Int32Array view over the buffer.
func NewInt32Array(buf js.Value) TypedArray {
	return TypedArray(js.Global().Get(Int32Array.String()).New(buf))
}

// NewBigInt64Array creates a new BigInt64Array view over the buffer.
func NewBigInt64Array(buf js.Value) TypedArray {
	return TypedArray(js.Global().Get(BigInt64Array.String()).New(buf))
}

// NewUint8Array creates a new Uint8Array view over the buffer.
func NewUint8Array(buf js.Value) TypedArray {
	return TypedArray(js.Global().Get(Uint8Array.String()).New(buf))
}

// NewUint16Array creates a new Uint16Array view over the buffer.
func NewUint16Array(buf js.Value) TypedArray {
	return TypedArray(js.Global().Get(Uint16Array.String()).New(buf))
}

// NewUint32Array creates a new Uint32Array view over the buffer.
func NewUint32Array(buf js.Value) TypedArray {
	return TypedArray(js.Global().Get(Uint32Array.String()).New(buf))
}

// NewBigUint64Array creates a new BigUint64Array view over the buffer.
func NewBigUint64Array(buf js.Value) TypedArray {
	return TypedArray(js.Global().Get(BigUint64Array.String()).New(buf))
}

// NewFloat32Array creates a new Float32Array view over the buffer.
func NewFloat32Array(buf js.Value) TypedArray {
	return TypedArray(js.Global().Get(Float32Array.String()).New(buf))
}

// NewFloat64Array creates a new Float64Array view over the buffer.
func NewFloat64Array(buf js.Value) TypedArray {
	return TypedArray(js.Global().Get(Float64Array.String()).New(buf))
}

// NewFromSlice creates a new read-only TypedArray.
func NewFromSlice(v interface{}) TypedArray {
	b := Encode(v)
	buf := js.Global().Get("ArrayBuffer").New(len(b))
	view := NewUint8Array(buf)

	if n := js.CopyBytesToJS(view.JSValue(), b); n != len(b) {
		panic(fmt.Errorf("NewArrayBufferFromSlice: copied: %d, expected: %d", n, len(b)))
	}

	switch v.(type) {
	case []int8:
		return NewInt8Array(buf)
	case []int16:
		return NewInt16Array(buf)
	case []int32:
		return NewInt32Array(buf)
	case []int64:
		return NewBigInt64Array(buf)
	case []uint8:
		return view
	case []uint16:
		return NewUint16Array(buf)
	case []uint32:
		return NewUint32Array(buf)
	case []uint64:
		return NewBigUint64Array(buf)
	case []float32:
		return NewFloat32Array(buf)
	case []float64:
		return NewFloat64Array(buf)
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
