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

var _ = Describe("ServiceInstanceURL", func() {

	const errMessage = "some error"

	var (
		fakeCliConnection   *pluginfakes.FakeCliConnection
		serviceInstanceName string
		accessToken         string
		authClient          *httpclientfakes.FakeAuthenticatedClient
		serviceInstanceURL  string
		err                 error
		testError           error
	)

	BeforeEach(func() {
		fakeCliConnection = &pluginfakes.FakeCliConnection{}
		serviceInstanceName = "service-instance-name"
		authClient = &httpclientfakes.FakeAuthenticatedClient{}
		testError = errors.New(errMessage)
	})

	JustBeforeEach(func() {
		serviceInstanceURL, err = serviceutil.ServiceInstanceURL(fakeCliConnection, serviceInstanceName, accessToken, authClient)
	})

	It("should get the service", func() {
		Expect(fakeCliConnection.GetServiceCallCount()).To(Equal(1))
		Expect(fakeCliConnection.GetServiceArgsForCall(0)).To(Equal(serviceInstanceName))
	})

	Context("when the service instance is not found", func() {
		BeforeEach(func() {
			fakeCliConnection.GetServiceReturns(plugin_models.GetService_Model{}, testError)
		})

		It("should propagate the error", func() {
			Expect(err).To(MatchError("Service instance not found: " + errMessage))
		})
	})

	Context("when the dashboard URL is not in the correct format", func() {
		Context("because it is malformed", func() {
			BeforeEach(func() {
				fakeCliConnection.GetServiceReturns(plugin_models.GetService_Model{
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
					DashboardUrl: "https://spring-cloud-broker.some.host.name",
				}, nil)
			})

			It("should return a suitable error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("path of https://spring-cloud-broker.some.host.name has no segments"))
			})
		})
	})

	Context("when the dashboard URL is in the correct format", func() {
		BeforeEach(func() {
			fakeCliConnection.GetServiceReturns(plugin_models.GetService_Model{
				DashboardUrl: "https://spring-cloud-broker.some.host.name/x/y/guid",
			}, nil)
		})

		Context("when dashboard cannot be contacted", func() {
			BeforeEach(func() {
				authClient.DoAuthenticatedGetReturns(nil, 502, testError)
			})

			It("should return a suitable error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("Invalid service definition response: " + errMessage))
			})
		})

		Context("when dashboard can be contacted", func() {
			Context("but the response body contains invalid JSON", func() {
				BeforeEach(func() {
					authClient.DoAuthenticatedGetReturns(ioutil.NopCloser(bytes.NewBufferString("")), 200, nil)
				})

				It("should return a suitable error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("JSON response failed to unmarshal: "))
				})
			})

			Context("but the response body has the wrong content", func() {
				BeforeEach(func() {
					authClient.DoAuthenticatedGetReturns(ioutil.NopCloser(bytes.NewBufferString(`{"credentials":0}`)), 200, nil)
				})

				It("should return a suitable error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(`JSON response failed to unmarshal: {"credentials":0}`))
				})
			})

			Context("but the '/cli/instance endpoint cannot be found", func() {
				BeforeEach(func() {
					authClient.DoAuthenticatedGetReturns(ioutil.NopCloser(bytes.NewBufferString(`{"credentials":0}`)), 404, nil)
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
					authClient.DoAuthenticatedGetReturns(&badReader{readErr: testError}, 200, nil)
				})

				It("should return a suitable error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("Cannot read service definition response body: " + errMessage))
				})
			})

			Context("and the response body has an empty URI", func() {
				BeforeEach(func() {
					authClient.DoAuthenticatedGetReturns(ioutil.NopCloser(bytes.NewBufferString(`{"credentials":{"uri":""}}`)), 200, nil)
				})

				It("should return a suitable error", func() {
					serviceDefinitionError := `JSON response contained empty property 'credentials.url', response body: '{"credentials":{"uri":""}}'`

					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(serviceDefinitionError))
				})
			})

			Context("and the response body has the correct content", func() {
				BeforeEach(func() {
					authClient.DoAuthenticatedGetReturns(ioutil.NopCloser(bytes.NewBufferString(`{"credentials":{"uri":"https://service.instance.url"}}`)), 200, nil)
				})

				It("should return a suitable error", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(serviceInstanceURL).To(Equal("https://service.instance.url/"))
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
