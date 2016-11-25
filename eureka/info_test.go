package eureka_test

import (
	"code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/eureka"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient/httpclientfakes"
	"io/ioutil"
	"net/http"
	"strings"
)

var _ = Describe("Service Registry Info", func() {
	var fakeCliConnection *pluginfakes.FakeCliConnection
	var fakeClient *httpclientfakes.FakeClient
	var getServiceModel plugin_models.GetService_Model

	BeforeEach(func() {
		fakeCliConnection = &pluginfakes.FakeCliConnection{}
		fakeClient = &httpclientfakes.FakeClient{}
	})

	Context("when the service is not found", func() {
		BeforeEach(func() {
			fakeCliConnection.GetServiceReturns(getServiceModel, errors.New("some error"))
		})

		It("should return a suitable error", func() {
			_, err := eureka.Info(fakeCliConnection, fakeClient, "some-service-registry")
			Expect(fakeCliConnection.GetServiceCallCount()).To(Equal(1))
			Expect(fakeCliConnection.GetServiceArgsForCall(0)).To(Equal("some-service-registry"))
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("Service registry instance not found: some error"))
		})
	})

	Context("when the dashboard URL is not in the correct format", func() {
		Context("and the dashboard URL is malformed", func() {
			BeforeEach(func() {
				getServiceModel.DashboardUrl = "://"
				fakeCliConnection.GetServiceReturns(getServiceModel, nil)
			})

			It("should return a suitable error", func() {
				_, err := eureka.Info(fakeCliConnection, fakeClient, "some-service-registry")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("Invalid service registry dashboard URL: parse ://: missing protocol scheme"))
			})
		})

		Context("and the hostname format is invalid", func() {
			BeforeEach(func() {
				getServiceModel.DashboardUrl = "https://a"
				fakeCliConnection.GetServiceReturns(getServiceModel, nil)
			})

			It("should return a suitable error", func() {
				_, err := eureka.Info(fakeCliConnection, fakeClient, "some-service-registry")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("Invalid service registry dashboard URL: hostname of https://a has less than two labels"))
			})
		})

		Context("and the path format is invalid", func() {
			BeforeEach(func() {
				getServiceModel.DashboardUrl = "https://spring-cloud-broker.some.host.name"
				fakeCliConnection.GetServiceReturns(getServiceModel, nil)
			})

			It("should return a suitable error", func() {
				_, err := eureka.Info(fakeCliConnection, fakeClient, "some-service-registry")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("Invalid service registry dashboard URL: path of https://spring-cloud-broker.some.host.name has no segments"))
			})
		})
	})

	Context("when the dashboard URL is in the correct format", func() {
		BeforeEach(func() {
			getServiceModel.DashboardUrl = "https://spring-cloud-broker.some.host.name/x/y/z/some-guid"
			fakeCliConnection.GetServiceReturns(getServiceModel, nil)
		})

		It("should send a request to the correct URL", func() {
			fakeClient.DoReturns(nil, errors.New("some error")) // easier than providing a valid response
			eureka.Info(fakeCliConnection, fakeClient, "some-service-registry")
			Expect(fakeClient.DoCallCount()).To(Equal(1))
			req := fakeClient.DoArgsForCall(0)
			Expect(req.URL.String()).To(Equal("https://eureka-some-guid.some.host.name/info"))

		})

		It("should send a request with the correct accept header", func() {
			fakeClient.DoReturns(nil, errors.New("some error")) // easier than providing a valid response
			eureka.Info(fakeCliConnection, fakeClient, "some-service-registry")
			Expect(fakeClient.DoCallCount()).To(Equal(1))
			req := fakeClient.DoArgsForCall(0)
			Expect(req.Header.Get("Accept")).To(Equal("application/json"))

		})

		Context("but eureka cannot be contacted", func() {
			BeforeEach(func() {
				fakeClient.DoReturns(nil, errors.New("some error"))
			})

			It("should return a suitable error", func() {
				_, err := eureka.Info(fakeCliConnection, fakeClient, "some-service-registry")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("Service registry unavailable: some error"))
			})
		})

		Context("and eureka responds", func() {
			Context("but the response body is missing", func() {
				BeforeEach(func() {
					resp := &http.Response{}
					fakeClient.DoReturns(resp, nil)
				})

				It("should return a suitable error", func() {
					_, err := eureka.Info(fakeCliConnection, fakeClient, "some-service-registry")
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("Invalid service registry response: missing body"))
				})
			})

			Context("but the response body contains invalid JSON", func() {
				BeforeEach(func() {
					resp := &http.Response{}
					resp.Body = ioutil.NopCloser(strings.NewReader(`{`))
					fakeClient.DoReturns(resp, nil)
				})

				It("should return a suitable error", func() {
					_, err := eureka.Info(fakeCliConnection, fakeClient, "some-service-registry")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(HavePrefix("Invalid service registry response JSON: "))
				})
			})

			Context("and the response is valid", func() {
				var (
					output string
					err    error
				)

				BeforeEach(func() {
					resp := &http.Response{}
					resp.Body = ioutil.NopCloser(strings.NewReader(`{"nodeCount":"1","peers":[{"uri":"uri1","issuer":"issuer1","skipSslValidation":true},{"uri":"uri2","issuer":"issuer2","skipSslValidation":false}]}`))
					fakeClient.DoReturns(resp, nil)
				})

				JustBeforeEach(func() {
					output, err = eureka.Info(fakeCliConnection, fakeClient, "some-service-registry")
				})

				It("should not return an error", func() {
					Expect(err).NotTo(HaveOccurred())
				})

				It("should return the service instance name", func() {
					Expect(output).To(ContainSubstring("Service instance: some-service-registry\n"))
				})

				It("should return the eureka server URL", func() {
					Expect(output).To(ContainSubstring("Server URL: https://eureka-some-guid.some.host.name/\n"))
				})

				It("should return the high availability count", func() {
					Expect(output).To(ContainSubstring("High availability count: 1\n"))
				})

				It("should return the peers", func() {
					Expect(output).To(ContainSubstring("Peers: uri1, uri2\n"))
				})
			})
		})

	})
})
