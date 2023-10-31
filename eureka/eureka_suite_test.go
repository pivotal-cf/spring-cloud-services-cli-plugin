package eureka_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestEureka(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Eureka Suite")
}
