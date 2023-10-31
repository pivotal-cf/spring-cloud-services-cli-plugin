package eureka_test

import (
	"errors"

	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/eureka"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient/httpclientfakes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Enable", func() {
	const testAccessToken = "someaccesstoken"

	var (
		fakeAuthClient *httpclientfakes.FakeAuthenticatedClient
		eurekaUrl      string
		eurekaAppName  string
		instanceId     string

		err error
	)

	BeforeEach(func() {
		fakeAuthClient = &httpclientfakes.FakeAuthenticatedClient{}
		eurekaUrl = "https://some.host/x/y/cli/instances/someguid"
		eurekaAppName = "eureakappname"
		instanceId = "instanceid"

	})

	JustBeforeEach(func() {
		err = eureka.Enable(fakeAuthClient, eurekaUrl, eurekaAppName, instanceId, testAccessToken)
	})

	It("should issue a DELETE with the correct parameters", func() {
		Expect(fakeAuthClient.DoAuthenticatedDeleteCallCount()).To(Equal(1))
		url, accessToken := fakeAuthClient.DoAuthenticatedDeleteArgsForCall(0)
		Expect(url).To(Equal("https://some.host/x/y/cli/instances/someguideureka/apps/eureakappname/instanceid/status?value=UP"))
		Expect(accessToken).To(Equal(testAccessToken))
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when DELETE return an error", func() {
		var testError error

		BeforeEach(func() {
			testError = errors.New("failure is not an option")
			fakeAuthClient.DoAuthenticatedDeleteReturns(99, testError)
		})

		It("should return the error", func() {
			Expect(err).To(Equal(testError))
		})
	})
})
