// +build js,wasm

package array_test

import (
	"reflect"
	"strings"

	"github.com/mgnsk/jsutil"
	"github.com/mgnsk/jsutil/array"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("TypedArray", func() {
	Context("Standard slice copied to JS typed buffer and back", func() {
		DescribeTable("data table",
			func(data interface{}) {
				a, err := array.CreateTypedArrayFromSlice(data)
				Expect(err).To(BeNil())

				b, err := a.CopyBytes()
				Expect(b).NotTo(BeEmpty())
				Expect(err).To(BeNil())

				if s, ok := data.([]byte); ok {
					Expect(s).To(Equal(b))
					return
				}

				// Get the first slice element.
				rv := reflect.ValueOf(data)
				element := rv.Index(0)
				elementType := element.Type().String()

				decoder := &jsutil.ByteDecoder{}
				rd := reflect.ValueOf(decoder)
				methodName := strings.Title(elementType) + "Slice"
				method := rd.MethodByName(methodName)
				results := method.Call([]reflect.Value{reflect.ValueOf(b)})

				Expect(data).To(Equal(results[0].Interface()))
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
})
