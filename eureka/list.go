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

	"code.cloudfoundry.org/cli/plugin"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/format"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
)

const (
	UnknownCfAppName       = "?????"
	UnknownCfInstanceIndex = "?"
)

type Instance struct {
	App        string
	InstanceId string
	Status     string
	Metadata   struct {
		CfAppGuid       string
		CfInstanceIndex string
		Zone            string
	}
}

type ApplicationInstance struct {
	Instance []Instance
}

type ListResp struct {
	Applications struct {
		Application []ApplicationInstance
	}
}

type SummaryResp struct {
	Name string
}

type SummaryFailure struct {
	Code        int
	Description string
	Error_code  string
}

func List(cliConnection plugin.CliConnection, srInstanceName string, authClient httpclient.AuthenticatedClient) (string, error) {
	return ListWithResolver(cliConnection, srInstanceName, authClient, EurekaUrlFromDashboardUrl)
}

func ListWithResolver(cliConnection plugin.CliConnection, srInstanceName string, authClient httpclient.AuthenticatedClient,
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
	tab := &format.Table{}
	tab.Entitle([]string{"eureka app name", "cf app name", "cf instance index", "zone", "status"})
	registeredApps, err := getAllRegisteredApps(cliConnection, authClient, accessToken, eureka)

	if err != nil {
		return "", err
	}
	if len(registeredApps) == 0 {
		return fmt.Sprintf("Service instance: %s\nServer URL: %s\n\nNo registered applications found\n", srInstanceName, eureka), nil
	}
	for _, app := range registeredApps {

		tab.AddRow([]string{app.eurekaAppName, app.cfAppName, app.instanceIndex, app.zone, app.status})
	}

	return fmt.Sprintf("Service instance: %s\nServer URL: %s\n\n%s", srInstanceName, eureka, tab.String()), nil
}
