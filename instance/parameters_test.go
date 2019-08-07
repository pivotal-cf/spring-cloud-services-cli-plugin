package instance_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/serviceutil"

	"bytes"
	"errors"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient/httpclientfakes"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/instance"
	"io/ioutil"
	"net/http"
)

var _ = Describe("Parameters", func() {

	const (
		testAccessToken    = "someaccesstoken"
		parametersResponse = "parameters response"
	)

	var (
		fakeAuthClient                 *httpclientfakes.FakeAuthenticatedClient
		serviceInstanceAdminParameters serviceutil.ManagementParameters
		output                         string
		err                            error
	)

	BeforeEach(func() {
		fakeAuthClient = &httpclientfakes.FakeAuthenticatedClient{}
		serviceInstanceAdminParameters = serviceutil.ManagementParameters{
			Url: "https://servicebroker.host/cli/instances/si-guid",
		}
	})

	JustBeforeEach(func() {
		output, err = instance.NewParametersOperation(fakeAuthClient).Run(serviceInstanceAdminParameters, testAccessToken)
	})

	It("should issue a GET to the service broker parameters endpoint with the correct parameters", func() {
		Expect(fakeAuthClient.DoAuthenticatedGetCallCount()).To(Equal(1))
		url, accessToken := fakeAuthClient.DoAuthenticatedGetArgsForCall(0)
		Expect(url).To(Equal("https://servicebroker.host/cli/instances/si-guid/parameters"))
		Expect(accessToken).To(Equal(testAccessToken))
	})

	Context("when GET to service broker returns an error", func() {
		var testError error

		BeforeEach(func() {
			testError = errors.New("it's all gone a bit Pete Tong")
			fakeAuthClient.DoAuthenticatedGetReturns(nil, 0, testError)
		})

		It("should return the error", func() {
			Expect(output).To(Equal(""))
			Expect(err).To(Equal(testError))
		})
	})

	Context("when GET to service broker returns a bad status code", func() {
		BeforeEach(func() {
			fakeAuthClient.DoAuthenticatedGetReturns(nil, http.StatusNotFound, nil)
		})

		It("should return the error", func() {
			Expect(err).To(MatchError("Service broker view instance configuration failed: 404"))
		})
	})

	Context("when GET to service broker returns no response reader", func() {
		BeforeEach(func() {
			fakeAuthClient.DoAuthenticatedGetReturns(nil, http.StatusOK, nil)
		})

		It("should return a suitable error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("Service broker view instance configuration response body missing"))
		})
	})

	Context("when GET to service broker returns response reader that cannot be read", func() {
		BeforeEach(func() {
			fakeAuthClient.DoAuthenticatedGetReturns(ioutil.NopCloser(badReader{}), http.StatusOK, nil)
		})

		It("should return a suitable error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("Cannot read service instance configuration response body: read error"))
		})
	})

	Context("when GET to service broker returns valid response", func() {
		BeforeEach(func() {
			fakeAuthClient.DoAuthenticatedGetReturns(ioutil.NopCloser(bytes.NewBufferString(parametersResponse)), http.StatusOK, nil)
		})

		It("should return the output from the service broker parameters endpoint", func() {
			Expect(output).To(Equal(parametersResponse))
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
