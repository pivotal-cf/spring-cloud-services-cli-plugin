package eureka_test

import (
	"code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/eureka"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/format"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient/httpclientfakes"
	"io/ioutil"
	"net/http"
	"strings"
)

var _ = Describe("Service Registry List", func() {
	var (
		fakeCliConnection *pluginfakes.FakeCliConnection
		fakeClient        *httpclientfakes.FakeClient
		getServiceModel   plugin_models.GetService_Model
		output            string
		err               error
	)

	BeforeEach(func() {
		fakeCliConnection = &pluginfakes.FakeCliConnection{}
		fakeClient = &httpclientfakes.FakeClient{}
	})

	JustBeforeEach(func() {
		output, err = eureka.List(fakeCliConnection, fakeClient, "some-service-registry")
	})

	Context("when the service is not found", func() {
		BeforeEach(func() {
			fakeCliConnection.GetServiceReturns(getServiceModel, errors.New("some error"))
		})

		It("should return a suitable error", func() {
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

		Context("but the access token is not available", func() {
			var accessTokenCallCount int

			BeforeEach(func() {
				accessTokenCallCount = 0
				fakeCliConnection.AccessTokenStub = func() (string, error) {
					accessTokenCallCount++
					return "", errors.New("some access token error")
				}
			})

			It("should return a suitable error", func() {
				Expect(accessTokenCallCount).To(Equal(1))
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("Access token not available: some access token error"))
			})
		})

		Context("and the access token is available", func() {
			var accessTokenCallCount int

			BeforeEach(func() {
				accessTokenCallCount = 0
				fakeCliConnection.AccessTokenStub = func() (string, error) {
					accessTokenCallCount++
					return "bearer someaccesstoken", nil
				}
			})

			Context("but eureka cannot be contacted", func() {
				BeforeEach(func() {
					fakeClient.DoReturns(nil, errors.New("some error"))
				})

				It("should return a suitable error", func() {
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
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(HavePrefix("Invalid service registry response JSON: "))
					})
				})

				Context("and the response is valid", func() {
					BeforeEach(func() {
						resp := &http.Response{}
						resp.Body = ioutil.NopCloser(strings.NewReader(`
{
   "applications":{
      "application":[
         {
            "instance":[
               {
                  "app":"APP-1",
                  "status":"UP",
                  "metadata":{
                     "zone":"zone1"
                  }
               }
            ]
         },
         {
            "instance":[
               {
                  "app":"APP-2",
                  "status":"OUT_OF_SERVICE",
                  "metadata":{
                     "zone":"zone2"
                  }
               }
            ]
         }
      ]
   }
}`))
						fakeClient.DoReturns(resp, nil)
					})

					It("should have obtained an access token", func() {
						Expect(accessTokenCallCount).To(Equal(1))
					})

					It("should have sent a request to the correct URL", func() {
						Expect(fakeClient.DoCallCount()).To(Equal(1))
						req := fakeClient.DoArgsForCall(0)
						Expect(req.URL.String()).To(Equal("https://eureka-some-guid.some.host.name/eureka/apps"))
					})

					It("should have sent a request with the correct accept header", func() {
						Expect(fakeClient.DoCallCount()).To(Equal(1))
						req := fakeClient.DoArgsForCall(0)
						Expect(req.Header.Get("Accept")).To(Equal("application/json"))
					})

					It("should have sent a request with the correct authorization header", func() {
						Expect(fakeClient.DoCallCount()).To(Equal(1))
						req := fakeClient.DoArgsForCall(0)
						Expect(req.Header.Get("Authorization")).To(Equal("bearer someaccesstoken"))
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

					It("should return the registered applications", func() {
						tab := &format.Table{}
						tab.Entitle([]string{"eureka app name", "zone", "status"})
						tab.AddRow([]string{"APP-1", "zone1", "UP"})
						tab.AddRow([]string{"APP-2", "zone2", "OUT_OF_SERVICE"})
						Expect(output).To(ContainSubstring(tab.String()))
					})
				})
			})
		})
	})
})
