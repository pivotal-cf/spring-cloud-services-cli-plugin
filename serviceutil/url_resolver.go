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
package serviceutil

import (
	plugin_models "code.cloudfoundry.org/cli/plugin/models"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
)

type serviceDefinitionResp struct {
	Credentials struct {
		URI string
	}
}

//go:generate counterfeiter . ServiceInstanceUrlResolver
type ServiceInstanceUrlResolver interface {
	GetServiceInstanceUrl(serviceInstanceName string, accessToken string) (string, error)
	GetManagementUrl(serviceInstanceName string, accessToken string, lifecycleOperation bool) (string, error)
}

type serviceInstanceUrlResolver struct {
	cliConnection plugin.CliConnection
	authClient    httpclient.AuthenticatedClient
}

func NewServiceInstanceUrlResolver(cliConnection plugin.CliConnection, authClient httpclient.AuthenticatedClient) ServiceInstanceUrlResolver {
	return &serviceInstanceUrlResolver{
		cliConnection: cliConnection,
		authClient:    authClient,
	}
}

func (s *serviceInstanceUrlResolver) GetServiceInstanceUrl(serviceInstanceName string, accessToken string) (string, error) {
	serviceModel, err := s.cliConnection.GetService(serviceInstanceName)
	if err != nil {
		return "", fmt.Errorf("Service instance not found: %s", err)
	}

	if isV2ServiceInstance(serviceModel) {
		return s.getV2ServiceInstanceUrl(serviceModel.DashboardUrl, accessToken)
	} else {
		return s.getV3ServiceInstanceUrl(serviceModel.DashboardUrl, accessToken)
	}
}

func (s *serviceInstanceUrlResolver) GetManagementUrl(serviceInstanceName string, accessToken string, lifecycleOperation bool) (string, error) {
	serviceModel, err := s.cliConnection.GetService(serviceInstanceName)
	if err != nil {
		return "", fmt.Errorf("Service instance not found: %s", err)
	}

	if isV2ServiceInstance(serviceModel) {
		return s.getV2ManagementUrl(serviceModel)
	} else {
		return s.getV3ManagementUrl(serviceModel, lifecycleOperation)
	}
}

func (s *serviceInstanceUrlResolver) getV2ServiceInstanceUrl(dashboardUrl string, accessToken string) (string, error) {
	parsedUrl, err := url.Parse(dashboardUrl)
	if err != nil {
		return "", err
	}
	path := parsedUrl.Path

	segments := strings.Split(path, "/")
	if len(segments) == 0 || (len(segments) == 1 && segments[0] == "") {
		return "", fmt.Errorf("path of %s has no segments", dashboardUrl)
	}
	guid := segments[len(segments)-1]

	parsedUrl.Path = "/cli/instance/" + guid

	bodyReader, statusCode, err := s.authClient.DoAuthenticatedGet(parsedUrl.String(), accessToken)

	//In the case of a 404, the most likely cause is that the CLI version is greater than the broker version.
	if statusCode == http.StatusNotFound {
		return "", errors.New("The /cli/instance endpoint could not be found.\n" +
			"This could be because the Spring Cloud Services broker version is too old.\n" +
			"Please ensure SCS is at least version 1.3.3.\n")
	}
	if err != nil {
		return "", fmt.Errorf("Invalid service definition response: %s", err)
	}

	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return "", fmt.Errorf("Cannot read service definition response body: %s", err)
	}

	var serviceDefinitionResp serviceDefinitionResp
	err = json.Unmarshal(body, &serviceDefinitionResp)
	if err != nil {
		return "", fmt.Errorf("JSON response failed to unmarshal: %s", string(body))
	}
	if serviceDefinitionResp.Credentials.URI == "" {
		return "", fmt.Errorf("JSON response contained empty property 'credentials.url', response body: '%s'", string(body))

	}
	return serviceDefinitionResp.Credentials.URI + "/", nil
}

func (s *serviceInstanceUrlResolver) getV3ServiceInstanceUrl(dashboardUrl string, accessToken string) (string, error) {
	parsedUrl, err := url.Parse(dashboardUrl)
	if err != nil {
		return "", err
	}

	parsedUrl.Path = ""

	return fmt.Sprintf("%s/", parsedUrl.String()), nil
}

func (s *serviceInstanceUrlResolver) getV2ManagementUrl(serviceModel plugin_models.GetService_Model) (string, error) {
	parsedUrl, err := url.Parse(serviceModel.DashboardUrl)
	if err != nil {
		return "", err
	}

	parsedUrl.Path = fmt.Sprintf("/cli/instances/%s", serviceModel.Guid)

	return parsedUrl.String(), nil
}

func (s *serviceInstanceUrlResolver) getV3ManagementUrl(serviceModel plugin_models.GetService_Model, serviceBrokerOperation bool) (string, error) {
	var parsedUrl *url.URL
	var err error

	if serviceBrokerOperation {
		serviceBrokerV3Url, err := s.getV3ServiceBrokerUrl()
		if err != nil {
			return "", err
		}
		parsedUrl, err = url.Parse(serviceBrokerV3Url)
	} else {
		parsedUrl, err = url.Parse(serviceModel.DashboardUrl)
	}
	if err != nil {
		return "", err
	}

	parsedUrl.Path = fmt.Sprintf("/cli/instances/%s", serviceModel.Guid)
	return parsedUrl.String(), nil
}

func (s *serviceInstanceUrlResolver) getV3ServiceBrokerUrl() (string, error) {
	apiUrl, err := s.cliConnection.ApiEndpoint()
	if err != nil {
		return "", err
	}

	posFirst := strings.Index(apiUrl, ".")
	systemDomain := apiUrl[posFirst+1:]
	serviceBrokerV3Url := "https://scs-service-broker." + systemDomain
	return serviceBrokerV3Url, nil
}

func isV2ServiceInstance(serviceModel plugin_models.GetService_Model) bool {
	return strings.HasPrefix(serviceModel.ServiceOffering.Name, "p-")
}
