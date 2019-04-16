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
	"errors"
	"fmt"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
	"net/url"
	"strings"
)

type ManagementEndpointResolver func(cliConnection plugin.CliConnection, authClient httpclient.AuthenticatedClient, serviceInstanceName string, accessToken string, isLifecycleOperation bool) (string, error)

func GetManagementEndpoint(cliConnection plugin.CliConnection, authClient httpclient.AuthenticatedClient, serviceInstanceName string, accessToken string, isLifecycleOperation bool) (string, error) {
	serviceModel, err := cliConnection.GetService(serviceInstanceName)
	if err != nil {
		return "", fmt.Errorf("service instance not found: %s", err)
	}

	isVersion3, err := isVersion3(cliConnection, authClient, accessToken)
	if err != nil {
		return "", err
	}

	if isVersion3 && isLifecycleOperation {
		serviceBrokerV3Url, err := serviceBrokerV3Url(cliConnection)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%s/cli/instances/%s", serviceBrokerV3Url, serviceModel.Guid), nil
	}

	parsedUrl, err := url.Parse(serviceModel.DashboardUrl)
	if err != nil {
		return "", err
	}

	parsedUrl.Path = fmt.Sprintf("/cli/instances/%s",  serviceModel.Guid)

	return parsedUrl.String(), nil
}

func isVersion3(cliConnection plugin.CliConnection, authClient httpclient.AuthenticatedClient, accessToken string) (bool, error) {
	serviceBrokerV3Url, err := serviceBrokerV3Url(cliConnection)
	if err != nil {
		return false, err
	}

	_, statusCode, err := authClient.DoAuthenticatedGet(serviceBrokerV3Url + "/actuator/info", accessToken)

	if err != nil && statusCode != 404 {
		return false, err
	}

	if statusCode == 200 {
		return true, nil
	}

	return false, nil
}

func serviceBrokerV3Url(cliConnection plugin.CliConnection) (string, error) {
	apiUrl, err := cliConnection.ApiEndpoint()
	if err != nil {
		return "", err
	}

	posFirst := strings.Index(apiUrl, "api.")
	if posFirst < 0 {
		return "", errors.New(fmt.Sprintf("unexpected CF API endpoint: %s", apiUrl))
	}

	systemDomain := apiUrl[posFirst+4:]
	serviceBrokerV3Url := "https://scs-service-broker." + systemDomain
	return serviceBrokerV3Url, nil
}
