package eureka_test

import (
	"errors"
	"strings"

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
		fakeAuthClient.DoAuthenticatedGetReturns(bytes.NewBufferString("https://fake.com"), nil)
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
						fakeAuthClient.DoAuthenticatedGetReturns(bytes.NewBufferString(`{"authenticated":true}`), errors.New("some error"))
					})

					It("should return a suitable error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("Service registry error: some error"))
					})
				})

				Context("and eureka responds", func() {
					Context("but the response body contains invalid JSON", func() {
						BeforeEach(func() {
							fakeAuthClient.DoAuthenticatedGetReturns(bytes.NewBufferString(""), nil)
						})

						It("should return a suitable error", func() {
							Expect(err).To(HaveOccurred())
							Expect(err).To(MatchError("Invalid service registry response JSON: unexpected end of JSON input, response body: ''"))
						})
					})

					Context("and the response is valid", func() {
						BeforeEach(func() {
							fakeAuthClient.DoAuthenticatedGetReturns(bytes.NewBufferString(`
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
}`), nil)
						})

						Context("but no applications are registered", func() {
							BeforeEach(func() {
								fakeAuthClient.DoAuthenticatedGetReturns(bytes.NewBufferString(`
{
   "applications":{
       "application":[]
   }
}`), nil)
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
									fakeAuthClient.DoAuthenticatedGetReturns(bytes.NewBufferString(`
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
}`), nil)
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

							Context("because curl returns an error", func() {
								BeforeEach(func() {
									fakeCliConnection.CliCommandWithoutTerminalOutputReturns([]string{}, errors.New("some error"))
								})

								It("should return a suitable error", func() {
									Expect(err).To(HaveOccurred())
									Expect(err.Error()).To(HavePrefix("Failed to determine cf app name corresponding to cf app GUID '062bd505-8b19-44ca-4451-4a932932143a': some error"))
								})
							})

							Context("because curl returns a failure payload", func() {
								BeforeEach(func() {
									fakeCliConnection.CliCommandWithoutTerminalOutputReturns([]string{`{`, `"code": 100004,`, `"description": "The app could not be found: 062bd505-8b19-44ca-4451-4a932932143a",`, `"error_code": "CF-AppNotFound"`, `}`}, nil)
								})

								It("should return a suitable error", func() {
									Expect(err).To(HaveOccurred())
									Expect(err.Error()).To(HavePrefix("Failed to determine cf app name corresponding to cf app GUID '062bd505-8b19-44ca-4451-4a932932143a': The app could not be found: 062bd505-8b19-44ca-4451-4a932932143a: code 100004, error_code CF-AppNotFound"))
								})
							})

							Context("because curl returns unsuitable JSON", func() {
								BeforeEach(func() {
									fakeCliConnection.CliCommandWithoutTerminalOutputReturns([]string{`{`, `"name": 99`, `}`}, nil)
								})

								It("should return a suitable error", func() {
									Expect(err).To(HaveOccurred())
									Expect(err.Error()).To(HavePrefix("Failed to determine cf app name corresponding to cf app GUID '062bd505-8b19-44ca-4451-4a932932143a': json: cannot unmarshal number into Go value of type string"))
								})
							})
						})

						Context("and the cf app name can be determined", func() {

							var (
								cliCommandWithoutTerminalOutputCallCount int
								cliCommandArgs                           [][]string
							)

							BeforeEach(func() {
								cliCommandWithoutTerminalOutputCallCount = 0
								cliCommandArgs = [][]string{}
								fakeCliConnection.CliCommandWithoutTerminalOutputStub = func(args ...string) ([]string, error) {
									cliCommandWithoutTerminalOutputCallCount++
									cliCommandArgs = append(cliCommandArgs, args)
									var cfAppName string
									if strings.Contains(args[1], "062bd505-8b19-44ca-4451-4a932932143a") {
										cfAppName = "cfapp1"
									} else {
										cfAppName = "cfapp2"
									}
									return []string{`{`, `"name": "` + cfAppName + `"`, `}`}, nil
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
								Expect(cliCommandWithoutTerminalOutputCallCount).To(Equal(2))
								Expect(len(cliCommandArgs)).To(Equal(2))

								argsForApp1 := cliCommandArgs[0]
								Expect(len(argsForApp1)).To(Equal(4))
								Expect(argsForApp1[0]).To(Equal("curl"))
								Expect(argsForApp1[1]).To(Equal("/v2/apps/062bd505-8b19-44ca-4451-4a932932143a/summary"))
								Expect(argsForApp1[2]).To(Equal("-H"))
								Expect(argsForApp1[3]).To(Equal("Accept: application/json"))

								argsForApp2 := cliCommandArgs[1]
								Expect(len(argsForApp2)).To(Equal(4))
								Expect(argsForApp2[0]).To(Equal("curl"))
								Expect(argsForApp2[1]).To(Equal("/v2/apps/162bd505-1b19-14ca-1451-1a9329321431/summary"))
								Expect(argsForApp2[2]).To(Equal("-H"))
								Expect(argsForApp2[3]).To(Equal("Accept: application/json"))
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
