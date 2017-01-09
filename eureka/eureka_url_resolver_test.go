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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/eureka"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient/httpclientfakes"
)

var _ = Describe("eurekaUrlFromDashboardUrl", func() {

	var (
		dashboardUrl string
		accessToken  string
		authClient   *httpclientfakes.FakeAuthenticatedClient
		eurekaUrl    string
		err          error
	)

	BeforeEach(func() {
		authClient = &httpclientfakes.FakeAuthenticatedClient{}
	})

	JustBeforeEach(func() {
		eurekaUrl, err = eureka.EurekaUrlFromDashboardUrl(dashboardUrl, accessToken, authClient)
	})

	Context("when the dashboard URL is not in the correct format", func() {
		Context("because it is malformed", func() {
			BeforeEach(func() {
				dashboardUrl = "://"
			})

			It("should return a suitable error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("parse ://: missing protocol scheme"))
			})
		})

		Context("because its path format is invalid", func() {
			BeforeEach(func() {
				dashboardUrl = "https://spring-cloud-broker.some.host.name"
			})

			It("should return a suitable error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("path of https://spring-cloud-broker.some.host.name has no segments"))
			})
		})
	})

	Context("when the dashboard URL is in the correct format", func() {
		BeforeEach(func() {
			dashboardUrl = "https://spring-cloud-broker.some.host.name/x/y/guid"
		})

		Context("when eureka cannot be contacted", func() {
			BeforeEach(func() {
				authClient.DoAuthenticatedGetReturns(nil, 502, errors.New("some error"))
			})

			It("should return a suitable error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("Invalid service registry definition response: some error"))
			})
		})

		Context("when eureka can be contacted", func() {
			Context("but the returned buffer is nil", func() {
				BeforeEach(func() {
					authClient.DoAuthenticatedGetReturns(nil, 200, nil)
				})

				It("should return a suitable error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("Buffer is nil"))
				})
			})

			Context("but the response body contains invalid JSON", func() {
				BeforeEach(func() {
					authClient.DoAuthenticatedGetReturns(bytes.NewBufferString(""), 200, nil)
				})

				It("should return a suitable error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("Invalid service registry definition response JSON: unexpected end of JSON input, response body: ''"))
				})
			})

			Context("but the response body has the wrong content", func() {
				BeforeEach(func() {
					authClient.DoAuthenticatedGetReturns(bytes.NewBufferString(`{"credentials":0}`), 200, nil)
				})

				It("should return a suitable error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(`Invalid service registry definition response JSON: json: cannot unmarshal number into Go value of type struct { Uri string }, response body: '{"credentials":0}'`))
				})
			})

			Context("but the '/cli/instance endpoint cannot be found", func() {
				BeforeEach(func() {
					authClient.DoAuthenticatedGetReturns(bytes.NewBufferString(`{"credentials":0}`), 404, nil)
				})
				It("should warn that the SCS version might be too old", func() {
					versionWarning := "The /cli/instance endpoint could not be found.\n"

					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(versionWarning))
				})
			})

			Context("and the response body has the correct content", func() {
				BeforeEach(func() {
					authClient.DoAuthenticatedGetReturns(bytes.NewBufferString(`{"credentials":{"uri":"https://eurekadashboardurl"}}`), 200, nil)
				})

				It("should return a suitable error", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(eurekaUrl).To(Equal("https://eurekadashboardurl/"))
				})
			})
		})
	})
})
