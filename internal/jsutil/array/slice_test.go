package array_test

import (
	"reflect"

	"github.com/mgnsk/go-wasm-demos/internal/jsutil/array"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("encoding and decoding slices", func() {
	DescribeTable("data table",
		func(data interface{}, expected []byte) {
			Expect(array.Encode(data)).To(Equal(expected))

			target := reflect.New(reflect.TypeOf(data))
			array.Decode(target.Interface(), expected)

			Expect(data).To(Equal(target.Elem().Interface()))
		},
		Entry(
			"[]int8",
			[]int8{-1},
			[]byte{0xff},
		),
		Entry(
			"[]int16",
			[]int16{-1},
			[]byte{0xff, 0xff},
		),
		Entry(
			"[]int32",
			[]int32{-1},
			[]byte{0xff, 0xff, 0xff, 0xff},
		),
		Entry(
			"[]int64",
			[]int64{-1},
			[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		),
		Entry(
			"[]uint16",
			[]uint16{1},
			[]byte{1, 0},
		),
		Entry(
			"[]uint32",
			[]uint32{1},
			[]byte{1, 0, 0, 0},
		),
		Entry(
			"[]uint64",
			[]uint64{1},
			[]byte{1, 0, 0, 0, 0, 0, 0, 0},
		),
		Entry(
			"[]float32",
			[]float32{-1.0},
			[]byte{0, 0, 0x80, 0xbf},
		),
		Entry(
			"[]float64",
			[]float64{-1.0},
			[]byte{0, 0, 0, 0, 0, 0, 0xf0, 0xbf},
		),
	)
})
