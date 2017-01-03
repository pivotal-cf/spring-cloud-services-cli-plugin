package eureka

import (
	"encoding/json"
	"fmt"
	"strings"

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

type eurekaAppRecord struct {
	cfAppName     string
	eurekaAppName string
	instanceId    string
	status        string
	zone          string
	instanceIndex string
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
	registeredApps, err := getRegisteredApps(cliConnection, authClient, accessToken, eureka, true)

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

func getRegisteredApps(cliConnection plugin.CliConnection, authClient httpclient.AuthenticatedClient, accessToken string, eureka string, asTable bool) ([]eurekaAppRecord, error) {
	registeredApps := []eurekaAppRecord{}
	buf, err := authClient.DoAuthenticatedGet(eureka+"eureka/apps", accessToken)
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
			cfInstanceIndex := metadata.CfInstanceIndex
			var cfAppNm string
			if metadata.CfAppGuid == "" {
				fmt.Printf("cf app GUID not present in metadata of eureka app %s. Perhaps the app was built with an old version of Spring Cloud Services starters.\n", instance.App)
				if asTable {
					cfAppNm = UnknownCfAppName
					cfInstanceIndex = UnknownCfInstanceIndex
				} else {
					continue
				}
			} else {
				cfAppNm, err = cfAppName(cliConnection, metadata.CfAppGuid)
				if err != nil {
					return registeredApps, fmt.Errorf("Failed to determine cf app name corresponding to cf app GUID '%s': %s", metadata.CfAppGuid, err)
				}

			}
			registeredApps = append(registeredApps[:i], eurekaAppRecord{
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
