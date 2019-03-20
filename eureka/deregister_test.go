package eureka_test

import (
	"errors"

	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient/httpclientfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/eureka"
)

var _ = Describe("Deregister", func() {

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
		err = eureka.Deregister(fakeAuthClient, eurekaUrl, eurekaAppName, instanceId, testAccessToken)
	})

	It("should issue a DELETE with the correct parameters", func() {
		Expect(fakeAuthClient.DoAuthenticatedDeleteCallCount()).To(Equal(1))
		url, accessToken := fakeAuthClient.DoAuthenticatedDeleteArgsForCall(0)
		Expect(url).To(Equal("http://some.host/x/y/cli/instances/someguideureka/apps/eureakappname/instanceid"))
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
