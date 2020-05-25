package jsutil_test

import (
	"reflect"
	"strings"

	"github.com/mgnsk/jsutil"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("JSUtil", func() {
	Context("Converting slices to bytes and back without copying", func() {
		DescribeTable("data table",
			func(data interface{}, expected []byte) {
				Expect(jsutil.SliceToByteSlice(data)).To(Equal(expected))

				// Get the first slice element.
				rv := reflect.ValueOf(data)
				element := rv.Index(0)
				elementType := element.Type().String()

				decoder := &jsutil.ByteDecoder{}
				rd := reflect.ValueOf(decoder)
				methodName := strings.Title(elementType) + "Slice"
				method := rd.MethodByName(methodName)
				results := method.Call([]reflect.Value{reflect.ValueOf(expected)})

				Expect(data).To(Equal(results[0].Interface()))
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
})
