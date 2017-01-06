/*
 * Copyright 2016-2017 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package eureka_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/eureka"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient/httpclientfakes"
)

var _ = Describe("Service Registry Info", func() {
	var (
		fakeCliConnection *pluginfakes.FakeCliConnection
		fakeClient        *httpclientfakes.FakeClient
		fakeAuthClient    *httpclientfakes.FakeAuthenticatedClient
		fakeResolver      func(dashboardUrl string, accessToken string, authClient httpclient.AuthenticatedClient) (string, error)
		getServiceModel   plugin_models.GetService_Model
		output            string
		err               error
	)

	BeforeEach(func() {
		fakeCliConnection = &pluginfakes.FakeCliConnection{}
		fakeClient = &httpclientfakes.FakeClient{}
		fakeAuthClient = &httpclientfakes.FakeAuthenticatedClient{}
		fakeResolver = func(dashboardUrl string, accessToken string, authClient httpclient.AuthenticatedClient) (string, error) {
			return "https://eureka-dashboard-url/", nil
		}
	})

	JustBeforeEach(func() {
		output, err = eureka.InfoWithResolver(fakeCliConnection, fakeClient, "some-service-registry", fakeAuthClient, fakeResolver)
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

	Context("when the service is found", func() {
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

			Context("but the eureka dashboard URL cannot be resolved", func() {
				BeforeEach(func() {
					fakeResolver = func(dashboardUrl string, accessToken string, authClient httpclient.AuthenticatedClient) (string, error) {
						return "", errors.New("resolution error")
					}
				})

				It("should return a suitable error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("Error obtaining service registry dashboard URL: resolution error"))
				})
			})

			Context("and the eureka dashboard URL can be resolved", func() {
				Context("but eureka cannot be contacted", func() {
					BeforeEach(func() {
						fakeClient.DoReturns(nil, errors.New("some error"))
					})

					It("should return a suitable error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("Service registry error: some error"))
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
							resp.Body = ioutil.NopCloser(strings.NewReader(`{"nodeCount":"1","peers":[{"uri":"uri1","issuer":"issuer1","skipSslValidation":true},{"uri":"uri2","issuer":"issuer2","skipSslValidation":false}]}`))
							fakeClient.DoReturns(resp, nil)
						})

						It("should have obtained an access token", func() {
							Expect(accessTokenCallCount).To(Equal(1))
						})

						It("should have sent a request to the correct URL", func() {
							Expect(fakeClient.DoCallCount()).To(Equal(1))
							req := fakeClient.DoArgsForCall(0)
							Expect(req.URL.String()).To(Equal("https://eureka-dashboard-url/info"))
						})

						It("should have sent a request with the correct accept header", func() {
							Expect(fakeClient.DoCallCount()).To(Equal(1))
							req := fakeClient.DoArgsForCall(0)
							Expect(req.Header.Get("Accept")).To(Equal("application/json"))
						})

						It("should not return an error", func() {
							Expect(err).NotTo(HaveOccurred())
						})

						It("should return the service instance name", func() {
							Expect(output).To(ContainSubstring("Service instance: some-service-registry\n"))
						})

						It("should return the eureka server URL", func() {
							Expect(output).To(ContainSubstring("Server URL: https://eureka-dashboard-url/\n"))
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
	})
})
