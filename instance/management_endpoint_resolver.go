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
package instance

import (
	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/cli/plugin/models"
	"fmt"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
	"net/url"
	"strings"
)

//go:generate counterfeiter -o resolverfakes/fake_management_endpoint_resolver.go . ManagementEndpointResolver
type ManagementEndpointResolver interface {
	GetManagementEndpoint(serviceInstanceName string, accessToken string, isLifecycleOperation bool) (string, error)
}

type authenticatedManagementEndpointResolver struct {
	cliConnection plugin.CliConnection
	authClient    httpclient.AuthenticatedClient
}

func (amer *authenticatedManagementEndpointResolver) GetManagementEndpoint(serviceInstanceName string, accessToken string, isLifecycleOperation bool) (string, error) {
	serviceModel, err := amer.cliConnection.GetService(serviceInstanceName)
	if err != nil {
		return "", fmt.Errorf("service instance not found: %s", err)
	}

	if isVersion3(serviceModel) && isLifecycleOperation {
		serviceBrokerV3Url, err := serviceBrokerV3Url(amer.cliConnection)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%s/cli/instances/%s", serviceBrokerV3Url, serviceModel.Guid), nil
	}

	parsedUrl, err := url.Parse(serviceModel.DashboardUrl)
	if err != nil {
		return "", err
	}

	parsedUrl.Path = fmt.Sprintf("/cli/instances/%s", serviceModel.Guid)

	return parsedUrl.String(), nil
}

func isVersion3(serviceModel plugin_models.GetService_Model) bool {
	return strings.HasPrefix(serviceModel.ServiceOffering.Name, "p.")
}

func serviceBrokerV3Url(cliConnection plugin.CliConnection) (string, error) {
	apiUrl, err := cliConnection.ApiEndpoint()
	if err != nil {
		return "", err
	}

	posFirst := strings.Index(apiUrl, ".")
	systemDomain := apiUrl[posFirst+1:]
	serviceBrokerV3Url := "https://scs-service-broker." + systemDomain
	return serviceBrokerV3Url, nil
}

func NewAuthenticatedManagementEndpointResolver(
	cliConnection plugin.CliConnection,
	authClient httpclient.AuthenticatedClient) ManagementEndpointResolver {

	return &authenticatedManagementEndpointResolver{
		cliConnection: cliConnection,
		authClient:    authClient,
	}
}
