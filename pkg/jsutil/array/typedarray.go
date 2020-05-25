// +build js,wasm

package array

import (
	"errors"
	"fmt"
	"syscall/js"
)

// TypedArray is a JS wrapper for typed array.
type TypedArray js.Value

// CreateTypedArrayFromSlice copies and creates a read only ArrayBuffer buffer.
func CreateTypedArrayFromSlice(s interface{}) (TypedArray, error) {
	ab, err := CreateBufferFromSlice(s)
	if err != nil {
		return TypedArray(js.Null()), err
	}

	switch s.(type) {
	case []int8:
		return ab.Int8Array(), nil
	case []int16:
		return ab.Int16Array(), nil
	case []int32:
		return ab.Int32Array(), nil
	case []int64:
		return ab.BigInt64Array(), nil
	case []uint8:
		return ab.Uint8Array(), nil
	case []uint16:
		return ab.Uint16Array(), nil
	case []uint32:
		return ab.Uint32Array(), nil
	case []uint64:
		return ab.BigUint64Array(), nil
	case []float32:
		return ab.Float32Array(), nil
	case []float64:
		return ab.Float64Array(), nil
	default:
		return TypedArray(js.Null()), fmt.Errorf("CreateTypedArrayFromSlice: invalid type")
	}
}

// JSValue returns the underlying js value.
func (a TypedArray) JSValue() js.Value {
	return js.Value(a)
}

// Buffer returns the underlying ArrayBuffer.
func (a TypedArray) Buffer() js.Value {
	return a.JSValue().Get("buffer")
}

// ByteLength returns the length of underlying data.
func (a TypedArray) ByteLength() int {
	return a.JSValue().Get("byteLength").Int()
}

// Type returns the type of buffer.
func (a TypedArray) Type() string {
	return a.JSValue().Get("constructor").Get("name").String()
}

// CopyBytes copies out bytes from js typed array.
func (a TypedArray) CopyBytes() ([]byte, error) {
	length := a.ByteLength()
	if length == 0 {
		return nil, errors.New("CopyBytes: 0 copy")
	}
	b := make([]byte, length)
	buf := Buffer(a.Buffer()).Uint8Array()
	if n := js.CopyBytesToGo(b, buf.JSValue()); n != length {
		return nil, fmt.Errorf("CopyBytes: copied: %d, expected: %d", n, length)
	}
	return b, nil
}

// Must panics on error.
func Must(arr TypedArray, err error) TypedArray {
	if err != nil {
		panic(err)
	}
	return arr
}
