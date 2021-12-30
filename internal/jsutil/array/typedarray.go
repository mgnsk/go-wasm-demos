//go:build js && wasm
// +build js,wasm

package array

import (
	"fmt"
	"syscall/js"
)

// TypedArray is a JS TypedArray.
type TypedArray js.Value

// NewTypedArrayFromSlice creates a new read-only TypedArray.
func NewTypedArrayFromSlice(s interface{}) TypedArray {
	ab := NewArrayBufferFromSlice(s)
	switch s.(type) {
	case []int8:
		return ab.Int8Array()
	case []int16:
		return ab.Int16Array()
	case []int32:
		return ab.Int32Array()
	case []int64:
		return ab.BigInt64Array()
	case []uint8:
		return ab.Uint8Array()
	case []uint16:
		return ab.Uint16Array()
	case []uint32:
		return ab.Uint32Array()
	case []uint64:
		return ab.BigUint64Array()
	case []float32:
		return ab.Float32Array()
	case []float64:
		return ab.Float64Array()
	default:
		panic(fmt.Errorf("NewTypedArrayFromSlice: invalid type '%T'", s))
	}
}

// JSValue returns the underlying JS value.
func (a TypedArray) JSValue() js.Value {
	return js.Value(a)
}

// ArrayBuffer returns the underlying ArrayBuffer.
func (a TypedArray) Buffer() ArrayBuffer {
	return ArrayBuffer(a.JSValue().Get("buffer"))
}

// Type returns the type of buffer.
func (a TypedArray) Type() string {
	return a.JSValue().Get("constructor").Get("name").String()
}
