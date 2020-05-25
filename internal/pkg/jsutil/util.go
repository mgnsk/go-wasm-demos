// +build js,wasm

// Package jsutil provides general functionality for any application running on wasm.
package jsutil

import (
	"fmt"
	"reflect"
	"runtime"
	"syscall/js"
	"unsafe"

	"github.com/davecgh/go-spew/spew"
)

// IsWorker boolean
var (
	IsWorker bool
)

func init() {
	IsWorker = js.Global().Get("WorkerGlobalScope").Type() != js.TypeUndefined
}

// CreateURLObject creates an url object from javascript source.
func CreateURLObject(data interface{}, mime string) js.Value {
	blob := js.Global().Get("Blob").New([]interface{}{data}, map[string]interface{}{"type": mime})
	return js.Global().Get("URL").Call("createObjectURL", blob)
}

// ConsoleLog console.log
func ConsoleLog(args ...interface{}) {
	js.Global().Get("console").Call("log", args...)
}

// Dump deep dumps to console.
func Dump(args ...interface{}) {
	ConsoleLog(spew.Sdump(args...))
}

// Sdump deep prints to string.
func Sdump(args ...interface{}) string {
	return spew.Sdump(args...)
}

// SliceToByteSlice creates a slice of bytes from numeric slices.
func SliceToByteSlice(s interface{}) (b []byte) {
	var h *reflect.SliceHeader
	switch s := s.(type) {
	case []int8:
		h = (*reflect.SliceHeader)(unsafe.Pointer(&s))
	case []int16:
		h = (*reflect.SliceHeader)(unsafe.Pointer(&s))
		h.Len *= 2
		h.Cap *= 2
	case []int32:
		h = (*reflect.SliceHeader)(unsafe.Pointer(&s))
		h.Len *= 4
		h.Cap *= 4
	case []int64:
		h = (*reflect.SliceHeader)(unsafe.Pointer(&s))
		h.Len *= 8
		h.Cap *= 8
	case []uint8:
		return s
	case []uint16:
		h = (*reflect.SliceHeader)(unsafe.Pointer(&s))
		h.Len *= 2
		h.Cap *= 2
	case []uint32:
		h = (*reflect.SliceHeader)(unsafe.Pointer(&s))
		h.Len *= 4
		h.Cap *= 4
	case []uint64:
		h = (*reflect.SliceHeader)(unsafe.Pointer(&s))
		h.Len *= 8
		h.Cap *= 8
	case []float32:
		h = (*reflect.SliceHeader)(unsafe.Pointer(&s))
		h.Len *= 4
		h.Cap *= 4
	case []float64:
		h = (*reflect.SliceHeader)(unsafe.Pointer(&s))
		h.Len *= 8
		h.Cap *= 8
	default:
		panic(fmt.Sprintf("util: unexpected value: %T", s))
	}
	b = *(*[]byte)(unsafe.Pointer(h))
	runtime.KeepAlive(s)
	return
}

// ByteDecoder decodes bytes into various slices.
type ByteDecoder struct{}

func (d *ByteDecoder) Int8Slice(b []byte) []int8 {
	h := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sl := *(*[]int8)(unsafe.Pointer(h))
	runtime.KeepAlive(b)
	return sl
}

func (d *ByteDecoder) Int16Slice(b []byte) []int16 {
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&b))
	// same as int(unsafe.Sizeof(int16(0)))
	header.Len /= 2
	header.Cap /= 2
	data := *(*[]int16)(unsafe.Pointer(&header))
	//runtime.KeepAlive(d.b)
	return data
}

func (d *ByteDecoder) Int32Slice(b []byte) []int32 {
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&b))
	header.Len /= 4
	header.Cap /= 4
	data := *(*[]int32)(unsafe.Pointer(&header))
	//runtime.KeepAlive(b)
	return data
}

func (d *ByteDecoder) Int64Slice(b []byte) []int64 {
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&b))
	header.Len /= 8
	header.Cap /= 8
	data := *(*[]int64)(unsafe.Pointer(&header))
	//runtime.KeepAlive(b)
	return data
}

func (d *ByteDecoder) Uint16Slice(b []byte) []uint16 {
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&b))
	header.Len /= 2
	header.Cap /= 2
	data := *(*[]uint16)(unsafe.Pointer(&header))
	//runtime.KeepAlive(b)
	return data
}

func (d *ByteDecoder) Uint32Slice(b []byte) []uint32 {
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&b))
	header.Len /= 4
	header.Cap /= 4
	data := *(*[]uint32)(unsafe.Pointer(&header))
	//runtime.KeepAlive(b)
	return data
}

func (d *ByteDecoder) Uint64Slice(b []byte) []uint64 {
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&b))
	header.Len /= 8
	header.Cap /= 8
	data := *(*[]uint64)(unsafe.Pointer(&header))
	//runtime.KeepAlive(b)
	return data
}

func (d *ByteDecoder) Float32Slice(b []byte) []float32 {
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&b))
	header.Len /= 4
	header.Cap /= 4
	data := *(*[]float32)(unsafe.Pointer(&header))
	//runtime.KeepAlive(b)
	return data
}

func (d *ByteDecoder) Float64Slice(b []byte) []float64 {
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&b))
	header.Len /= 8
	header.Cap /= 8
	data := *(*[]float64)(unsafe.Pointer(&header))
	//runtime.KeepAlive(b)
	return data
}
