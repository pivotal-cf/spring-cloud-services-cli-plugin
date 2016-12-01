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
)

type Instance struct {
	App      string
	Status   string
	Metadata struct {
		Zone string
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

	req, err := http.NewRequest("GET", eureka+"eureka/apps", nil)
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
	tab.Entitle([]string{"eureka app name", "zone", "status"})
	for _, app := range listResp.Applications.Application {
		instances := app.Instance
		for _, instance := range instances {
			tab.AddRow([]string{instance.App, instance.Metadata.Zone, instance.Status})
		}
	}

	return fmt.Sprintf("Service instance: %s\nServer URL: %s\n\n%s", srInstanceName, eureka, tab.String()), nil
}
