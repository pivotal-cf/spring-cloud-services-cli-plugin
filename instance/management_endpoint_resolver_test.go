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
	"code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient/httpclientfakes"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/instance"
)

var _ = Describe("GetManagementEndpoint", func() {

	const testAccessToken = "someaccesstoken"

	var (
		fakeCliConnection *pluginfakes.FakeCliConnection
		fakeAuthClient    *httpclientfakes.FakeAuthenticatedClient

		serviceInstanceName string

		isLifecycleOperation bool

		errMessage string
		testError  error

		output string
		err    error
	)

	BeforeEach(func() {
		fakeCliConnection = &pluginfakes.FakeCliConnection{}
		fakeAuthClient = &httpclientfakes.FakeAuthenticatedClient{}

		errMessage = "failure is not an option"
		testError = errors.New(errMessage)

		isLifecycleOperation = true

		serviceInstanceName = "serviceinstance"
	})

	JustBeforeEach(func() {
		managementEndpointResolver := instance.NewAuthenticatedManagementEndpointResolver(fakeCliConnection, fakeAuthClient)
		output, err = managementEndpointResolver.GetManagementEndpoint(
			serviceInstanceName,
			testAccessToken,
			isLifecycleOperation)
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
			Expect(err).To(MatchError("service instance not found: " + errMessage))
		})
	})

	Context("when the service instance is found", func() {
		BeforeEach(func() {
			fakeCliConnection.GetServiceReturns(plugin_models.GetService_Model{
				DashboardUrl: "https://dasboard-url.some.host.name/x/y/guid",
				Guid:         "guid",
			}, nil)
		})

		Context("when there is an error checking SCS version", func() {
			Context("getting the API endpoint", func() {
				BeforeEach(func() {
					fakeCliConnection.ApiEndpointReturns("", errors.New("error getting API endpoint"))
				})

				It("should propagate the error", func() {
					Expect(err).To(MatchError("error getting API endpoint"))
				})
			})

			Context("checking the V3 broker URL", func() {
				BeforeEach(func() {
					fakeCliConnection.ApiEndpointReturns("https://api.some.host.name", nil)
					fakeAuthClient.DoAuthenticatedGetReturns(nil, 500, errors.New("error pinging V3 broker"))
				})

				It("should propagate the error", func() {
					Expect(err).To(MatchError("error pinging V3 broker"))
				})
			})
		})

		Context("when we are in SCS3", func() {
			BeforeEach(func() {
				fakeCliConnection.ApiEndpointReturns("https://api.some.host.name", nil)
				fakeAuthClient.DoAuthenticatedGetReturns(nil, 200, nil)
			})

			Context("when it's a lifecycle operation", func() {
				BeforeEach(func() {
					isLifecycleOperation = true
				})

				It("should return the broker URL for SCS3", func() {
					Expect(output).To(Equal("https://scs-service-broker.some.host.name/cli/instances/guid"))
				})
			})

			Context("when it's not a lifecycle operation", func() {
				BeforeEach(func() {
					isLifecycleOperation = false
				})

				It("should return the dashboard URL", func() {
					Expect(output).To(Equal("https://dasboard-url.some.host.name/cli/instances/guid"))
				})
			})
		})

		Context("when we are in SCS2", func() {
			BeforeEach(func() {
				fakeCliConnection.ApiEndpointReturns("https://api.some.host.name", nil)
				fakeAuthClient.DoAuthenticatedGetReturns(nil, 404, nil)
			})

			Context("when it's a lifecycle operation", func() {
				BeforeEach(func() {
					isLifecycleOperation = true
				})

				It("should return the broker URL for SCS3", func() {
					Expect(output).To(Equal("https://dasboard-url.some.host.name/cli/instances/guid"))
				})
			})

			Context("when it's not a lifecycle operation", func() {
				BeforeEach(func() {
					isLifecycleOperation = false
				})

				It("should return the dashboard URL", func() {
					Expect(output).To(Equal("https://dasboard-url.some.host.name/cli/instances/guid"))
				})
			})
		})
	})
})
