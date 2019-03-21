/*
 * Copyright (C) 2016-Present Pivotal Software, Inc. All rights reserved.
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
package eureka_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"github.com/fatih/color"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/eureka"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/format"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient/httpclientfakes"
)

var _ = Describe("OperateOnApplication", func() {

	const testAccessToken = "someaccesstoken"

	type operationArg struct {
		accessToken   string
		eurekaUrl     string
		eurekaAppName string
		instanceId    string
	}

	var (
		fakeCliConnection   *pluginfakes.FakeCliConnection
		fakeAuthClient      *httpclientfakes.FakeAuthenticatedClient
		fakeResolver        func(cliConnection plugin.CliConnection, serviceInstanceName string, accessToken string, authClient httpclient.AuthenticatedClient) (string, error)
		resolverAccessToken string
		progressWriter      *bytes.Buffer
		output              string

		fakeOperation      eureka.InstanceOperation
		operationCallCount int
		operationArgs      []operationArg

		operationReturn error

		err           error
		instanceIndex *int
	)

	BeforeEach(func() {
		color.NoColor = false // ensure predictable colour behaviour independent of test environment

		fakeCliConnection = &pluginfakes.FakeCliConnection{}
		fakeAuthClient = &httpclientfakes.FakeAuthenticatedClient{}
		fakeAuthClient.DoAuthenticatedGetReturns(ioutil.NopCloser(bytes.NewBufferString("https://fake.com")), 200, nil)
		resolverAccessToken = ""
		fakeResolver = func(cliConnection plugin.CliConnection, serviceInstanceName string, accessToken string, authClient httpclient.AuthenticatedClient) (string, error) {
			return "https://eureka-dashboard-url/", nil
		}
		progressWriter = new(bytes.Buffer)

		operationCallCount = 0
		operationArgs = []operationArg{}
		operationReturn = nil
		fakeOperation = func(authClient httpclient.AuthenticatedClient, eurekaUrl string, eurekaAppName string, instanceId string, accessToken string) error {
			operationCallCount++
			operationArgs = append(operationArgs, operationArg{
				accessToken:   accessToken,
				eurekaUrl:     eurekaUrl,
				eurekaAppName: eurekaAppName,
				instanceId:    instanceId,
			})
			return operationReturn
		}
	})

	JustBeforeEach(func() {
		output, err = eureka.OperateOnApplication(fakeCliConnection, "some-service-registry", "some-cf-app", fakeAuthClient, instanceIndex, progressWriter, fakeResolver, fakeOperation)
	})

	It("should attempt to obtain an access token", func() {
		Expect(fakeCliConnection.AccessTokenCallCount()).To(Equal(1))
	})

	Context("when the access token is not available", func() {
		BeforeEach(func() {
			fakeCliConnection.AccessTokenReturns("", errors.New("some access token error"))
		})

		It("should return a suitable error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("Access token not available: some access token error"))
		})
	})

	Context("when the access token is available", func() {
		BeforeEach(func() {
			fakeCliConnection.AccessTokenReturns("bearer "+testAccessToken, nil)
		})

		Context("but the eureka URL cannot be resolved", func() {
			BeforeEach(func() {
				fakeResolver = func(cliConnection plugin.CliConnection, serviceInstanceName string, accessToken string, authClient httpclient.AuthenticatedClient) (string, error) {
					return "", errors.New("resolution error")
				}
			})

			It("should return a suitable error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("Error obtaining service registry URL: resolution error"))
			})
		})

		Context("when the eureka URL can be resolved", func() {
			var testErr error

			BeforeEach(func() {
				testErr = errors.New("failed")
				fakeResolver = func(cliConnection plugin.CliConnection, serviceInstanceName string, accessToken string, authClient httpclient.AuthenticatedClient) (string, error) {
					resolverAccessToken = accessToken
					return "https://spring-cloud-broker.some.host.name/x/y/z/some-guid/", nil
				}

				fakeCliConnection.GetAppsStub = func() ([]plugin_models.GetAppsModel, error) {
					apps := []plugin_models.GetAppsModel{}
					app1 := plugin_models.GetAppsModel{
						Name: "some-cf-app",
						Guid: "062bd505-8b19-44ca-4451-4a932932143a",
					}
					return append(apps, app1), nil
				}

				fakeAuthClient.DoAuthenticatedGetReturns(ioutil.NopCloser(bytes.NewBufferString(`
						{
							"applications":{
								"application":[
									{
										"instance":[
											{
												"app":"APP-1",
												"instanceId":"instance-1",
												"status":"UP",
												"metadata":{
													"zone":"zone-a",
													"cfAppGuid":"062bd505-8b19-44ca-4451-4a932932143a",
													"cfInstanceIndex":"2"
												}
											}
										]
									}
								]
							}
						}`)), 200, nil)

			})

			It("should pass the access token to the resolver", func() {
				Expect(resolverAccessToken).To(Equal(testAccessToken))
			})

			It("should pass the access token to the GET request", func() {
				url, tok := fakeAuthClient.DoAuthenticatedGetArgsForCall(0)
				Expect(url).To(Equal("https://spring-cloud-broker.some.host.name/x/y/z/some-guid/eureka/apps"))
				Expect(tok).To(Equal(testAccessToken))
			})

			It("should successfully operate on the service", func() {
				Expect(operationCallCount).To(Equal(1))
				args := operationArgs[0]
				Expect(args.accessToken).To(Equal(testAccessToken))
				Expect(args.eurekaUrl).To(Equal("https://spring-cloud-broker.some.host.name/x/y/z/some-guid/"))
				Expect(args.eurekaAppName).To(Equal("APP-1"))
				Expect(args.instanceId).To(Equal("instance-1"))
			})

			It("should log progress", func() {
				Expect(progressWriter.String()).To(Equal(fmt.Sprintf("Processing service instance %s with index %s\n", format.Bold(format.Cyan("APP-1")), format.Bold(format.Cyan("2")))))
			})

			Context("when the operation fails", func() {
				BeforeEach(func() {
					operationReturn = testErr
				})

				It("should return an error", func() {
					Expect(err.Error()).To(ContainSubstring("Operation failed"))
				})
			})

			Context("when obtaining the application instances from the service registry fails", func() {
				BeforeEach(func() {
					testErr = errors.New("failed")
					fakeAuthClient.DoAuthenticatedGetReturns(nil, 0, testErr)
				})

				It("should return the error", func() {
					Expect(err.Error()).To(ContainSubstring("Service registry error: failed"))
				})
			})

			Context("when obtaining the application instances from the service registry returns a bad status code", func() {
				BeforeEach(func() {
					fakeAuthClient.DoAuthenticatedGetReturns(nil, http.StatusNotFound, nil)
				})

				It("should return the error", func() {
					Expect(err.Error()).To(ContainSubstring("Service registry failed: 404"))
				})
			})

			Context("but only two out of three eureka instance names can be resolved", func() {
				BeforeEach(func() {

					fakeAuthClient.DoAuthenticatedGetReturns(ioutil.NopCloser(bytes.NewBufferString(`
						{
						   "applications":{
						      "application":[
							 {
							    "instance":[
							       {
								  "app":"APP-1",
 			                      "instanceId":"instance-1",
								  "status":"UP",
								  "metadata":{
								     "zone":"zone-a",
								     "cfAppGuid":"062bd505-8b19-44ca-4451-4a932932143a",
								     "cfInstanceIndex":"1"
								  }
							       },
							       {
								  "app":"APP-2",
 			                      "instanceId":"instance-1",
								  "status":"UNKNOWN",
								  "metadata":{
								     "zone":"zone-a",
								     "cfInstanceIndex":"2"
								  }
							       },
							       {
								  "app":"APP-3",
 			                      "instanceId":"instance-1",
								  "status":"UP",
								  "metadata":{
								     "zone":"zone-a",
								     "cfAppGuid":"062bd505-8b19-44ca-4451-4a932932143a",
								     "cfInstanceIndex":"3"
								  }
							       }
							    ]
							 }
						      ]
						   }
						}`)), 200, nil)

				})

				It("should operate on the instances with guids", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(operationCallCount).To(Equal(2))

					args := operationArgs[0]
					Expect(args.accessToken).To(Equal(testAccessToken))
					Expect(args.eurekaUrl).To(Equal("https://spring-cloud-broker.some.host.name/x/y/z/some-guid/"))
					Expect(args.eurekaAppName).To(Equal("APP-1"))
					Expect(args.instanceId).To(Equal("instance-1"))

					args = operationArgs[1]
					Expect(args.accessToken).To(Equal(testAccessToken))
					Expect(args.eurekaUrl).To(Equal("https://spring-cloud-broker.some.host.name/x/y/z/some-guid/"))
					Expect(args.eurekaAppName).To(Equal("APP-3"))
					Expect(args.instanceId).To(Equal("instance-1"))
				})

				It("should inform the user that 2 instances are being processed", func() {
					template := "Processing service instance %s with index %s\n"
					line1 := fmt.Sprintf(template, format.Bold(format.Cyan("APP-1")), format.Bold(format.Cyan("1")))
					line2 := fmt.Sprintf(template, format.Bold(format.Cyan("APP-3")), format.Bold(format.Cyan("3")))

					Expect(output).To(BeEmpty()) // only output is progress indication
					Expect(progressWriter.String()).To(ContainSubstring(line1 + line2))
				})
			})

			Context("but the cf app name cannot be found", func() {

				BeforeEach(func() {
					fakeCliConnection.GetAppsStub = func() ([]plugin_models.GetAppsModel, error) {
						apps := []plugin_models.GetAppsModel{}
						app1 := plugin_models.GetAppsModel{
							Name: "unknown-app",
							Guid: "062bd505-8b19-44ca-4451-4a932932143a",
						}
						return append(apps, app1), nil
					}
				})

				It("should return a suitable error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("cf app name some-cf-app not found"))
				})
			})

			Context("when an instance index is specified", func() {
				BeforeEach(func() {
					fakeCliConnection.GetAppsStub = func() ([]plugin_models.GetAppsModel, error) {
						apps := []plugin_models.GetAppsModel{}
						app1 := plugin_models.GetAppsModel{
							Name: "some-cf-app",
							Guid: "062bd505-8b19-44ca-4451-4a932932143a",
						}
						return append(apps, app1), nil
					}
					fakeAuthClient.DoAuthenticatedGetReturns(ioutil.NopCloser(bytes.NewBufferString(`
						{
						   "applications":{
							  "application":[
							 		{
										"instance":[
								  			{
								  				"app":"APP-1",
								  				"instanceId":"instance-1",
								  				"status":"UP",
								  				"metadata":{
									 				"zone":"zone-a",
									 				"cfAppGuid":"062bd505-8b19-44ca-4451-4a932932143a",
									 				"cfInstanceIndex":"0"
								  				}
								   			},
								 			{
								  				"app":"APP-1",
								  				"instanceId":"instance-2",
								  				"status":"UP",
								  				"metadata":{
									 				"zone":"zone-a",
									 				"cfAppGuid":"062bd505-8b19-44ca-4451-4a932932143a",
									 				"cfInstanceIndex":"1"
								  				}
								   			},
								   			{
								  				"app":"APP-1",
								  				"instanceId":"instance-3",
								  				"status":"UP",
								  				"metadata":{
										 			"zone":"zone-a",
									 				"cfAppGuid":"062bd505-8b19-44ca-4451-4a932932143a",
									 				"cfInstanceIndex":"2"
								  				}
								   			}
										]
							 		}
							  ]
						   }
						}`)), 200, nil)

					//Set the instance index argument
					var idx = 1
					instanceIndex = &idx
				})

				It("should not raise an error", func() {
					Expect(err).ToNot(HaveOccurred())
				})

				It("should process just the required instance", func() {
					Expect(operationCallCount).To(Equal(1))
					args := operationArgs[0]
					Expect(args.accessToken).To(Equal(testAccessToken))
					Expect(args.eurekaUrl).To(Equal("https://spring-cloud-broker.some.host.name/x/y/z/some-guid/"))
					Expect(args.eurekaAppName).To(Equal("APP-1"))
					Expect(args.instanceId).To(Equal("instance-2"))
				})

				It("should inform the user about the instance deregistration", func() {
					template := "Processing service instance %s with index %s\n"
					line1 := fmt.Sprintf(template, format.Bold(format.Cyan("APP-1")), format.Bold(format.Cyan("1")))

					Expect(output).To(BeEmpty()) // only output is progress indication
					Expect(progressWriter.String()).To(ContainSubstring(line1))
				})

				Context("when the operation fails", func() {
					BeforeEach(func() {
						operationReturn = testErr
					})

					It("should return a suitable error", func() {
						Expect(err).To(MatchError("Operation failed"))
					})
				})

				Context("when an incorrect instance index is specified", func() {
					BeforeEach(func() {
						var idx = 99
						instanceIndex = &idx
					})

					It("should return a suitable error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("No instance found with index 99"))
					})
				})
			})

			Context("when an invalid instance index is returned in the metadata", func() {
				BeforeEach(func() {
					fakeCliConnection.GetAppsStub = func() ([]plugin_models.GetAppsModel, error) {
						apps := []plugin_models.GetAppsModel{}
						app1 := plugin_models.GetAppsModel{
							Name: "some-cf-app",
							Guid: "062bd505-8b19-44ca-4451-4a932932143a",
						}
						return append(apps, app1), nil
					}
					fakeAuthClient.DoAuthenticatedGetReturns(ioutil.NopCloser(bytes.NewBufferString(`
						{
						   "applications":{
							  "application":[
								{
									"instance":[
										{
								  			"app":"APP-1",
								  			"instanceId":"instance-1",
								 			"status":"UP",
								  			"metadata":{
									 			"zone":"zone-a",
									 			"cfAppGuid":"062bd505-8b19-44ca-4451-4a932932143a",
									 			"cfInstanceIndex":"bad-integer"
									  		}
									   	}
									]
							 	}
							  ]
						   }
						}`)), 200, nil)

					//Set the instance index argument
					var idx = 1
					instanceIndex = &idx
				})

				It("should raise a suitable error", func() {
					Expect(err.Error()).To(ContainSubstring(`parsing "bad-integer": invalid syntax`))
				})
			})
		})
	})
})
