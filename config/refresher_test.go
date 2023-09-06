package config_test

import (
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/config"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient/httpclientfakes"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/serviceutil/serviceutilfakes"
	"io"
	"strings"
)

var _ = Describe("Refresher", func() {

	const (
		accessToken       = "fake-access-token"
		bearerAccessToken = "bearer fake-access-token"
		serviceURI        = "service-uri/"
		refreshURI        = "service-uri/actuator/refreshmirrors"
		configServerName  = "fake-config-server-name"
	)

	var (
		fakeCliConnection *pluginfakes.FakeCliConnection
		refresher         config.Refresher
		fakeAuthClient    *httpclientfakes.FakeAuthenticatedClient
		fakeResolver      *serviceutilfakes.FakeServiceInstanceResolver
		refreshError      error
	)

	BeforeEach(func() {
		fakeCliConnection = &pluginfakes.FakeCliConnection{}
		fakeAuthClient = &httpclientfakes.FakeAuthenticatedClient{}
		fakeResolver = &serviceutilfakes.FakeServiceInstanceResolver{}

		fakeCliConnection.AccessTokenReturns(bearerAccessToken, nil)
		fakeResolver.GetServiceInstanceUrlReturns(serviceURI, nil)
	})

	JustBeforeEach(func() {
		refresher = config.NewRefresher(fakeCliConnection, fakeAuthClient, fakeResolver)
	})

	Describe("Refresh", func() {

		JustBeforeEach(func() {
			refreshError = refresher.Refresh(configServerName)
		})

		Context("when refresh endpoint returns success", func() {
			body := `{
				"rep-1-url" : { "commitId" : "11", "status": "SUCCESS"},
				"rep-2-url" : { "commitId" : "21", "status": "SUCCESS"}
			}`

			BeforeEach(func() {
				fakeAuthClient.DoAuthenticatedPostReturns(io.NopCloser(strings.NewReader(body)), 200, nil)
			})

			It("should call refresh endpoint", func() {
				Expect(refreshError).To(BeNil())
				Expect(fakeAuthClient.DoAuthenticatedPostCallCount()).Should(Equal(1))

				url, bodyType, _, token := fakeAuthClient.DoAuthenticatedPostArgsForCall(0)
				Expect(url).To(Equal(refreshURI))
				Expect(token).To(Equal(accessToken))
				Expect(bodyType).To(Equal("application/json"))
			})
		})

		Context("when refresh endpoint fails with error", func() {

			var e = errors.New("failed to refresh mirror with error")

			BeforeEach(func() {
				fakeAuthClient.DoAuthenticatedPostReturns(io.NopCloser(strings.NewReader("")), 500, e)
				refreshError = refresher.Refresh("fake-service-name")
			})

			It("should return the provided error message", func() {
				Expect(refreshError).To(Equal(e))
			})
		})

		Context("when refresh fails without error message", func() {

			BeforeEach(func() {
				fakeAuthClient.DoAuthenticatedPostReturns(io.NopCloser(strings.NewReader("")), 500, nil)
				refreshError = refresher.Refresh("fake-service-name")
			})

			It("should return the default error message", func() {
				Expect(refreshError).To(Equal(errors.New("failed to refresh mirror")))
			})
		})

		Context("when refresh fails for at least one mirror", func() {
			body := `{
				"rep-1-url" : { "commitId" : "11", "status": "SUCCESS"},
				"rep-2-url" : { "commitId" : "21", "status": "FAILED"},
				"rep-3-url" : { "commitId" : "31", "status": "SUCCESS"}
			}`
			BeforeEach(func() {
				fakeAuthClient.DoAuthenticatedPostReturns(io.NopCloser(strings.NewReader(body)), 200, nil)
				refreshError = refresher.Refresh("fake-service-name")
			})

			It("should return the error message", func() {
				Expect(refreshError).To(Equal(errors.New("failed to refresh mirror")))
			})
		})
	})
})
