package cfutil_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCfutil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cfutil Suite")
}
