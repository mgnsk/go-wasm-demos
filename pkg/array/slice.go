package array

import (
	"reflect"
	"runtime"
	"unsafe"

	"golang.org/x/exp/constraints"
)

// Encode a numeric slice into bytes.
func Encode[E constraints.Integer | constraints.Float](s []E) []byte {
	h := (*reflect.SliceHeader)(unsafe.Pointer(&s))
	h.Len *= int(unsafe.Sizeof(E(0)))
	h.Cap *= int(unsafe.Sizeof(E(0)))
	b := *(*[]byte)(unsafe.Pointer(h))

	runtime.KeepAlive(s)

	return b
}

// Decode bytes into target numeric slice.
func Decode[E constraints.Integer | constraints.Float](target *[]E, b []byte) {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	h := (*reflect.SliceHeader)(unsafe.Pointer(target))
	h.Data = bh.Data
	h.Len = bh.Len / int(unsafe.Sizeof(E(0)))
	h.Cap = bh.Cap / int(unsafe.Sizeof(E(0)))

	runtime.KeepAlive(b)
}
