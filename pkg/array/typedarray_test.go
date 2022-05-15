package array_test

import (
	"reflect"
	"testing"

	"github.com/mgnsk/go-wasm-demos/pkg/array"
	. "github.com/onsi/gomega"
)

func TestTypedArray(t *testing.T) {
	g := NewGomegaWithT(t)

	for _, tc := range []struct {
		typ  array.Type
		data interface{}
	}{
		{array.Int8Array, []int8{-1, 0, 1}},
		{array.Int16Array, []int16{-1, 0, 1}},
		{array.Int32Array, []int32{-1}},
		{array.BigInt64Array, []int64{-1}},
		{array.Uint8Array, []uint8{1}},
		{array.Uint16Array, []uint16{1}},
		{array.Uint32Array, []uint32{1}},
		{array.BigUint64Array, []uint64{1}},
		{array.Float32Array, []float32{-1.0}},
		{array.Float64Array, []float64{-1.0}},
	} {
		arr := array.NewFromSlice(tc.data)
		g.Expect(arr.Type()).To(Equal(tc.typ))
		g.Expect(reflect.ValueOf(tc.data).Len()).To(Equal(arr.Len()))
	}
}
