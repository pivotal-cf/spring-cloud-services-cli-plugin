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
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/instance"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/instance/instancefakes"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/serviceutil"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/serviceutil/serviceutilfakes"
)

var _ = Describe("OperationRunner", func() {

	const testAccessToken = "someaccesstoken"

	var (
		operationRunner instance.OperationRunner

		fakeCliConnection *pluginfakes.FakeCliConnection
		fakeOperation     *instancefakes.FakeOperation
		output            string

		fakeServiceInstanceResolver *serviceutilfakes.FakeServiceInstanceResolver

		errMessage string
		testError  error

		serviceInstanceName string

		err error
	)

	BeforeEach(func() {
		fakeCliConnection = &pluginfakes.FakeCliConnection{}
		fakeOperation = &instancefakes.FakeOperation{}

		fakeOperation.IsServiceBrokerOperationReturns(true)

		fakeServiceInstanceResolver = &serviceutilfakes.FakeServiceInstanceResolver{}

		fakeServiceInstanceResolver.GetManagementParametersReturns(serviceutil.ManagementParameters{
			Url: "https://spring-cloud-broker.some.host.name/cli/instances/guid",
		}, nil)

		errMessage = "failure is not an option"
		testError = errors.New(errMessage)

		serviceInstanceName = "serviceinstance"
	})

	JustBeforeEach(func() {
		operationRunner = instance.NewAuthenticatedOperationRunner(fakeCliConnection, fakeServiceInstanceResolver)
		output, err = operationRunner.RunOperation(
			serviceInstanceName,
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

		Context("when the admin parameters are not retrieved correctly", func() {
			BeforeEach(func() {
				fakeServiceInstanceResolver.GetManagementParametersReturns(serviceutil.ManagementParameters{}, errors.New("some error retrieving the admin parameters"))
			})

			It("should return a suitable error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("some error retrieving the admin parameters"))
			})
		})

		Context("when the admin parameters are retrieved correctly", func() {
			It("invoke the operation with the correct parameters", func() {
				Expect(fakeOperation.RunCallCount()).To(Equal(1))
				serviceInstanceAdminParameters, accessToken := fakeOperation.RunArgsForCall(0)

				Expect(serviceInstanceAdminParameters).To(Equal(
					serviceutil.ManagementParameters{
						Url: "https://spring-cloud-broker.some.host.name/cli/instances/guid",
					}))
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
