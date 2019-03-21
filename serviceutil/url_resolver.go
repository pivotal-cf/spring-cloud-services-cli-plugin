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

// ServiceInstanceURL obtains the service instance URL of a service with a specific name. This is a secure operation and an access token is provided for authentication and authorisation.
func ServiceInstanceURL(cliConnection plugin.CliConnection, serviceInstanceName string, accessToken string, authClient httpclient.AuthenticatedClient) (string, error) {
	serviceModel, err := cliConnection.GetService(serviceInstanceName)
	if err != nil {
		return "", fmt.Errorf("Service instance not found: %s", err)
	}

	parsedUrl, err := url.Parse(serviceModel.DashboardUrl)
	if err != nil {
		return "", err
	}
	path := parsedUrl.Path

	segments := strings.Split(path, "/")
	if len(segments) == 0 || (len(segments) == 1 && segments[0] == "") {
		return "", fmt.Errorf("path of %s has no segments", serviceModel.DashboardUrl)
	}
	guid := segments[len(segments)-1]

	parsedUrl.Path = "/cli/instance/" + guid

	bodyReader, statusCode, err := authClient.DoAuthenticatedGet(parsedUrl.String(), accessToken)

	//In the case of a 404, the most likely cause is that the CLI version is greater than the broker version.
	if statusCode == http.StatusNotFound {
		return "", errors.New("The /cli/instance endpoint could not be found.\n" +
			"This could be because the Spring Cloud Services broker version is too old.\n" +
			"Please ensure SCS is at least version 1.3.3.\n")
	}
	var serviceDefinitionResp serviceDefinitionResp
	if err != nil {
		return "", fmt.Errorf("Invalid service definition response: %s", err)
	}

	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return "", fmt.Errorf("Cannot read service definition response body: %s", err)
	}

	err = json.Unmarshal(body, &serviceDefinitionResp)
	if err != nil {
		return "", fmt.Errorf("JSON response failed to unmarshal: %s", string(body))
	}
	if serviceDefinitionResp.Credentials.URI == "" {
		return "", fmt.Errorf("JSON response contained empty property 'credentials.url', response body: '%s'", string(body))

	}
	return serviceDefinitionResp.Credentials.URI + "/", nil
}
