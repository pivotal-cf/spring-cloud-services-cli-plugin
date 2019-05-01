/*
 * Copyright (C) 2017-Present Pivotal Software, Inc. All rights reserved.
 *
 * This program and the accompanying materials are made available under
 * the terms of the under the Apache License, Version 2.0 (the "License”);
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package serviceutil_test

import (
	"fmt"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/serviceutil"

	"bytes"
	"errors"
	"io/ioutil"

	"code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient/httpclientfakes"
)

var _ = Describe("ServiceInstanceUrlResolver", func() {

	const (
		errMessage              = "some error"
		testAccessToken         = "someaccesstoken"
		testServiceInstanceName = "service-instance-name"
		testGuid                = "1558aa56-97a5-47cd-a73c-7d5cd0818b22"
		scs2DashboardUrl        = "https://spring-cloud-broker.example.com/dashboard/p-config-server/1558aa56-97a5-47cd-a73c-7d5cd0818b22"
		scs3DashboardUrl        = "https://config-server-1558aa56-97a5-47cd-a73c-7d5cd0818b22.example.com/dashboard"
	)

	var (
		fakeCliConnection *pluginfakes.FakeCliConnection
		accessToken       string
		fakeAuthClient    *httpclientfakes.FakeAuthenticatedClient
		resolvedUrl       string
		err               error
		testError         error
		resolver          serviceutil.ServiceInstanceUrlResolver
	)

	BeforeEach(func() {
		fakeCliConnection = &pluginfakes.FakeCliConnection{}
		fakeAuthClient = &httpclientfakes.FakeAuthenticatedClient{}
		testError = errors.New(errMessage)

		fakeCliConnection.ApiEndpointReturns("https://api.example.com", nil)

		resolver = serviceutil.NewServiceInstanceUrlResolver(fakeCliConnection, fakeAuthClient)
	})

	Describe("GetServiceInstanceUrl", func() {
		JustBeforeEach(func() {
			resolvedUrl, err = resolver.GetServiceInstanceUrl(testServiceInstanceName, accessToken)
		})

		It("should get the service", func() {
			Expect(fakeCliConnection.GetServiceCallCount()).To(Equal(1))
			Expect(fakeCliConnection.GetServiceArgsForCall(0)).To(Equal(testServiceInstanceName))
		})

		Context("when the service instance is not found", func() {
			BeforeEach(func() {
				fakeCliConnection.GetServiceReturns(plugin_models.GetService_Model{}, testError)
			})

			It("should propagate the error", func() {
				Expect(err).To(MatchError("Service instance not found: " + errMessage))
			})
		})

		Context("when an SCS 2 service instance is found", func() {
			BeforeEach(func() {
				fakeCliConnection.GetServiceReturns(plugin_models.GetService_Model{
					ServiceOffering: plugin_models.GetService_ServiceFields{
						Name: "p-config-server",
					},
					DashboardUrl: "https://spring-cloud-broker.example.com/dashboard/p-config-server/guid",
				}, nil)
			})

			Context("when the dashboard URL is not in the correct format", func() {
				Context("because it is malformed", func() {
					BeforeEach(func() {
						fakeCliConnection.GetServiceReturns(plugin_models.GetService_Model{
							ServiceOffering: plugin_models.GetService_ServiceFields{
								Name: "p-config-server",
							},
							DashboardUrl: "://",
						}, nil)
					})

					It("should return a suitable error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("parse ://: missing protocol scheme"))
					})
				})

				Context("because its path format is invalid", func() {
					BeforeEach(func() {
						fakeCliConnection.GetServiceReturns(plugin_models.GetService_Model{
							ServiceOffering: plugin_models.GetService_ServiceFields{
								Name: "p-config-server",
							},
							DashboardUrl: "https://spring-cloud-broker.example.com",
						}, nil)
					})

					It("should return a suitable error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("path of https://spring-cloud-broker.example.com has no segments"))
					})
				})
			})

			Context("when dashboard cannot be contacted", func() {
				BeforeEach(func() {
					fakeAuthClient.DoAuthenticatedGetReturns(nil, 502, testError)
				})

				It("should return a suitable error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("Invalid service definition response: " + errMessage))
				})
			})

			Context("when dashboard can be contacted", func() {
				Context("but the response body contains invalid JSON", func() {
					BeforeEach(func() {
						fakeAuthClient.DoAuthenticatedGetReturns(ioutil.NopCloser(bytes.NewBufferString("")), 200, nil)
					})

					It("should return a suitable error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("JSON response failed to unmarshal: "))
					})
				})

				Context("but the response body has the wrong content", func() {
					BeforeEach(func() {
						fakeAuthClient.DoAuthenticatedGetReturns(ioutil.NopCloser(bytes.NewBufferString(`{"credentials":0}`)), 200, nil)
					})

					It("should return a suitable error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError(`JSON response failed to unmarshal: {"credentials":0}`))
					})
				})

				Context("but the '/cli/instance endpoint cannot be found", func() {
					BeforeEach(func() {
						fakeAuthClient.DoAuthenticatedGetReturns(ioutil.NopCloser(bytes.NewBufferString(`{"credentials":0}`)), 404, nil)
					})
					It("should warn that the SCS version might be too old", func() {
						versionWarning := "The /cli/instance endpoint could not be found.\n" +
							"This could be because the Spring Cloud Services broker version is too old.\n" +
							"Please ensure SCS is at least version 1.3.3.\n"

						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError(versionWarning))
					})
				})

				Context("and the response body cannot be read", func() {
					BeforeEach(func() {
						fakeAuthClient.DoAuthenticatedGetReturns(&badReader{readErr: testError}, 200, nil)
					})

					It("should return a suitable error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("Cannot read service definition response body: " + errMessage))
					})
				})

				Context("and the response body has an empty URI", func() {
					BeforeEach(func() {
						fakeAuthClient.DoAuthenticatedGetReturns(ioutil.NopCloser(bytes.NewBufferString(`{"credentials":{"uri":""}}`)), 200, nil)
					})

					It("should return a suitable error", func() {
						serviceDefinitionError := `JSON response contained empty property 'credentials.url', response body: '{"credentials":{"uri":""}}'`

						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError(serviceDefinitionError))
					})
				})

				Context("and the response body has the correct content", func() {
					BeforeEach(func() {
						fakeAuthClient.DoAuthenticatedGetReturns(ioutil.NopCloser(bytes.NewBufferString(`{"credentials":{"uri":"https://service.instance.url"}}`)), 200, nil)
					})

					It("should return the service instance url", func() {
						Expect(err).NotTo(HaveOccurred())
						Expect(resolvedUrl).To(Equal("https://service.instance.url/"))
					})
				})
			})
		})

		Context("when an SCS 3 service instance is found", func() {
			BeforeEach(func() {
				fakeCliConnection.GetServiceReturns(plugin_models.GetService_Model{
					ServiceOffering: plugin_models.GetService_ServiceFields{
						Name: "p.config-server",
					},
					DashboardUrl: "https://config-server-guid.example.com/dashboard",
				}, nil)
			})

			Context("when the dashboard URL is not in the correct format", func() {
				Context("because it is malformed", func() {
					BeforeEach(func() {
						fakeCliConnection.GetServiceReturns(plugin_models.GetService_Model{
							ServiceOffering: plugin_models.GetService_ServiceFields{
								Name: "p.config-server",
							},
							DashboardUrl: "://",
						}, nil)
					})

					It("should return a suitable error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("parse ://: missing protocol scheme"))
					})
				})

				It("should return the service instance url", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(resolvedUrl).To(Equal("https://config-server-guid.example.com/"))
				})
			})
		})
	})

	Describe("GetManagementEndpoint", func() {
		var (
			isLifecycleOperation bool
		)

		BeforeEach(func() {
			isLifecycleOperation = true
		})

		JustBeforeEach(func() {
			resolvedUrl, err = resolver.GetManagementUrl(
				testServiceInstanceName,
				testAccessToken,
				isLifecycleOperation)
		})

		It("should get the service", func() {
			Expect(fakeCliConnection.GetServiceCallCount()).To(Equal(1))
			Expect(fakeCliConnection.GetServiceArgsForCall(0)).To(Equal(testServiceInstanceName))
		})

		Context("when the service instance is not found", func() {
			BeforeEach(func() {
				fakeCliConnection.GetServiceReturns(plugin_models.GetService_Model{}, testError)
			})

			It("propagates the error", func() {
				Expect(err).To(MatchError("Service instance not found: " + errMessage))
			})
		})

		Context("when an SCS 2 service instance is found", func() {
			BeforeEach(func() {
				fakeCliConnection.GetServiceReturns(plugin_models.GetService_Model{
					ServiceOffering: plugin_models.GetService_ServiceFields{
						Name: "p-config-server",
					},
					DashboardUrl: scs2DashboardUrl,
					Guid:         testGuid,
				}, nil)
			})

			Context("when it's a lifecycle operation", func() {
				BeforeEach(func() {
					isLifecycleOperation = true
				})

				It("returns the cli endpoint on the broker", func() {
					Expect(resolvedUrl).To(Equal(fmt.Sprintf("https://spring-cloud-broker.example.com/cli/instances/%s", testGuid)))
				})
			})

			Context("when it's not a lifecycle operation", func() {
				BeforeEach(func() {
					isLifecycleOperation = false
				})

				It("returns the cli endpoint on the broker", func() {
					Expect(resolvedUrl).To(Equal(fmt.Sprintf("https://spring-cloud-broker.example.com/cli/instances/%s", testGuid)))
				})
			})
		})

		Context("when an SCS 3 service instance is found", func() {
			BeforeEach(func() {
				fakeCliConnection.GetServiceReturns(plugin_models.GetService_Model{
					ServiceOffering: plugin_models.GetService_ServiceFields{
						Name: "p.config-server",
					},
					DashboardUrl: scs3DashboardUrl,
					Guid:         testGuid,
				}, nil)
			})

			Context("when it's a lifecycle operation", func() {
				BeforeEach(func() {
					isLifecycleOperation = true
				})

				It("returns the cli endpoint on the broker", func() {
					Expect(resolvedUrl).To(Equal(fmt.Sprintf("https://scs-service-broker.example.com/cli/instances/%s", testGuid)))
				})
			})

			Context("when it's not a lifecycle operation", func() {
				BeforeEach(func() {
					isLifecycleOperation = false
				})

				It("returns the cli endpoint on the backing app", func() {
					Expect(resolvedUrl).To(Equal(fmt.Sprintf("https://config-server-%[1]s.example.com/cli/instances/%[1]s", testGuid)))
				})
			})
		})
	})
})

type badReader struct {
	readErr error
}

func (br *badReader) Read(p []byte) (n int, err error) {
	return 0, br.readErr
}

func (*badReader) Close() error {
	return nil
}
