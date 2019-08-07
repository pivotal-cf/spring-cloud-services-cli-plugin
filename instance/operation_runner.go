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
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/serviceutil"
)

//go:generate counterfeiter . Operation
type Operation interface {
	Run(serviceInstanceAdminParameters serviceutil.ManagementParameters, accessToken string) (string, error)
	IsServiceBrokerOperation() bool
}

type OperationRunner interface {
	RunOperation(serviceInstanceName string, operation Operation) (string, error)
}

type authenticatedOperationRunner struct {
	cliConnection              plugin.CliConnection
	serviceInstanceUrlResolver serviceutil.ServiceInstanceResolver
}

func NewAuthenticatedOperationRunner(
	cliConnection plugin.CliConnection,
	serviceInstanceUrlResolver serviceutil.ServiceInstanceResolver) OperationRunner {

	return &authenticatedOperationRunner{
		cliConnection:              cliConnection,
		serviceInstanceUrlResolver: serviceInstanceUrlResolver,
	}
}

func (aor *authenticatedOperationRunner) RunOperation(
	serviceInstanceName string,
	operation Operation) (string, error) {

	accessToken, err := cfutil.GetToken(aor.cliConnection)
	if err != nil {
		return "", err
	}

	managementParameters, err := aor.serviceInstanceUrlResolver.GetManagementParameters(
		serviceInstanceName,
		accessToken,
		operation.IsServiceBrokerOperation())

	if err != nil {
		return "", err
	}

	return operation.Run(managementParameters, accessToken)
}
