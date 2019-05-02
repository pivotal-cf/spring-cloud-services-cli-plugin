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

	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
)

type restageOperation struct {
	authenticatedClient httpclient.AuthenticatedClient
}

func (ro *restageOperation) Run(serviceInstanceAdminURL string, accessToken string) (string, error) {
	_, err := ro.authenticatedClient.DoAuthenticatedPut(fmt.Sprintf("%s/command?restage=", serviceInstanceAdminURL), accessToken)
	return "", err
}

func (ro *restageOperation) IsServiceBrokerEndpoint() bool {
	return true
}

func NewRestageOperation(authenticatedClient httpclient.AuthenticatedClient) Operation {
	return &restageOperation{
		authenticatedClient: authenticatedClient,
	}
}
