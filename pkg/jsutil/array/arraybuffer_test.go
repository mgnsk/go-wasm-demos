// +build js,wasm

package array_test

import (
	"github.com/mgnsk/jsutil/array"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ArrayBuffer", func() {
	var a array.Buffer

	When("Array buffer is created", func() {
		BeforeEach(func() {
			a = array.NewBuffer(129)
		})

		It("has correct size", func() {
			Expect(a.JSValue().Get("byteLength").Int()).To(Equal(129))
		})
	})
})
