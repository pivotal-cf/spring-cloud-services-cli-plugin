package config_test

import (
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/config"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient/httpclientfakes"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/serviceutil/serviceutilfakes"
)

var _ = Describe("CredhubSecret", func() {

	const (
		accessToken       = "fake-access-token"
		bearerAccessToken = "bearer fake-access-token"
		serviceURI        = "service-uri/"
		secretsURI        = "service-uri/secrets"
		configServerName  = "fake-config-server-name"
	)

	var (
		fakeCliConnection *pluginfakes.FakeCliConnection
		credhubSecret     config.CredHubSecret
		fakeAuthClient    *httpclientfakes.FakeAuthenticatedClient
		fakeResolver      *serviceutilfakes.FakeServiceInstanceResolver
		secretsError      error
	)

	BeforeEach(func() {
		fakeCliConnection = &pluginfakes.FakeCliConnection{}
		fakeAuthClient = &httpclientfakes.FakeAuthenticatedClient{}
		fakeResolver = &serviceutilfakes.FakeServiceInstanceResolver{}

		fakeCliConnection.AccessTokenReturns(bearerAccessToken, nil)
		fakeResolver.GetServiceInstanceUrlReturns(serviceURI, nil)
	})

	JustBeforeEach(func() {
		credhubSecret = config.NewCredHubSecret(fakeCliConnection, fakeAuthClient, fakeResolver)
	})

	Describe("CredHub Add Secret", func() {

		JustBeforeEach(func() {
			secretsError = credhubSecret.Add(configServerName, "application/cloud/one", "{\"key\":\"secret\"}")
		})

		BeforeEach(func() {
			fakeAuthClient.DoAuthenticatedPutReturns(200, nil)
		})

		It("calls the secrets endpoint", func() {
			Expect(secretsError).To(BeNil())
			Expect(fakeAuthClient.DoAuthenticatedPutCallCount()).Should(Equal(1))

			url, bodyType, body, token := fakeAuthClient.DoAuthenticatedPutArgsForCall(0)
			Expect(url).To(Equal(secretsURI + "/application/cloud/one"))
			Expect(token).To(Equal(accessToken))
			Expect(bodyType).To(Equal("application/json"))
			Expect(body).To(Equal("{\"key\":\"secret\"}"))
		})

		Context("when add fails", func() {

			var e = errors.New("failed to add secret or is not supported by this config server")

			BeforeEach(func() {
				fakeAuthClient.DoAuthenticatedPutReturns(500, e)
				secretsError = credhubSecret.Add("fake-service-name", "application/cloud/one", "{\"key\":\"secret\"}")
			})

			It("should return error message", func() {
				Expect(secretsError).To(Equal(e))
			})
		})

		Context("when add calls an old version of SCS", func() {

			var e = errors.New("failed to add secret or is not supported by this config server")

			BeforeEach(func() {
				fakeAuthClient.DoAuthenticatedPutReturns(403, e)
				secretsError = credhubSecret.Add("old-service-name", "application/cloud/one", "{\"key\":\"secret\"}")
			})

			It("should return error message", func() {
				Expect(secretsError).To(Equal(e))
			})
		})
	})

	Describe("CredHub Remove Secret", func() {

		JustBeforeEach(func() {
			secretsError = credhubSecret.Remove(configServerName, "application/cloud/one")
		})

		BeforeEach(func() {
			fakeAuthClient.DoAuthenticatedDeleteReturns(200, nil)
		})

		It("call the secrets endpoint", func() {
			Expect(secretsError).To(BeNil())
			Expect(fakeAuthClient.DoAuthenticatedDeleteCallCount()).Should(Equal(1))

			url, token := fakeAuthClient.DoAuthenticatedDeleteArgsForCall(0)
			Expect(url).To(Equal(secretsURI + "/application/cloud/one"))
			Expect(token).To(Equal(accessToken))
		})

		Context("when remove fails", func() {

			var e = errors.New("failed to remove secret or is not supported by this config server")

			BeforeEach(func() {
				fakeAuthClient.DoAuthenticatedDeleteReturns(500, e)
				secretsError = credhubSecret.Remove("fake-service-name", "application/cloud/one")
			})

			It("should return error message", func() {
				Expect(secretsError).To(Equal(e))
			})
		})

		Context("when remove calls an old version of SCS", func() {

			var e = errors.New("failed to remove secret or is not supported by this config server")

			BeforeEach(func() {
				fakeAuthClient.DoAuthenticatedDeleteReturns(403, e)
				secretsError = credhubSecret.Remove("old-service-name", "application/cloud/one")
			})

			It("should return error message", func() {
				Expect(secretsError).To(Equal(e))
			})
		})
	})
})
