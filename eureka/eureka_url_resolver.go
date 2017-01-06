/*
 * Copyright 2016-2017 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
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

	response, err := authClient.DoAuthenticatedGet(parsedUrl.String(), accessToken)
	buffer := response.Body

	//In the case of a 404, the most likely cause is that the CLI version is greater than the broker version.
	if response.StatusCode == 404 {
		return "", errors.New("The /cli/instance endpoint could not be found.\n" +
			"This could be because the Spring Cloud Services broker version is too old.\n" +
			"Please ensure SCS is at least version 1.3.3.\n")
	}
	var serviceDefinitionResp ServiceDefinitionResp
	if err != nil {
		return "", fmt.Errorf("Invalid service registry definition response: %s", err)
	}
	if buffer == nil {
		return "", errors.New("Buffer is nil")
	}

	err = json.Unmarshal(buffer.Bytes(), &serviceDefinitionResp)
	if serviceDefinitionResp.Credentials.Uri == "" {
		return "", fmt.Errorf("Invalid service registry definition response JSON: %s, response body: '%s'", err, string(buffer.Bytes()))

	}
	if err != nil {
		return "", fmt.Errorf("JSON reponse contained empty property 'credentials.url': %s", string(buffer.Bytes()))
	}
	return serviceDefinitionResp.Credentials.Uri + "/", nil
}
