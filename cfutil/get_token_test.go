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
package cfutil_test

import (
	"errors"

	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/cfutil"
)

var _ = Describe("GetToken", func() {
	const (
		errMessage = "no dice"
		testToken  = "some-token"
	)

	var (
		fakeCliConnection *pluginfakes.FakeCliConnection
		tok               string
		err               error
		testError         error
	)

	BeforeEach(func() {
		fakeCliConnection = &pluginfakes.FakeCliConnection{}
		testError = errors.New(errMessage)
	})

	JustBeforeEach(func() {
		tok, err = cfutil.GetToken(fakeCliConnection)
	})

	It("should attempt to obtain an access token", func() {
		Expect(fakeCliConnection.AccessTokenCallCount()).To(Equal(1))
	})

	Context("when the token is not available", func() {
		BeforeEach(func() {
			fakeCliConnection.AccessTokenReturns("", testError)
		})

		It("should propagate the error", func() {
			Expect(err).To(MatchError("Access token not available: " + errMessage))
		})
	})

	Context("when the output of obtaining the access token is in the expected form without a trailing newline", func() {
		BeforeEach(func() {
			fakeCliConnection.AccessTokenReturns("bearer "+testToken, nil)
		})

		It("should return the token", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(tok).To(Equal(testToken))
		})
	})

	Context("when the output of obtaining the access token is in the expected form with a trailing newline", func() {
		BeforeEach(func() {
			fakeCliConnection.AccessTokenReturns("bearer "+testToken+"\n", nil)
		})

		It("should return the token", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(tok).To(Equal(testToken))
		})
	})

	Context("when the output of obtaining the access token is invalid", func() {
		Context("when the output does not start with 'bearer '", func() {
			BeforeEach(func() {
				fakeCliConnection.AccessTokenReturns("bearer"+testToken, nil)
			})

			It("should return an error", func() {
				Expect(err).To(MatchError("Access token output invalid: bearersome-token"))
			})
		})

		Context("when the output has too few tokens", func() {
			BeforeEach(func() {
				fakeCliConnection.AccessTokenReturns("bearer ", nil)
			})

			It("should return an error", func() {
				Expect(err).To(MatchError("Access token output invalid: bearer "))
			})
		})

		Context("when the output has too many tokens", func() {
			BeforeEach(func() {
				fakeCliConnection.AccessTokenReturns("bearer x y", nil)
			})

			It("should return an error", func() {
				Expect(err).To(MatchError("Access token output invalid: bearer x y"))
			})
		})
	})
})
