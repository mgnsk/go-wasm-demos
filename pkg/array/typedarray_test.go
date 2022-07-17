package array_test

import (
	"testing"

	"github.com/mgnsk/go-wasm-demos/pkg/array"
	. "github.com/onsi/gomega"
	"golang.org/x/exp/constraints"
)

func expectedSliceType[E constraints.Integer | constraints.Float](g *WithT, s []E, expectedType string) {
	arr := array.NewFromSlice(s)
	g.Expect(arr.Type()).To(Equal(expectedType))
	g.Expect(s).To(HaveLen(arr.Len()))
}

func TestTypedArray(t *testing.T) {
	g := NewGomegaWithT(t)

	expectedSliceType(g, []int8{-1, 0, 1}, "Int8Array")
	expectedSliceType(g, []int16{-1, 0, 1}, "Int16Array")
	expectedSliceType(g, []int32{-1}, "Int32Array")
	expectedSliceType(g, []int64{-1}, "BigInt64Array")
	expectedSliceType(g, []uint8{1}, "Uint8Array")
	expectedSliceType(g, []uint16{1}, "Uint16Array")
	expectedSliceType(g, []uint32{1}, "Uint32Array")
	expectedSliceType(g, []uint64{1}, "BigUint64Array")
	expectedSliceType(g, []float32{-1.0}, "Float32Array")
	expectedSliceType(g, []float64{-1.0}, "Float64Array")
}
