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
	"fmt"

	"errors"

	"io/ioutil"

	"net/http"

	"strconv"

	"io"

	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/cli/plugin/models"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/cfutil"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/format"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
)

// Functions for accessing the eureka service registry.

type eurekaAppRecord struct {
	cfAppGuid     string
	cfAppName     string
	eurekaAppName string
	instanceId    string
	status        string
	zone          string
	instanceIndex string
}

func getRegisteredApps(cliConnection plugin.CliConnection, authClient httpclient.AuthenticatedClient, accessToken string, eurekaUrl string) ([]eurekaAppRecord, error) {
	appRecords := []eurekaAppRecord{}
	allAppRecords, err := getAllRegisteredApps(cliConnection, authClient, accessToken, eurekaUrl)
	if err != nil {
		return appRecords, err
	}

	for _, ar := range allAppRecords {
		if ar.cfAppGuid != "" {
			appRecords = append(appRecords, ar)
		}
	}
	return appRecords, nil
}

func getAllRegisteredApps(cliConnection plugin.CliConnection, authClient httpclient.AuthenticatedClient, accessToken string, eurekaUrl string) ([]eurekaAppRecord, error) {
	registeredApps := []eurekaAppRecord{}
	bodyReader, statusCode, err := authClient.DoAuthenticatedGet(eurekaUrl+"eureka/apps", accessToken)
	if err != nil {
		return registeredApps, fmt.Errorf("Service registry error: %s", err)
	}
	if statusCode != http.StatusOK {
		return registeredApps, fmt.Errorf("Service registry failed: %d", statusCode)
	}

	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return registeredApps, fmt.Errorf("Cannot read service registry response body: %s", err)
	}

	var listResp ListResp
	err = json.Unmarshal(body, &listResp)
	if err != nil {
		return registeredApps, fmt.Errorf("Invalid service registry response JSON: %s, response body: '%s'", err, string(body))
	}

	cfApps, err := cliConnection.GetApps()
	if err != nil {
		return registeredApps, err
	}

	apps := listResp.Applications.Application
	for _, app := range apps {
		instances := app.Instance
		for _, instance := range instances {
			metadata := instance.Metadata
			cfAppGuid := metadata.CfAppGuid
			cfInstanceIndex := metadata.CfInstanceIndex
			var cfAppNm string
			if cfAppGuid == "" {
				fmt.Printf("cf app GUID not present in metadata of eureka app %s. Perhaps the app was built with an old version of Spring Cloud Services starters.\n", instance.App)
				cfAppNm = UnknownCfAppName
				cfInstanceIndex = UnknownCfInstanceIndex
			} else {
				cfAppNm, err = cfAppName(cfApps, metadata.CfAppGuid)
				if err != nil {
					return registeredApps, fmt.Errorf("Failed to determine cf app name corresponding to cf app GUID '%s': %s", metadata.CfAppGuid, err)
				}

			}
			registeredApps = append(registeredApps, eurekaAppRecord{
				cfAppGuid:     cfAppGuid,
				cfAppName:     cfAppNm,
				eurekaAppName: instance.App,
				instanceId:    instance.InstanceId,
				instanceIndex: cfInstanceIndex,
				zone:          instance.Metadata.Zone,
				status:        instance.Status,
			})
		}
	}
	return registeredApps, nil
}

func cfAppName(cfApps []plugin_models.GetAppsModel, cfAppGuid string) (string, error) {
	for _, app := range cfApps {
		if app.Guid == cfAppGuid {
			return app.Name, nil
		}
	}

	return "", errors.New("CF App not found")
}

// Utility for operating on an application instance in the service registry

type InstanceOperation func(authClient httpclient.AuthenticatedClient, eurekaUrl string, eurekaAppName string, instanceId string, accessToken string) error

func OperateOnApplication(cliConnection plugin.CliConnection, srInstanceName string, cfAppName string, authClient httpclient.AuthenticatedClient, instanceIndex *int, progressWriter io.Writer,
	serviceInstanceURL func(cliConnection plugin.CliConnection, serviceInstanceName string, accessToken string, authClient httpclient.AuthenticatedClient) (string, error),
	operate InstanceOperation) (string, error) {
	accessToken, err := cfutil.GetToken(cliConnection)
	if err != nil {
		return "", err
	}

	eureka, err := serviceInstanceURL(cliConnection, srInstanceName, accessToken, authClient)
	if err != nil {
		return "", fmt.Errorf("Error obtaining service registry URL: %s", err)
	}

	apps, err := getRegisteredAppsWithCfAppName(cliConnection, authClient, accessToken, eureka, cfAppName)
	if err != nil {
		return "", err
	}
	statusTemplate := "Processing service instance %s with index %s\n"
	success := true
	if instanceIndex == nil { //Index is omitted, deregister all instances
		for _, app := range apps {
			fmt.Fprintf(progressWriter, statusTemplate, format.Bold(format.Cyan(app.eurekaAppName)), format.Bold(format.Cyan(app.instanceIndex)))
			err := operate(authClient, eureka, app.eurekaAppName, app.instanceId, accessToken)
			if err != nil {
				success = false
				fmt.Fprintf(progressWriter, "Failed: %s\n", err)
			}
		}
	} else { //Instance ID provided, deregister a single instance
		app, err := getRegisteredAppByInstanceIndex(apps, *instanceIndex)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(progressWriter, statusTemplate, format.Bold(format.Cyan(app.eurekaAppName)), format.Bold(format.Cyan(app.instanceIndex)))
		err = operate(authClient, eureka, app.eurekaAppName, app.instanceId, accessToken)
		if err != nil {
			success = false
			fmt.Fprintf(progressWriter, "Failed: %s\n", err)
		}
	}
	if !success {
		return "", errors.New("Operation failed")
	}
	return "", nil
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
