package cfutil_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCfutil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cfutil Suite")
}
