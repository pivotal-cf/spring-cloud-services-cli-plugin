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
	"errors"
	"fmt"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
	"io/ioutil"
	"net/http"
)

type parametersOperation struct{}

func (so *parametersOperation) Run(authenticatedClient httpclient.AuthenticatedClient, serviceInstanceAdminURL string, accessToken string) (string, error) {
	bodyReader, statusCode, err := authenticatedClient.DoAuthenticatedGet(serviceInstanceAdminURL+"/parameters", accessToken)
	if err != nil {
		return "", err
	}

	if statusCode != http.StatusOK {
		return "", fmt.Errorf("Service broker view instance configuration failed: %d", statusCode)
	}

	if bodyReader == nil {
		return "", errors.New("Service broker view instance configuration response body missing")
	}

	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return "", fmt.Errorf("Cannot read service instance configuration response body: %s", err)
	}

	return string(body), nil
}

func (so *parametersOperation) IsLifecycleOperation() (bool) {
	return false
}

func NewParametersOperation() *parametersOperation {
	return &parametersOperation{}
}
