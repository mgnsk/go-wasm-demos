package array_test

import (
	"reflect"
	"testing"

	"github.com/mgnsk/go-wasm-demos/pkg/array"
	. "github.com/onsi/gomega"
)

func TestSlice(t *testing.T) {
	g := NewGomegaWithT(t)

	for _, tc := range []struct {
		data     interface{}
		expected []byte
	}{
		{[]int8{-1}, []byte{0xff}},
		{[]int16{-1}, []byte{0xff, 0xff}},
		{[]int32{-1}, []byte{0xff, 0xff, 0xff, 0xff}},
		{[]int64{-1}, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{[]uint16{1}, []byte{1, 0}},
		{[]uint32{1}, []byte{1, 0, 0, 0}},
		{[]uint64{1}, []byte{1, 0, 0, 0, 0, 0, 0, 0}},
		{[]float32{-1.0}, []byte{0, 0, 0x80, 0xbf}},
		{[]float64{-1.0}, []byte{0, 0, 0, 0, 0, 0, 0xf0, 0xbf}},
	} {
		g.Expect(array.Encode(tc.data)).To(Equal(tc.expected))

		target := reflect.New(reflect.TypeOf(tc.data))
		array.Decode(target.Interface(), tc.expected)

		g.Expect(tc.data).To(Equal(target.Elem().Interface()))
	}
}
