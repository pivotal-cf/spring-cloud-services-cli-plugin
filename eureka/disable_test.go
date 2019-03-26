package eureka_test

import (
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/eureka"

	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient/httpclientfakes"
)

var _ = Describe("Disable", func() {
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
		err = eureka.Disable(fakeAuthClient, eurekaUrl, eurekaAppName, instanceId, testAccessToken)
	})

	It("should issue a PUT with the correct parameters", func() {
		Expect(fakeAuthClient.DoAuthenticatedPutCallCount()).To(Equal(1))
		url, accessToken := fakeAuthClient.DoAuthenticatedPutArgsForCall(0)
		Expect(url).To(Equal("http://some.host/x/y/cli/instances/someguideureka/apps/eureakappname/instanceid/status?value=OUT_OF_SERVICE"))
		Expect(accessToken).To(Equal(testAccessToken))
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when PUT return an error", func() {
		var testError error

		BeforeEach(func() {
			testError = errors.New("failure is not an option")
			fakeAuthClient.DoAuthenticatedPutReturns(99, testError)
		})

		It("should return the error", func() {
			Expect(err).To(Equal(testError))
		})
	})
})
