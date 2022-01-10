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
		func(typ array.Type, data interface{}) {
			arr := array.NewTypedArrayFromSlice(data)
			Expect(arr.Type()).To(Equal(typ))
			Expect(reflect.ValueOf(data).Len()).To(Equal(arr.Len()))

		},
		Entry(
			"[]int8",
			array.Int8Array,
			[]int8{-1, 0, 1},
		),
		Entry(
			"[]int16",
			array.Int16Array,
			[]int16{-1, 0, 1},
		),
		Entry(
			"[]int32",
			array.Int32Array,
			[]int32{-1},
		),
		Entry(
			"[]int64",
			array.BigInt64Array,
			[]int64{-1},
		),
		Entry(
			"[]uint8",
			array.Uint8Array,
			[]uint8{1},
		),
		Entry(
			"[]uint16",
			array.Uint16Array,
			[]uint16{1},
		),
		Entry(
			"[]uint32",
			array.Uint32Array,
			[]uint32{1},
		),
		Entry(
			"[]uint64",
			array.BigUint64Array,
			[]uint64{1},
		),
		Entry(
			"[]float32",
			array.Float32Array,
			[]float32{-1.0},
		),
		Entry(
			"[]float64",
			array.Float64Array,
			[]float64{-1.0},
		),
	)
})
