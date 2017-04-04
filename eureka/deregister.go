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
	"fmt"

	"errors"

	"strconv"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/format"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
)

func Deregister(cliConnection plugin.CliConnection, srInstanceName string, cfAppName string, authenticatedClient httpclient.AuthenticatedClient, instanceIndex *int) (string, error) {
	return DeregisterWithResolver(cliConnection, srInstanceName, cfAppName, authenticatedClient, instanceIndex, EurekaUrlFromDashboardUrl)
}

func DeregisterWithResolver(cliConnection plugin.CliConnection, srInstanceName string, cfAppName string, authClient httpclient.AuthenticatedClient, instanceIndex *int,
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
	statusMessage := ""
	statusTemplate := "Deregistered service instance %s with index %s\n"
	if instanceIndex == nil { //Index is omitted, deregister all instances
		var err error
		for _, app := range apps {
			err = deregister(authClient, accessToken, eureka, app.eurekaAppName, app.instanceId)
			statusMessage += fmt.Sprintf(statusTemplate, format.Bold(format.Cyan(app.eurekaAppName)), format.Bold(format.Cyan(app.instanceIndex)))
		}
		if err != nil {
			return "", fmt.Errorf("Error deregistering service instance: %s", err)
		}
	} else { //Instance ID provided, deregister a single instance
		app, err := getRegisteredAppByInstanceIndex(apps, *instanceIndex)
		if err != nil {
			return "", err
		}
		err = deregister(authClient, accessToken, eureka, app.eurekaAppName, app.instanceId)
		statusMessage += fmt.Sprintf(statusTemplate, format.Bold(format.Cyan(app.eurekaAppName)), format.Bold(format.Cyan(app.instanceIndex)))

		if err != nil {
			return "", fmt.Errorf("Error deregistering service instance: %s", err)
		}
	}
	return statusMessage, nil
}

func deregister(authClient httpclient.AuthenticatedClient, accessToken string, eureka string, eurekaAppName string, instanceId string) error {
	_, err := authClient.DoAuthenticatedDelete(eureka+fmt.Sprintf("eureka/apps/%s/%s", eurekaAppName, instanceId), accessToken)
	return err
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

func getRegisteredAppByInstanceIndex(appRecords []eurekaAppRecord, requestedIndex int) (eurekaAppRecord, error) {
	for _, app := range appRecords {
		registeredIndex, err := strconv.Atoi(app.instanceIndex)
		if err != nil {
			return eurekaAppRecord{}, err
		}
		if registeredIndex == requestedIndex {
			return app, nil
		}
	}
	return eurekaAppRecord{}, fmt.Errorf("No instance found with index %d", requestedIndex)
}
