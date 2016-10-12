package eureka_test

import (
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/eureka"

	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"code.cloudfoundry.org/cli/plugin/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/cloudfoundry/cli/cf/errors"
)

var _ = Describe("Dump", func() {

	var fakeCliConnection *pluginfakes.FakeCliConnection
	var getAppModel plugin_models.GetAppModel

	BeforeEach(func() {
		fakeCliConnection = &pluginfakes.FakeCliConnection{}
	})

	Context("When the application is not found", func() {
		BeforeEach(func() {
			fakeCliConnection.GetAppReturns(getAppModel, errors.New("some error"))
		})

		It("should return a suitable error", func() {
			_, err := eureka.Dump(fakeCliConnection, "someApp")
			Expect(fakeCliConnection.GetAppArgsForCall(0)).To(Equal("someApp"))
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("Application not found: some error"))
		})
	})

	Context("When the application is found", func() {
		BeforeEach(func() {
			getAppModel.Guid = "test-guid"
			fakeCliConnection.GetAppReturns(getAppModel, nil)
		})

		Context("When the application environment cannot be obtained", func() {
			BeforeEach(func() {
				fakeCliConnection.CliCommandWithoutTerminalOutputReturns([]string{}, errors.New("some error"))
			})

			It("should return a suitable error", func() {
				_, err := eureka.Dump(fakeCliConnection, "someApp")
				Expect(fakeCliConnection.CliCommandWithoutTerminalOutputArgsForCall(0)[0]).To(Equal("curl"))
				Expect(fakeCliConnection.CliCommandWithoutTerminalOutputArgsForCall(0)[1]).To(Equal("/v2/apps/test-guid/env"))
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("Application environment error: some error"))
			})
		})

		// etc. etc.

	})
})
