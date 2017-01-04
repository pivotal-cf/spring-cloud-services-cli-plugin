package eureka

import (
	"encoding/json"
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
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
	buf, err := authClient.DoAuthenticatedGet(eurekaUrl+"eureka/apps", accessToken)
	if err != nil {
		return registeredApps, fmt.Errorf("Service registry error: %s", err)
	}

	var listResp ListResp
	err = json.Unmarshal(buf.Bytes(), &listResp)
	if err != nil {
		return registeredApps, fmt.Errorf("Invalid service registry response JSON: %s, response body: '%s'", err, string(buf.Bytes()))
	}

	apps := listResp.Applications.Application
	for i, app := range apps {
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
				cfAppNm, err = cfAppName(cliConnection, metadata.CfAppGuid)
				if err != nil {
					return registeredApps, fmt.Errorf("Failed to determine cf app name corresponding to cf app GUID '%s': %s", metadata.CfAppGuid, err)
				}

			}
			registeredApps = append(registeredApps[:i], eurekaAppRecord{
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

func cfAppName(cliConnection plugin.CliConnection, cfAppGuid string) (string, error) {
	output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v2/apps/%s/summary", cfAppGuid), "-H", "Accept: application/json")
	if err != nil {
		return "", err
	}

	// Cope with some errors coming back with err == nil.
	// See https://www.pivotaltracker.com/story/show/130060949 for a potential alternative.
	err = diagnoseCurlError(output)
	if err != nil {
		return "", err
	}

	var summaryResp SummaryResp
	err = json.Unmarshal([]byte(strings.Join(output, "\n")), &summaryResp)
	if err != nil {
		return "", err
	}

	return summaryResp.Name, err
}

func diagnoseCurlError(output []string) error {
	var summaryFailure SummaryFailure
	err := json.Unmarshal([]byte(strings.Join(output, "\n")), &summaryFailure)
	if err == nil && summaryFailure.Code != 0 {
		return fmt.Errorf("%s: code %d, error_code %s", summaryFailure.Description, summaryFailure.Code, summaryFailure.Error_code)
	}
	return nil
}
