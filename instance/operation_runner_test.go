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
package instance_test

import (
	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient/httpclientfakes"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/instance"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/instance/operationfakes"
)

var _ = Describe("RunOperation", func() {

	const testAccessToken = "someaccesstoken"

	var (
		fakeCliConnection *pluginfakes.FakeCliConnection
		fakeAuthClient    *httpclientfakes.FakeAuthenticatedClient
		fakeOperation     *operationfakes.FakeOperation
		output            string

		fakeManagementEndpointResolver instance.ManagementEndpointResolver

		errMessage string
		testError  error

		serviceInstanceName string

		err error
	)

	BeforeEach(func() {
		fakeCliConnection = &pluginfakes.FakeCliConnection{}
		fakeAuthClient = &httpclientfakes.FakeAuthenticatedClient{}
		fakeOperation = &operationfakes.FakeOperation{}

		fakeOperation.IsLifecycleOperationReturns(true)

		fakeManagementEndpointResolver = func(cliConnection plugin.CliConnection, authClient httpclient.AuthenticatedClient, serviceInstanceName string, accessToken string, isLifecycleOperation bool) (string, error) {
			return "https://spring-cloud-broker.some.host.name/cli/instances/guid", nil
		}

		errMessage = "failure is not an option"
		testError = errors.New(errMessage)

		serviceInstanceName = "serviceinstance"
	})

	JustBeforeEach(func() {
		output, err = instance.RunOperation(
			fakeCliConnection,
			fakeAuthClient,
			serviceInstanceName,
			fakeManagementEndpointResolver,
			fakeOperation)
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

		Context("when the admin URL is not retrieved correctly", func() {
			BeforeEach(func() {
				fakeManagementEndpointResolver = func(cliConnection plugin.CliConnection, authClient httpclient.AuthenticatedClient, serviceInstanceName string, accessToken string, isLifecycleOperation bool) (string, error) {
					return "", errors.New("some error retrieving the admin URL")
				}
			})

			It("should return a suitable error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("some error retrieving the admin URL"))
			})
		})

		Context("when the admin URL is retrieved correctly", func() {
			It("invoke the operation with the correct parameters", func() {
				Expect(fakeOperation.RunCallCount()).To(Equal(1))
				_, serviceInstanceAdminURL, accessToken := fakeOperation.RunArgsForCall(0)

				Expect(serviceInstanceAdminURL).To(Equal("https://spring-cloud-broker.some.host.name/cli/instances/guid"))
				Expect(accessToken).To(Equal(testAccessToken))
			})

			Context("when the operation returns some output", func() {
				BeforeEach(func() {
					fakeOperation.RunReturns("some output", nil)
				})

				It("should return the output", func() {
					Expect(output).To(Equal("some output"))
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when the operation returns an error", func() {
				BeforeEach(func() {
					fakeOperation.RunReturns("", testError)
				})

				It("should return the error", func() {
					Expect(err).To(Equal(testError))
				})
			})
		})
	})
})
