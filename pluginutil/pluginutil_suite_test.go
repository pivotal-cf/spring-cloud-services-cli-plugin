package pluginutil_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPluginutil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pluginutil Suite")
}
