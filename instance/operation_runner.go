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
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/cfutil"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
)

//go:generate counterfeiter -o operationfakes/fake_operation.go . Operation
type Operation interface {
	Run(authClient httpclient.AuthenticatedClient, serviceInstanceAdminURL string, accessToken string) (string, error)
	IsLifecycleOperation() bool
}

func RunOperation(
	cliConnection plugin.CliConnection,
	authClient httpclient.AuthenticatedClient,
	serviceInstanceName string,
	managementEndpointResolver ManagementEndpointResolver,
	operation Operation) (string, error) {

	accessToken, err := cfutil.GetToken(cliConnection)
	if err != nil {
		return "", err
	}

	serviceInstanceAdminURL, err := managementEndpointResolver(
		cliConnection,
		authClient,
		serviceInstanceName,
		accessToken,
		operation.IsLifecycleOperation())

	if err != nil {
		return "", err
	}

	return operation.Run(authClient, serviceInstanceAdminURL, accessToken)
}


