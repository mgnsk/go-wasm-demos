package array

import (
	"fmt"
	"reflect"
	"runtime"
	"unsafe"
)

// Encode a numeric slice into bytes.
func Encode(s interface{}) []byte {
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
		panic(fmt.Sprintf("Encode: invalid type: %T", s))
	}
	b := *(*[]byte)(unsafe.Pointer(h))
	runtime.KeepAlive(s)
	return b
}

// Decode bytes into target numeric slice.
func Decode(target interface{}, b []byte) {
	h := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	switch target := target.(type) {
	case *[]int8:
		sh := (*reflect.SliceHeader)(unsafe.Pointer(target))
		sh.Data = h.Data
		sh.Len = h.Len
		sh.Cap = h.Cap
	case *[]int16:
		sh := (*reflect.SliceHeader)(unsafe.Pointer(target))
		sh.Data = h.Data
		sh.Len = h.Len / 2
		sh.Cap = h.Cap / 2
	case *[]int32:
		sh := (*reflect.SliceHeader)(unsafe.Pointer(target))
		sh.Data = h.Data
		sh.Len = h.Len / 4
		sh.Cap = h.Cap / 4
	case *[]int64:
		sh := (*reflect.SliceHeader)(unsafe.Pointer(target))
		sh.Data = h.Data
		sh.Len = h.Len / 8
		sh.Cap = h.Cap / 8
	case *[]uint8:
		sh := (*reflect.SliceHeader)(unsafe.Pointer(target))
		sh.Data = h.Data
		sh.Len = h.Len
		sh.Cap = h.Cap
	case *[]uint16:
		sh := (*reflect.SliceHeader)(unsafe.Pointer(target))
		sh.Data = h.Data
		sh.Len = h.Len / 2
		sh.Cap = h.Cap / 2
	case *[]uint32:
		sh := (*reflect.SliceHeader)(unsafe.Pointer(target))
		sh.Data = h.Data
		sh.Len = h.Len / 4
		sh.Cap = h.Cap / 4
	case *[]uint64:
		sh := (*reflect.SliceHeader)(unsafe.Pointer(target))
		sh.Data = h.Data
		sh.Len = h.Len / 8
		sh.Cap = h.Cap / 8
	case *[]float32:
		sh := (*reflect.SliceHeader)(unsafe.Pointer(target))
		sh.Data = h.Data
		sh.Len = h.Len / 4
		sh.Cap = h.Cap / 4
	case *[]float64:
		sh := (*reflect.SliceHeader)(unsafe.Pointer(target))
		sh.Data = h.Data
		sh.Len = h.Len / 8
		sh.Cap = h.Cap / 8
	default:
		panic(fmt.Sprintf("Decode: invalid type: %T", target))
	}
	runtime.KeepAlive(b)
}
