package jsutil_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestJSUtil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "jsutil")
}
