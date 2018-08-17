package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/config"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient/httpclientfakes"
	"errors"
	"code.cloudfoundry.org/cli/plugin/models"
	"net/http"
)

var _ = Describe("Delete", func() {

	const (
		testConfigServerSIName = "my-configserver"
		testGitUri             = "my-testuri"
		testAccessToken        = "someaccesstoken"
		errorText              = "keep calm and carry on"
	)

	var (
		configServerSIName string
		gitURI             string
		fakeCliConnection  *pluginfakes.FakeCliConnection
		fakeAuthClient     *httpclientfakes.FakeAuthenticatedClient
		testError          error
		output             string
		err                error
	)

	BeforeEach(func() {
		fakeAuthClient = &httpclientfakes.FakeAuthenticatedClient{}
		fakeCliConnection = &pluginfakes.FakeCliConnection{}
		testError = errors.New(errorText)
		configServerSIName = testConfigServerSIName
		gitURI = testGitUri
	})

	JustBeforeEach(func() {
		output, err = config.DeleteGitRepo(fakeCliConnection, fakeAuthClient, configServerSIName, gitURI)
	})

	It("should create an access token", func() {
		Expect(fakeCliConnection.AccessTokenCallCount()).To(Equal(1))
	})

	Context("when the access token is not available", func() {
		BeforeEach(func() {
			fakeCliConnection.AccessTokenReturns("", errors.New("error occurred retrieving access token"))
		})

		It("should return a suitable error", func() {
			Expect(fakeCliConnection.AccessTokenCallCount()).To(Equal(1))
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("Access token not available: error occurred retrieving access token"))
		})
	})

	Context("when the access token is available", func() {
		BeforeEach(func() {
			fakeCliConnection.AccessTokenReturns("bearer "+testAccessToken, nil)
		})

		It("should get the config server service instance details", func() {
			Expect(fakeCliConnection.GetServiceCallCount()).To(Equal(1))
			Expect(fakeCliConnection.GetServiceArgsForCall(0)).To(Equal(testConfigServerSIName))
		})

		Context("when the config server service instance is not found", func() {
			BeforeEach(func() {
				fakeCliConnection.GetServiceReturns(plugin_models.GetService_Model{}, testError)
			})

			It("should propagate the error", func() {
				Expect(err).To(MatchError("Config server service instance not found: " + errorText))
			})
		})

		Context("when the config server service instance is successfully found", func() {
			BeforeEach(func() {
				fakeCliConnection.GetServiceReturns(plugin_models.GetService_Model{
					DashboardUrl: "https://spring-cloud-broker.some.host.name/x/y/guid",
				}, nil)
			})

			It("should make an authenticated PATCH call with the expected parameters", func() {
				Expect(fakeAuthClient.DoAuthenticatedPatchCallCount()).To(Equal(1))
				url, bodyType, bodyStr, accessToken := fakeAuthClient.DoAuthenticatedPatchArgsForCall(0)
				Expect(url).To(Equal("https://spring-cloud-broker.some.host.name/cli/configserver/guid"))
				Expect(bodyType).To(Equal("application/json"))
				Expect(bodyStr).To(Equal(`{"operation":"delete","repo":"my-testuri"}`))
				Expect(accessToken).To(Equal(testAccessToken))
			})

			Context("when the authenticated PATCH call returns an error", func() {
				BeforeEach(func() {
					testError = errors.New("oh dear")
					fakeAuthClient.DoAuthenticatedPatchReturns(0, testError)
				})

				It("should return the error", func() {
					Expect(output).To(Equal(""))
					Expect(err).To(MatchError("Unable to delete git repo my-testuri from config server service instance my-configserver: oh dear"))
				})
			})

			Context("when the authenticated PATCH call returns a bad status code", func() {
				BeforeEach(func() {
					fakeAuthClient.DoAuthenticatedPatchReturns(http.StatusNotFound, nil)
				})

				It("should return a suitable error", func() {
					Expect(err).To(MatchError("Unable to delete git repo my-testuri from config server service instance my-configserver: 404"))
				})
			})
			
			Context("when the authenticated PATCH call is successful", func() {
				BeforeEach(func() {
					fakeAuthClient.DoAuthenticatedPatchReturns(http.StatusOK, nil)
				})

				It("should not return an error", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})
})
