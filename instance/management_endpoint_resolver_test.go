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
	const scs2DashboardUrl = "https://spring-cloud-broker.example.com/dashboard/p-config-server/ae019348-e113-4a84-8c8d-9b31ae3cb5ee"
	const scs3DashboardUrl = "https://config-server-1558aa56-97a5-47cd-a73c-7d5cd0818b22.example.com/dashboard"

	var (
		fakeCliConnection *pluginfakes.FakeCliConnection
		fakeAuthClient    *httpclientfakes.FakeAuthenticatedClient

		isServiceBrokerEndpoint bool

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

		isServiceBrokerEndpoint = true

		fakeCliConnection.ApiEndpointReturns("https://api.some.host.name", nil)
	})

	JustBeforeEach(func() {
		managementEndpointResolver := instance.NewAuthenticatedManagementEndpointResolver(fakeCliConnection, fakeAuthClient)
		output, err = managementEndpointResolver.GetManagementEndpoint(
			"serviceinstance",
			testAccessToken,
			isServiceBrokerEndpoint)
	})

	Context("when the service instance is not found", func() {
		BeforeEach(func() {
			fakeCliConnection.GetServiceReturns(plugin_models.GetService_Model{}, testError)
		})

		It("propagates the error", func() {
			Expect(err).To(MatchError("service instance not found: " + errMessage))
		})
	})

	Context("when an SCS 2 service instance is found", func() {
		BeforeEach(func() {
			fakeCliConnection.GetServiceReturns(plugin_models.GetService_Model{
				ServiceOffering: plugin_models.GetService_ServiceFields{
					Name: "p-config-server",
				},
				DashboardUrl: scs2DashboardUrl,
				Guid:         "guid",
			}, nil)
		})

		Context("when it's a lifecycle operation", func() {
			BeforeEach(func() {
				isServiceBrokerEndpoint = true
			})

			It("returns the cli endpoint on the broker", func() {
				Expect(output).To(Equal("https://spring-cloud-broker.example.com/cli/instances/guid"))
			})
		})

		Context("when it's not a lifecycle operation", func() {
			BeforeEach(func() {
				isServiceBrokerEndpoint = false
			})

			It("returns the cli endpoint on the broker", func() {
				Expect(output).To(Equal("https://spring-cloud-broker.example.com/cli/instances/guid"))
			})
		})
	})

	Context("when an SCS3 service instance is found", func() {
		BeforeEach(func() {
			fakeCliConnection.GetServiceReturns(plugin_models.GetService_Model{
				ServiceOffering: plugin_models.GetService_ServiceFields{
					Name: "p.config-server",
				},
				DashboardUrl: scs3DashboardUrl,
				Guid:         "guid",
			}, nil)
		})

		Context("when it's a lifecycle operation", func() {
			BeforeEach(func() {
				isServiceBrokerEndpoint = true
			})

			It("returns the cli endpoint on the broker", func() {
				Expect(output).To(Equal("https://scs-service-broker.some.host.name/cli/instances/guid"))
			})
		})

		Context("when it's not a lifecycle operation", func() {
			BeforeEach(func() {
				isServiceBrokerEndpoint = false
			})

			It("returns the cli endpoint on the backing app", func() {
				Expect(output).To(Equal("https://config-server-1558aa56-97a5-47cd-a73c-7d5cd0818b22.example.com/cli/instances/guid"))
			})
		})
	})
})
