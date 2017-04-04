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

	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/cli/plugin/models"
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
