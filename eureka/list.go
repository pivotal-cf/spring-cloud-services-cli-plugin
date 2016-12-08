package eureka

import (
	"bytes"
	"code.cloudfoundry.org/cli/plugin"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/format"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
	"net/http"
	"strings"
)

type Instance struct {
	App      string
	Status   string
	Metadata struct {
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

const (
	UnknownCfAppName = "?????"
	UnknownCfInstanceIndex = "?"
)

func List(cliConnection plugin.CliConnection, client httpclient.Client, srInstanceName string) (string, error) {
	serviceModel, err := cliConnection.GetService(srInstanceName)
	if err != nil {
		return "", fmt.Errorf("Service registry instance not found: %s", err)
	}

	dashboardUrl := serviceModel.DashboardUrl
	eureka, err := eurekaFromDashboard(dashboardUrl)
	if err != nil {
		return "", fmt.Errorf("Invalid service registry dashboard URL: %s", err)
	}

	accessToken, err := cliConnection.AccessToken()
	if err != nil {
		return "", fmt.Errorf("Access token not available: %s", err)
	}

	req, err := http.NewRequest("GET", eureka + "eureka/apps", nil)
	if err != nil {
		// Should never get here
		return "", fmt.Errorf("Unexpected error: %s", err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", accessToken)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Service registry unavailable: %s", err)
	}

	buf := new(bytes.Buffer)
	body := resp.Body
	if body == nil {
		return "", errors.New("Invalid service registry response: missing body")
	}
	buf.ReadFrom(resp.Body)

	var listResp ListResp
	err = json.Unmarshal(buf.Bytes(), &listResp)
	if err != nil {
		return "", fmt.Errorf("Invalid service registry response JSON: %s", err)
	}

	tab := &format.Table{}
	tab.Entitle([]string{"eureka app name", "cf app name", "cf instance index", "zone", "status"})
	for _, app := range listResp.Applications.Application {
		instances := app.Instance
		for _, instance := range instances {
			metadata := instance.Metadata
			var cfAppNm string
			cfInstanceIndex := metadata.CfInstanceIndex
			if metadata.CfAppGuid == "" {
				fmt.Printf("cf app GUID not present in metadata of eureka app %s. Perhaps the app was built with an old version of Spring Cloud Services starters.\n", instance.App)
				cfAppNm = UnknownCfAppName
				cfInstanceIndex = UnknownCfInstanceIndex
			} else {
				cfAppNm, err = cfAppName(cliConnection, metadata.CfAppGuid)
				if err != nil {
					return "", fmt.Errorf("Failed to determine cf app name corresponding to cf app GUID '%s': %s", metadata.CfAppGuid, err)
				}
			}
			tab.AddRow([]string{instance.App, cfAppNm, cfInstanceIndex, metadata.Zone, instance.Status})
		}
	}

	return fmt.Sprintf("Service instance: %s\nServer URL: %s\n\n%s", srInstanceName, eureka, tab.String()), nil
}

type SummaryResp struct {
	Name string
}

type SummaryFailure struct {
	Code        int
	Description string
	Error_code  string
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
