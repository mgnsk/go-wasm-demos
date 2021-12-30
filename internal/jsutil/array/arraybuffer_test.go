//go:build js && wasm
// +build js,wasm

package array_test

import (
	"github.com/mgnsk/go-wasm-demos/internal/jsutil/array"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ArrayBuffer", func() {
	var a array.ArrayBuffer

	When("Array buffer is created", func() {
		BeforeEach(func() {
			a = array.NewArrayBuffer(129)
		})

		It("has correct size", func() {
			Expect(a.Len()).To(Equal(129))
		})

		It("holds correct data", func() {
			data := []byte("Hello world!")
			ab := array.NewArrayBufferFromSlice(data)
			Expect(ab.Bytes()).To(Equal(data))
		})
	})
})
