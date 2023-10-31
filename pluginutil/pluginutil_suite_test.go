package pluginutil_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPluginutil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pluginutil Suite")
}
