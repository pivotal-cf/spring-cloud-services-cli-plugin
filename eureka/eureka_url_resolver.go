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
package eureka

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"io/ioutil"

	"net/http"

	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
)

func EurekaUrlFromDashboardUrl(dashboardUrl string, accessToken string, authClient httpclient.AuthenticatedClient) (string, error) {
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

	bodyReader, statusCode, err := authClient.DoAuthenticatedGet(parsedUrl.String(), accessToken)

	//In the case of a 404, the most likely cause is that the CLI version is greater than the broker version.
	if statusCode == http.StatusNotFound {
		return "", errors.New("The /cli/instance endpoint could not be found.\n" +
			"This could be because the Spring Cloud Services broker version is too old.\n" +
			"Please ensure SCS is at least version 1.3.3.\n")
	}
	var serviceDefinitionResp ServiceDefinitionResp
	if err != nil {
		return "", fmt.Errorf("Invalid service registry definition response: %s", err)
	}

	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return "", fmt.Errorf("Cannot read service registry definition response body: %s", err)
	}

	err = json.Unmarshal(body, &serviceDefinitionResp)
	if serviceDefinitionResp.Credentials.Uri == "" {
		return "", fmt.Errorf("Invalid service registry definition response JSON: %s, response body: '%s'", err, string(body))

	}
	if err != nil {
		return "", fmt.Errorf("JSON reponse contained empty property 'credentials.url': %s", string(body))
	}
	return serviceDefinitionResp.Credentials.Uri + "/", nil
}
