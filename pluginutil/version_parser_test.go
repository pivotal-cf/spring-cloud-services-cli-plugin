package pluginutil_test

import (
	"fmt"

	"code.cloudfoundry.org/cli/plugin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/pluginutil"
)

var _ = Describe("ParsePluginVersion", func() {

	var (
		pluginVersion string
		fail          func(format string, inserts ...interface{})
		failed        bool
		firstFailure  string
		parsedVersion plugin.VersionType
	)

	BeforeEach(func() {
		failed = false
		firstFailure = ""
		fail = func(format string, inserts ...interface{}) {
			// Capture just the first failure because ParsePluginVersion is designed to take a function that does not return normally
			if !failed {
				firstFailure = fmt.Sprintf(format, inserts...)
			}
			failed = true

		}
	})

	JustBeforeEach(func() {
		parsedVersion = pluginutil.ParsePluginVersion(pluginVersion, fail)
	})

	Context("when the input version has three integer components", func() {
		BeforeEach(func() {
			pluginVersion = "5.4.3"
		})

		It("should not fail", func() {
			Expect(failed).To(BeFalse())
		})

		It("should parse the version into its components", func() {
			Expect(parsedVersion).To(Equal(plugin.VersionType{
				Major: 5,
				Minor: 4,
				Build: 3,
			}))
		})
	})

	Context("when the input version has the wrong number of components", func() {
		BeforeEach(func() {
			pluginVersion = "2.0"
		})

		It("should fail", func() {
			Expect(failed).To(BeTrue())
		})

		It("should provide a suitable message", func() {
			Expect(firstFailure).To(Equal(`pluginVersion "2.0" has invalid format. Expected 3 dot-separated integer components.`))
		})
	})

	Context("when the input version has a non-integer component", func() {
		BeforeEach(func() {
			pluginVersion = "2.0."
		})

		It("should fail", func() {
			Expect(failed).To(BeTrue())
		})

		It("should provide a suitable message", func() {
			Expect(firstFailure).To(Equal(`pluginVersion "2.0." has invalid format. Expected integer components.`))
		})
	})

})
