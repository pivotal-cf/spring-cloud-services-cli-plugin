package httpclient_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestHttpclient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Httpclient Suite")
}
