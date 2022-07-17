package array_test

import (
	"testing"

	"github.com/mgnsk/go-wasm-demos/pkg/array"
	. "github.com/onsi/gomega"
	"golang.org/x/exp/constraints"
)

func expectSliceBytes[E constraints.Integer | constraints.Float](g *WithT, s []E, expected []byte) {
	g.Expect(array.Encode(s)).To(Equal(expected))

	var target []E
	array.Decode(&target, expected)

	g.Expect(s).To(Equal(target))
}

func TestSlice(t *testing.T) {
	g := NewGomegaWithT(t)

	expectSliceBytes(g, []int8{-1}, []byte{0xff})
	expectSliceBytes(g, []int16{-1}, []byte{0xff, 0xff})
	expectSliceBytes(g, []int32{-1}, []byte{0xff, 0xff, 0xff, 0xff})
	expectSliceBytes(g, []int64{-1}, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
	expectSliceBytes(g, []uint16{1}, []byte{1, 0})
	expectSliceBytes(g, []uint32{1}, []byte{1, 0, 0, 0})
	expectSliceBytes(g, []uint64{1}, []byte{1, 0, 0, 0, 0, 0, 0, 0})
	expectSliceBytes(g, []float32{-1.0}, []byte{0, 0, 0x80, 0xbf})
	expectSliceBytes(g, []float64{-1.0}, []byte{0, 0, 0, 0, 0, 0, 0xf0, 0xbf})
}
