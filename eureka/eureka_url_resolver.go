package eureka

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
)

func EurekaUrlFromDashboardUrl(dashboardUrl string, accessToken string, authClient httpclient.AuthenticatedClient) (string, error) {
	parsedUrl, err := url.Parse(dashboardUrl)
	if err != nil {
		return "", err
	}
	path := parsedUrl.Path

	segments := strings.Split(path, "/")
	if len(segments) == 0 || (len(segments) == 1 && segments[0] == "") {
		return "", fmt.Errorf("path of %s has no segments", dashboardUrl)
	}
	guid := segments[len(segments)-1]

	parsedUrl.Path = "/cli/instance/" + guid

	buffer, err := authClient.DoAuthenticatedGet(parsedUrl.String(), accessToken)
	var serviceDefinitionResp ServiceDefinitionResp
	if err != nil {
		return "", fmt.Errorf("Invalid service registry definition response: %s", err)
	}
	if buffer == nil {
		return "", errors.New("Buffer is nil")
	}

	err = json.Unmarshal(buffer.Bytes(), &serviceDefinitionResp)
	if err != nil /*TODO: valid JSON with wrong content serviceDefinitionResp.Credentials.Uri != "" */ {
		return "", fmt.Errorf("Invalid service registry definition response JSON: %s, response body: '%s'", err, string(buffer.Bytes()))
	}

	return serviceDefinitionResp.Credentials.Uri + "/", nil
}
