/*
 * Copyright 2016-2017 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package eureka

import (
	"fmt"

	"errors"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
)

func Deregister(cliConnection plugin.CliConnection, srInstanceName string, cfAppName string, authenticatedClient httpclient.AuthenticatedClient) (string, error) {
	return DeregisterWithResolver(cliConnection, srInstanceName, cfAppName, authenticatedClient, EurekaUrlFromDashboardUrl)
}

func DeregisterWithResolver(cliConnection plugin.CliConnection, srInstanceName string, cfAppName string, authClient httpclient.AuthenticatedClient,
	eurekaUrlFromDashboardUrl func(dashboardUrl string, accessToken string, authClient httpclient.AuthenticatedClient) (string, error)) (string, error) {
	serviceModel, err := cliConnection.GetService(srInstanceName)
	if err != nil {
		return "", fmt.Errorf("Service registry instance not found: %s", err)
	}
	accessToken, err := cliConnection.AccessToken()
	if err != nil {
		return "", fmt.Errorf("Access token not available: %s", err)
	}

	eureka, err := eurekaUrlFromDashboardUrl(serviceModel.DashboardUrl, accessToken, authClient)
	if err != nil {
		return "", fmt.Errorf("Error obtaining service registry dashboard URL: %s", err)
	}

	apps, err := getRegisteredAppsWithCfAppName(cliConnection, authClient, accessToken, eureka, cfAppName)
	if err != nil {
		return "", err
	}

	for _, app := range apps {
		err = deregister(authClient, accessToken, eureka, app.eurekaAppName, app.instanceId)
	}

	return "", nil
}

func deregister(authClient httpclient.AuthenticatedClient, accessToken string, eureka string, eurekaAppName string, instanceId string) error {
	return authClient.DoAuthenticatedDelete(eureka+fmt.Sprintf("eureka/apps/%s/%s", eurekaAppName, instanceId), accessToken)
}

func getRegisteredAppsWithCfAppName(cliConnection plugin.CliConnection, authClient httpclient.AuthenticatedClient, accessToken string, eureka string, cfAppName string) ([]eurekaAppRecord, error) {
	registeredAppsWithCfAppName := []eurekaAppRecord{}

	registeredApps, err := getRegisteredApps(cliConnection, authClient, accessToken, eureka)

	if err != nil {
		return registeredAppsWithCfAppName, err
	}

	for _, app := range registeredApps {
		if app.cfAppName == cfAppName {
			registeredAppsWithCfAppName = append(registeredAppsWithCfAppName, app)
		}
	}
	if len(registeredAppsWithCfAppName) == 0 {
		return registeredAppsWithCfAppName, errors.New(fmt.Sprintf("cf app name %s not found", cfAppName))
	}

	return registeredAppsWithCfAppName, nil
}
