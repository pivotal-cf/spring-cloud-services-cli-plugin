package config

import (
	"code.cloudfoundry.org/cli/plugin"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/cfutil"
	"fmt"
	"net/url"
	"strings"
	"encoding/json"
	"net/http"
)

func DeleteGitRepo(cliConnection plugin.CliConnection, authenticatedClient httpclient.AuthenticatedClient, configServerInstanceName string, gitRepoURI string) (string, error) {
	accessToken, err := cfutil.GetToken(cliConnection)
	if err != nil {
		return "", err
	}

	patchUrl, err := getCliUrlForConfigServerServiceInstance(cliConnection, configServerInstanceName)
	if err != nil {
		return "", err
	}

	bodyMap, _ := json.Marshal(map[string]string{"operation": "delete", "repo": gitRepoURI})
	statusCode, err := authenticatedClient.DoAuthenticatedPatch(patchUrl, "application/json", string(bodyMap), accessToken)
	if err != nil {
		return "", fmt.Errorf("Unable to delete git repo %s from config server service instance %s: %s", gitRepoURI, configServerInstanceName, err)
	}
	if statusCode != http.StatusOK {
		return "", fmt.Errorf("Unable to delete git repo %s from config server service instance %s: %d", gitRepoURI, configServerInstanceName, statusCode)
	}

	return "", nil
}

func getCliUrlForConfigServerServiceInstance(cliConnection plugin.CliConnection, configServerInstanceName string) (string, error) {
	serviceModel, err := cliConnection.GetService(configServerInstanceName)
	if err != nil {
		return "", fmt.Errorf("Config server service instance not found: %s", err)
	}

	parsedUrl, err := url.Parse(serviceModel.DashboardUrl)
	if err != nil {
		return "", err
	}
	path := parsedUrl.Path

	segments := strings.Split(path, "/")
	if len(segments) == 0 || (len(segments) == 1 && segments[0] == "") {
		return "", fmt.Errorf("Unable to determine config server service instance guid (path of %s has no segments)", serviceModel.DashboardUrl)
	}
	guid := segments[len(segments)-1]
	parsedUrl.Path = fmt.Sprintf("/cli/configserver/%s", guid)

	return parsedUrl.String(), nil
}
