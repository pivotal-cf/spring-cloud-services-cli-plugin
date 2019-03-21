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
	"fmt"
	"net/url"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/cfutil"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
)

type Operation func(authClient httpclient.AuthenticatedClient, serviceInstanceAdminURL string, accessToken string) (string, error)

func RunOperation(cliConnection plugin.CliConnection, authClient httpclient.AuthenticatedClient, serviceInstanceName string, operation Operation) (string, error) {
	serviceInstanceAdminURL, accessToken, err := accessServiceBroker(serviceInstanceName, cliConnection)
	if err != nil {
		return "", err
	}

	return operation(authClient, serviceInstanceAdminURL, accessToken)
}

func accessServiceBroker(serviceInstanceName string, cliConnection plugin.CliConnection) (string, string, error) {
	accessToken, err := cfutil.GetToken(cliConnection)
	if err != nil {
		return "", "", err
	}

	serviceModel, err := cliConnection.GetService(serviceInstanceName)
	if err != nil {
		return "", "", fmt.Errorf("Service instance not found: %s", err)
	}

	parsedUrl, err := url.Parse(serviceModel.DashboardUrl)
	if err != nil {
		return "", "", err
	}
	path := parsedUrl.Path

	segments := strings.Split(path, "/")
	if len(segments) == 0 || (len(segments) == 1 && segments[0] == "") {
		return "", "", fmt.Errorf("path of %s has no segments", serviceModel.DashboardUrl)
	}
	guid := segments[len(segments)-1]

	parsedUrl.Path = fmt.Sprintf("/cli/instances/%s", guid)

	return parsedUrl.String(), accessToken, nil
}
