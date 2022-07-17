package array

import (
	"bytes"
	"fmt"
	"io"
	"syscall/js"

	"golang.org/x/exp/constraints"
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

	abLen := r.ab.Get("byteLength").Int()
	view := js.Global().Get("Uint8Array").New(r.ab)

	if len(b) >= abLen {
		// Can fit exactly.
		n := js.CopyBytesToGo(b, view)
		if n != abLen {
			return n, io.ErrUnexpectedEOF
		}

		r.eof = true

		return n, nil
	}

	// Not enough room.
	buf := make([]byte, abLen)
	n := js.CopyBytesToGo(buf, view)
	if n != abLen {
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

// TypedArray is a JS TypedArray.
type TypedArray struct {
	js.Value
}

// NewInt8Array creates a new Int8Array view over the buffer.
func NewInt8Array(ab js.Value) TypedArray {
	return TypedArray{js.Global().Get("Int8Array").New(ab)}
}

// NewInt16Array creates a new Int16Array view over the buffer.
func NewInt16Array(ab js.Value) TypedArray {
	return TypedArray{js.Global().Get("Int16Array").New(ab)}
}

// NewInt32Array creates a new Int32Array view over the buffer.
func NewInt32Array(ab js.Value) TypedArray {
	return TypedArray{js.Global().Get("Int32Array").New(ab)}
}

// NewBigInt64Array creates a new BigInt64Array view over the buffer.
func NewBigInt64Array(ab js.Value) TypedArray {
	return TypedArray{js.Global().Get("BigInt64Array").New(ab)}
}

// NewUint8Array creates a new Uint8Array view over the buffer.
func NewUint8Array(ab js.Value) TypedArray {
	return TypedArray{js.Global().Get("Uint8Array").New(ab)}
}

// NewUint16Array creates a new Uint16Array view over the buffer.
func NewUint16Array(ab js.Value) TypedArray {
	return TypedArray{js.Global().Get("Uint16Array").New(ab)}
}

// NewUint32Array creates a new Uint32Array view over the buffer.
func NewUint32Array(ab js.Value) TypedArray {
	return TypedArray{js.Global().Get("Uint32Array").New(ab)}
}

// NewBigUint64Array creates a new BigUint64Array view over the buffer.
func NewBigUint64Array(ab js.Value) TypedArray {
	return TypedArray{js.Global().Get("BigUint64Array").New(ab)}
}

// NewFloat32Array creates a new Float32Array view over the buffer.
func NewFloat32Array(ab js.Value) TypedArray {
	return TypedArray{js.Global().Get("Float32Array").New(ab)}
}

// NewFloat64Array creates a new Float64Array view over the buffer.
func NewFloat64Array(ab js.Value) TypedArray {
	return TypedArray{js.Global().Get("Float64Array").New(ab)}
}

// NewFromSlice creates a new read-only TypedArray.
func NewFromSlice[E constraints.Integer | constraints.Float](s []E) TypedArray {
	b := Encode(s)
	ab := js.Global().Get("ArrayBuffer").New(len(b))
	view := NewUint8Array(ab)

	if n := js.CopyBytesToJS(view.Value, b); n != len(b) {
		panic(fmt.Errorf("NewArrayBufferFromSlice: copied: %d, expected: %d", n, len(b)))
	}

	switch any(E(0)).(type) {
	case int8:
		return NewInt8Array(ab)
	case int16:
		return NewInt16Array(ab)
	case int32:
		return NewInt32Array(ab)
	case int64:
		return NewBigInt64Array(ab)
	case uint8:
		return view
	case uint16:
		return NewUint16Array(ab)
	case uint32:
		return NewUint32Array(ab)
	case uint64:
		return NewBigUint64Array(ab)
	case float32:
		return NewFloat32Array(ab)
	case float64:
		return NewFloat64Array(ab)
	default:
		panic(fmt.Errorf("NewTypedArrayFromSlice: invalid type '%T'", s))
	}
}

// ArrayBuffer returns the underlying ArrayBuffer.
func (a TypedArray) ArrayBuffer() js.Value {
	return a.Get("buffer")
}

// Bytes copies bytes from the underlying ArrayBuffer.
func (a TypedArray) Bytes() []byte {
	b := make([]byte, a.ByteLength())
	if _, err := NewReader(a.ArrayBuffer()).Read(b); err != nil {
		panic(err)
	}
	return b
}

// Len returns the length of the array.
func (a TypedArray) Len() int {
	return a.Get("length").Int()
}

// ByteLength returns the byte length of the array.
func (a TypedArray) ByteLength() int {
	return a.Get("byteLength").Int()
}

// Type returns the type of the array.
func (a TypedArray) Type() string {
	return a.Get("constructor").Get("name").String()
}
