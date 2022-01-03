//go:build js && wasm
// +build js,wasm

package array_test

import (
	"reflect"

	"github.com/mgnsk/go-wasm-demos/internal/jsutil/array"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("TypedArray", func() {
	DescribeTable("has correct size",
		func(data interface{}) {
			arr := array.NewTypedArrayFromSlice(data)
			Expect(reflect.ValueOf(data).Len()).To(Equal(arr.Len()))
		},
		Entry(
			"[]int8",
			[]int8{-1, 0, 1},
		),
		Entry(
			"[]int16",
			[]int16{-1, 0, 1},
		),
		Entry(
			"[]int32",
			[]int32{-1},
		),
		Entry(
			"[]int64",
			[]int64{-1},
		),
		Entry(
			"[]uint8",
			[]uint8{1},
		),
		Entry(
			"[]uint16",
			[]uint16{1},
		),
		Entry(
			"[]uint32",
			[]uint32{1},
		),
		Entry(
			"[]uint64",
			[]uint64{1},
		),
		Entry(
			"[]float32",
			[]float32{-1.0},
		),
		Entry(
			"[]float64",
			[]float64{-1.0},
		),
	)
})
