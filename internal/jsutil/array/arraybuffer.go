//go:build js && wasm
// +build js,wasm

package array

import (
	"fmt"
	"syscall/js"

	"github.com/mgnsk/go-wasm-demos/internal/jsutil"
)

// ArrayBuffer is a JS ArrayBuffer.
type ArrayBuffer js.Value

// NewArrayBufferFromSlice creates a new read-only ArrayBuffer from slice.
func NewArrayBufferFromSlice(s interface{}) ArrayBuffer {
	switch slice := s.(type) {
	case []int8,
		[]int16,
		[]int32,
		[]int64,
		[]uint8,
		[]uint16,
		[]uint32,
		[]uint64,
		[]float32,
		[]float64:
		s := jsutil.SliceToByteSlice(slice)
		buf := NewArrayBuffer(len(s))
		if n := js.CopyBytesToJS(buf.Uint8Array().JSValue(), s); n != len(s) {
			panic(fmt.Errorf("NewArrayBufferFromSlice: copied: %d, expected: %d", n, len(s)))
		}
		return buf
	default:
		panic(fmt.Errorf("NewArrayBufferFromSlice: invalid type '%T'", s))
	}
}

// NewArrayBuffer creates a new JS ArrayBuffer.
func NewArrayBuffer(size int) ArrayBuffer {
	return ArrayBuffer(
		js.Global().Get("ArrayBuffer").New(size),
	)
}

// JSValue returns the underlying JS value..
func (a ArrayBuffer) JSValue() js.Value {
	return js.Value(a)
}

// Int8Array view over the array.
func (a ArrayBuffer) Int8Array() TypedArray {
	return TypedArray(
		js.Global().Get("Int8Array").New(a.JSValue(), 0, a.Len()),
	)
}

// Int16Array view over the array.
func (a ArrayBuffer) Int16Array() TypedArray {
	return TypedArray(
		js.Global().Get("Int16Array").New(a.JSValue(), 0, a.Len()/2),
	)
}

// Int32Array view over the array.
func (a ArrayBuffer) Int32Array() TypedArray {
	return TypedArray(
		js.Global().Get("Int32Array").New(a.JSValue(), 0, a.Len()/4),
	)
}

// BigInt64Array view over the array.
func (a ArrayBuffer) BigInt64Array() TypedArray {
	return TypedArray(
		js.Global().Get("BigInt64Array").New(a.JSValue(), 0, a.Len()/8),
	)
}

// Uint8Array view over the array buffer.
func (a ArrayBuffer) Uint8Array() TypedArray {
	return TypedArray(
		js.Global().Get("Uint8Array").New(a.JSValue(), 0, a.Len()),
	)
}

// Uint16Array view over the array buffer.
func (a ArrayBuffer) Uint16Array() TypedArray {
	return TypedArray(
		js.Global().Get("Uint16Array").New(a.JSValue(), 0, a.Len()/2),
	)
}

// Uint32Array view over the array buffer.
func (a ArrayBuffer) Uint32Array() TypedArray {
	return TypedArray(
		js.Global().Get("Uint32Array").New(a.JSValue(), 0, a.Len()/4),
	)
}

// BigUint64Array view over the array.
func (a ArrayBuffer) BigUint64Array() TypedArray {
	return TypedArray(
		js.Global().Get("BigUint64Array").New(a.JSValue(), 0, a.Len()/8),
	)
}

// Float32Array view over the array buffer.
func (a ArrayBuffer) Float32Array() TypedArray {
	return TypedArray(
		js.Global().Get("Float32Array").New(a.JSValue(), 0, a.Len()/4),
	)
}

// Float64Array view over the array buffer.
func (a ArrayBuffer) Float64Array() TypedArray {
	return TypedArray(
		js.Global().Get("Float64Array").New(a.JSValue(), 0, a.Len()/8),
	)
}

// Len returns the length of byte array.
func (a ArrayBuffer) Len() int {
	return a.JSValue().Get("byteLength").Int()
}

// Bytes returns the ArrayBuffer bytes.
func (a ArrayBuffer) Bytes() []byte {
	buf := make([]byte, a.Len())
	if n := js.CopyBytesToGo(buf, a.Uint8Array().JSValue()); n != len(buf) {
		panic("CopyBytesToGo: invalid copied length")
	}
	return buf
}
