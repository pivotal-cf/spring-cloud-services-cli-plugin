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

	"bytes"

	"code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/eureka"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/format"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient/httpclientfakes"
)

var _ = Describe("Service Registry List", func() {
	var (
		fakeCliConnection *pluginfakes.FakeCliConnection
		fakeAuthClient    *httpclientfakes.FakeAuthenticatedClient
		fakeResolver      func(dashboardUrl string, accessToken string, authClient httpclient.AuthenticatedClient) (string, error)
		getServiceModel   plugin_models.GetService_Model
		output            string
		err               error
	)

	BeforeEach(func() {
		fakeCliConnection = &pluginfakes.FakeCliConnection{}
		fakeAuthClient = &httpclientfakes.FakeAuthenticatedClient{}
		fakeAuthClient.DoAuthenticatedGetReturns(httpclient.AuthClientResponse{bytes.NewBufferString("https://fake.com"), 200}, nil)
		fakeResolver = func(dashboardUrl string, accessToken string, authClient httpclient.AuthenticatedClient) (string, error) {
			return "https://eureka-dashboard-url/", nil
		}
	})

	JustBeforeEach(func() {
		output, err = eureka.ListWithResolver(fakeCliConnection, "some-service-registry", fakeAuthClient, fakeResolver)
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
					return "someaccesstoken", nil
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
						fakeAuthClient.DoAuthenticatedGetReturns(httpclient.AuthClientResponse{bytes.NewBufferString(`{"authenticated":true}`), 200}, errors.New("some error"))
					})

					It("should return a suitable error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("Service registry error: some error"))
					})
				})

				Context("and eureka responds", func() {
					Context("but the response body contains invalid JSON", func() {
						BeforeEach(func() {
							fakeAuthClient.DoAuthenticatedGetReturns(httpclient.AuthClientResponse{bytes.NewBufferString(""), 200}, nil)
						})

						It("should return a suitable error", func() {
							Expect(err).To(HaveOccurred())
							Expect(err).To(MatchError("Invalid service registry response JSON: unexpected end of JSON input, response body: ''"))
						})
					})

					Context("and the response is valid", func() {
						BeforeEach(func() {
							fakeAuthClient.DoAuthenticatedGetReturns(httpclient.AuthClientResponse{bytes.NewBufferString(`
{
   "applications":{
      "application":[
         {
            "instance":[
               {
                  "app":"APP-1",
                  "status":"UP",
                  "metadata":{
                     "zone":"zone1",
                     "cfAppGuid":"062bd505-8b19-44ca-4451-4a932932143a",
                     "cfInstanceIndex":"2"
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
                     "zone":"zone2",
                     "cfAppGuid":"162bd505-1b19-14ca-1451-1a9329321431",
                     "cfInstanceIndex":"3"
                  }
               }
            ]
         }
      ]
   }
}`), 200}, nil)
						})

						Context("but no applications are registered", func() {
							BeforeEach(func() {
								fakeAuthClient.DoAuthenticatedGetReturns(httpclient.AuthClientResponse{bytes.NewBufferString(`
{
   "applications":{
       "application":[]
   }
}`), 200}, nil)
							})

							It("should not return an error", func() {
								Expect(err).NotTo(HaveOccurred())
							})

							It("should print a suitable message", func() {
								Expect(output).To(ContainSubstring("No registered applications found"))
							})

						})

						Context("but the cf app name cannot be determined", func() {
							Context("because the cf app GUID is not present in the registered metadata", func() {
								BeforeEach(func() {
									fakeAuthClient.DoAuthenticatedGetReturns(httpclient.AuthClientResponse{bytes.NewBufferString(`
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
         }
      ]
   }
}`), 200}, nil)
								})

								It("should not return an error", func() {
									Expect(err).NotTo(HaveOccurred())
								})

								It("should omit the cf app name and cf instance index", func() {
									tab := &format.Table{}
									tab.Entitle([]string{"eureka app name", "cf app name", "cf instance index", "zone", "status"})
									tab.AddRow([]string{"APP-1", "?????", "?", "zone1", "UP"})
									Expect(output).To(ContainSubstring(tab.String()))
								})
							})

							Context("because GetApps returns an error", func() {
								BeforeEach(func() {
									fakeCliConnection.GetAppsReturns([]plugin_models.GetAppsModel{}, errors.New("some error"))
								})

								It("should return a suitable error", func() {
									Expect(err).To(HaveOccurred())
									Expect(err.Error()).To(HavePrefix("Failed to determine cf app name corresponding to cf app GUID '062bd505-8b19-44ca-4451-4a932932143a': some error"))
								})
							})

						})

						Context("and the cf app name can be determined", func() {

							var (
								getAppsCallCount int
							)

							BeforeEach(func() {
								getAppsCallCount = 0
								fakeCliConnection.GetAppsStub = func() ([]plugin_models.GetAppsModel, error) {
									getAppsCallCount++
									apps := []plugin_models.GetAppsModel{}
									app1 := plugin_models.GetAppsModel{
										Name: "cfapp1",
										Guid: "062bd505-8b19-44ca-4451-4a932932143a",
									}
									app2 := plugin_models.GetAppsModel{
										Name: "cfapp2",
										Guid: "162bd505-1b19-14ca-1451-1a9329321431",
									}
									return append(apps, app1, app2), nil
								}
							})

							It("should have obtained an access token", func() {
								Expect(accessTokenCallCount).To(Equal(1))
							})

							It("should have sent a request to the correct URL with the correct access token", func() {
								Expect(fakeAuthClient.DoAuthenticatedGetCallCount()).To(Equal(1))
								url, accessToken := fakeAuthClient.DoAuthenticatedGetArgsForCall(0)
								Expect(url).To(Equal("https://eureka-dashboard-url/eureka/apps"))
								Expect(accessToken).To(Equal("someaccesstoken"))
							})

							It("should have looked up the cf app names", func() {
								Expect(getAppsCallCount).To(Equal(2))
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

							It("should return the registered applications", func() {
								tab := &format.Table{}
								tab.Entitle([]string{"eureka app name", "cf app name", "cf instance index", "zone", "status"})
								tab.AddRow([]string{"APP-1", "cfapp1", "2", "zone1", "UP"})
								tab.AddRow([]string{"APP-2", "cfapp2", "3", "zone2", "OUT_OF_SERVICE"})
								Expect(output).To(ContainSubstring(tab.String()))
							})
						})
					})
				})
			})
		})
	})
})
