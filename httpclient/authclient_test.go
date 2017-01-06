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
package httpclient_test

import (
	"io/ioutil"
	"net/http"
	"strings"

	"errors"

	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient/httpclientfakes"
)

type badReader struct{}

func (b badReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

var _ = Describe("Authclient", func() {
	const testAccessToken = "securetoken"
	const testUrl = "https://eureka.pivotal.io/auth/request"
	var (
		fakeClient  *httpclientfakes.FakeClient
		authClient  httpclient.AuthenticatedClient
		url         string
		accessToken string
		body        *bytes.Buffer
		status      int
		err         error
	)

	BeforeEach(func() {
		fakeClient = &httpclientfakes.FakeClient{}
		url = testUrl
		accessToken = testAccessToken
	})

	JustBeforeEach(func() {
		authClient = httpclient.NewAuthenticatedClient(fakeClient)
		body, status, err = authClient.DoAuthenticatedGet(url, accessToken)
	})

	Context("when the underlying request cannot be created", func() {
		BeforeEach(func() {
			url = ":"
		})

		It("should return a suitable error if the request cannot be created", func() {
			Expect(body).To(BeNil())
			Expect(err).To(MatchError("Request creation error: parse :: missing protocol scheme"))
		})
	})

	Context("when the underlying request can be created", func() {
		Context("and an authenticated request is made", func() {

			Context("but the request fails", func() {
				BeforeEach(func() {
					fakeClient.DoReturns(nil, errors.New(`request failed`))
					authClient = httpclient.NewAuthenticatedClient(fakeClient)
				})

				It("should produce an error", func() {
					Expect(body).To(BeNil())
					Expect(err).To(MatchError("authenticated get of 'https://eureka.pivotal.io/auth/request' failed: request failed"))
				})
			})

			Context("and the request succeeds", func() {
				Context("but the body body is nil", func() {
					BeforeEach(func() {
						resp := &http.Response{}
						//resp.Body = nil
						fakeClient.DoReturns(resp, nil)
					})

					It("should produce an error", func() {
						Expect(body).To(BeNil())
						Expect(err).To(MatchError("authenticated get of 'https://eureka.pivotal.io/auth/request' failed: nil response body"))
					})
				})

				Context("but the body body cannot be read", func() {
					BeforeEach(func() {
						resp := &http.Response{}

						resp.Body = ioutil.NopCloser(badReader{})
						fakeClient.DoReturns(resp, nil)
					})

					It("should produce an error", func() {
						Expect(body).To(BeNil())
						Expect(err).To(MatchError("authenticated get of 'https://eureka.pivotal.io/auth/request' failed: body cannot be read"))
					})
				})

				Context("and the body body can be read", func() {
					BeforeEach(func() {
						resp := &http.Response{}
						resp.Body = ioutil.NopCloser(strings.NewReader("payload"))
						fakeClient.DoReturns(resp, nil)
					})

					It("should have sent a request with the correct accept header", func() {
						Expect(fakeClient.DoCallCount()).To(Equal(1))
						req := fakeClient.DoArgsForCall(0)
						Expect(req.Header.Get("Accept")).To(Equal("application/json"))
					})

					It("should have sent a request with the correct authorization header", func() {
						Expect(fakeClient.DoCallCount()).To(Equal(1))
						req := fakeClient.DoArgsForCall(0)
						Expect(req.Header.Get("Authorization")).To(Equal("securetoken"))
					})

					It("should produce a non-empty body body", func() {
						Expect(body.String()).Should(Equal("payload"))
						Expect(err).To(BeNil())
					})
				})
			})
		})
	})
})
