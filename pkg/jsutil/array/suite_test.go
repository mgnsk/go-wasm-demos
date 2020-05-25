// +build js,wasm

package array_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestArray(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "array")
}
