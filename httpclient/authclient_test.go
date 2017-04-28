/*
 * Copyright (C) 2016-Present Pivotal Software, Inc. All rights reserved.
 *
 * This program and the accompanying materials are made available under
 * the terms of the under the Apache License, Version 2.0 (the "License”);
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
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

	"io"

	"fmt"

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
	const (
		testUrl               = "https://eureka.pivotal.io/auth/request"
		testAccessToken       = "access-token"
		testBearerAccessToken = "bearer " + testAccessToken
		errMessage            = "I'm sorry Dave, I'm afraid I can't do that."
	)

	var (
		fakeClient *httpclientfakes.FakeClient
		URL        string
		testErr    error
		err        error
		status     int
	)

	BeforeEach(func() {
		fakeClient = &httpclientfakes.FakeClient{}
		testErr = errors.New(errMessage)
	})

	Describe("DoAuthenticatedGet", func() {
		var (
			body io.ReadCloser
		)

		BeforeEach(func() {
			URL = testUrl
			resp := &http.Response{StatusCode: http.StatusOK}
			resp.Body = ioutil.NopCloser(strings.NewReader("payload"))
			fakeClient.DoReturns(resp, nil)
		})

		JustBeforeEach(func() {
			authClient := httpclient.NewAuthenticatedClient(fakeClient)
			body, status, err = authClient.DoAuthenticatedGet(URL, testAccessToken)
		})

		Context("when the underlying request cannot be created", func() {
			BeforeEach(func() {
				URL = ":"
			})

			It("should return a suitable error if the request cannot be created", func() {
				Expect(body).To(BeNil())
				Expect(err).To(MatchError("Request creation error: parse :: missing protocol scheme"))
			})
		})

		It("should send a request with the correct accept header", func() {
			Expect(fakeClient.DoCallCount()).To(Equal(1))
			req := fakeClient.DoArgsForCall(0)
			Expect(req.Header.Get("Accept")).To(Equal("application/json"))
		})

		It("should send a request with the correct authorization header", func() {
			Expect(fakeClient.DoCallCount()).To(Equal(1))
			req := fakeClient.DoArgsForCall(0)
			Expect(req.Header.Get("Authorization")).To(Equal(testBearerAccessToken))
		})

		It("should pass the response back", func() {
			Expect(err).NotTo(HaveOccurred())
			op, readErr := ioutil.ReadAll(body)
			Expect(readErr).NotTo(HaveOccurred())
			Expect(string(op)).Should(Equal("payload"))
		})

		Context("when the request fails", func() {
			BeforeEach(func() {
				fakeClient.DoReturns(nil, testErr)
			})

			It("should produce an error", func() {
				Expect(body).To(BeNil())
				Expect(err).To(MatchError(fmt.Sprintf("Authenticated get of 'https://eureka.pivotal.io/auth/request' failed: %s", errMessage)))
			})
		})

		Context("when the request returns a bad status", func() {
			BeforeEach(func() {
				resp := &http.Response{StatusCode: http.StatusNotFound, Status: "404 Not found"}
				fakeClient.DoReturns(resp, nil)
			})

			It("should return the error", func() {
				Expect(body).To(BeNil())
				Expect(err).To(MatchError("Authenticated get of 'https://eureka.pivotal.io/auth/request' failed: 404 Not found"))
			})
		})
	})

	Describe("DoAuthenticatedDelete", func() {
		BeforeEach(func() {
			URL = testUrl
			resp := &http.Response{StatusCode: http.StatusOK}
			fakeClient.DoReturns(resp, nil)
		})

		JustBeforeEach(func() {
			authClient := httpclient.NewAuthenticatedClient(fakeClient)
			status, err = authClient.DoAuthenticatedDelete(URL, testAccessToken)
		})

		Context("when the URL is invalid", func() {
			BeforeEach(func() {
				URL = ":"
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError("Request creation error: parse :: missing protocol scheme"))
			})
		})

		It("should send a request with the correct accept header", func() {
			Expect(fakeClient.DoCallCount()).To(Equal(1))
			req := fakeClient.DoArgsForCall(0)
			Expect(req.Header.Get("Accept")).To(Equal("application/json"))
		})

		It("should send a request with the correct authorization header", func() {
			Expect(fakeClient.DoCallCount()).To(Equal(1))
			req := fakeClient.DoArgsForCall(0)
			Expect(req.Header.Get("Authorization")).To(Equal(testBearerAccessToken))
		})

		It("should pass the status code back", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(http.StatusOK))
		})

		Context("when the request fails", func() {
			BeforeEach(func() {
				fakeClient.DoReturns(nil, testErr)
			})

			It("should produce an error", func() {
				Expect(err).To(MatchError(fmt.Sprintf("Authenticated delete of 'https://eureka.pivotal.io/auth/request' failed: %s", errMessage)))
			})
		})

		Context("when the request returns a bad status", func() {
			BeforeEach(func() {
				resp := &http.Response{StatusCode: http.StatusNotFound, Status: "404 Not found"}
				fakeClient.DoReturns(resp, nil)
			})

			It("should return the error", func() {
				Expect(err).To(MatchError("Authenticated delete of 'https://eureka.pivotal.io/auth/request' failed: 404 Not found"))
			})
		})
	})

	Describe("DoAuthenticatedPost", func() {
		const (
			testBodyType = "body-type"
			testBody     = "body"
		)

		var (
			respBody io.ReadCloser
			bodyType string
			body     string
		)

		BeforeEach(func() {
			URL = testUrl
			bodyType = testBodyType
			body = testBody
			resp := &http.Response{StatusCode: http.StatusOK}
			fakeClient.DoReturns(resp, nil)
		})

		JustBeforeEach(func() {
			authClient := httpclient.NewAuthenticatedClient(fakeClient)
			respBody, status, err = authClient.DoAuthenticatedPost(URL, bodyType, body, testAccessToken)
		})

		Context("when the URL is invalid", func() {
			BeforeEach(func() {
				URL = ":"
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError("Request creation error: parse :: missing protocol scheme"))
			})
		})

		It("should send a request with the correct body", func() {
			Expect(fakeClient.DoCallCount()).To(Equal(1))
			req := fakeClient.DoArgsForCall(0)
			bodyContents, readErr := ioutil.ReadAll(req.Body)
			Expect(readErr).NotTo(HaveOccurred())
			Expect(string(bodyContents)).To(Equal(testBody))
		})

		It("should send a request with the correct authorization header", func() {
			Expect(fakeClient.DoCallCount()).To(Equal(1))
			req := fakeClient.DoArgsForCall(0)
			Expect(req.Header.Get("Authorization")).To(Equal(testBearerAccessToken))
		})

		It("should send a request with the correct content type header", func() {
			Expect(fakeClient.DoCallCount()).To(Equal(1))
			req := fakeClient.DoArgsForCall(0)
			Expect(req.Header.Get("Content-Type")).To(Equal(bodyType))
		})

		It("should pass the status code back", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(http.StatusOK))
		})

		Context("when the request fails", func() {
			BeforeEach(func() {
				fakeClient.DoReturns(nil, testErr)
			})

			It("should produce an error", func() {
				Expect(err).To(MatchError(fmt.Sprintf("Authenticated post to 'https://eureka.pivotal.io/auth/request' failed: %s", errMessage)))
			})
		})

		Context("when the request returns a bad status", func() {
			BeforeEach(func() {
				resp := &http.Response{StatusCode: http.StatusNotFound, Status: "404 Not found"}
				fakeClient.DoReturns(resp, nil)
			})

			It("should return the error", func() {
				Expect(err).To(MatchError("Authenticated post to 'https://eureka.pivotal.io/auth/request' failed: 404 Not found"))
			})
		})
	})
})
