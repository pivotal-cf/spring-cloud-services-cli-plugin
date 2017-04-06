package serviceutil_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestServiceutil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Serviceutil Suite")
}
