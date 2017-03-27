package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
)

func Encrypt(cliConnection plugin.CliConnection, configServerInstanceName string, plainText string, authenticatedClient httpclient.AuthenticatedClient) (string, error) {
	serviceKeyInfo, err := getServiceKeyInfo(cliConnection, configServerInstanceName)
	if err != nil {
		return "", err
	}

	accessToken, err := authenticatedClient.GetClientCredentialsAccessToken(serviceKeyInfo.AccessTokenURI, serviceKeyInfo.ClientId, serviceKeyInfo.ClientSecret)
	if err != nil {
		return "", err
	}

	return encrypt(plainText, serviceKeyInfo.URI, accessToken, authenticatedClient)
}

type serviceKeyInfo struct {
	AccessTokenURI string `json:"access_token_uri"`
	ClientId       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
	URI            string
}

func getServiceKeyInfo(cliConnection plugin.CliConnection, serviceInstance string) (*serviceKeyInfo, error) {
	serviceKeyInfo := &serviceKeyInfo{}
	serviceKey, err := serviceKey(cliConnection, serviceInstance)
	if err != nil {
		return serviceKeyInfo, err
	}
	op, err := cliConnection.CliCommandWithoutTerminalOutput("service-key", serviceInstance, serviceKey)
	if err != nil {
		return serviceKeyInfo, fmt.Errorf("Failed to obtain service key info: %s", err)
	}
	if len(op) < 3 {
		return serviceKeyInfo, fmt.Errorf("Malformed service key info: %v", op)
	}
	serviceKeyInfoJSON := strings.Join(op[2:], "\n")
	err = json.Unmarshal([]byte(serviceKeyInfoJSON), serviceKeyInfo)
	if err != nil {
		return serviceKeyInfo, fmt.Errorf("Failed to unmarshal service key info: %s", err)
	}
	return serviceKeyInfo, nil
}

func serviceKey(cliConnection plugin.CliConnection, serviceInstance string) (string, error) {
	var serviceKey string
	op, err := cliConnection.CliCommandWithoutTerminalOutput("service-keys", serviceInstance)
	if err != nil {
		return "", fmt.Errorf("Service keys not found: %s", err)
	} else if len(op) == 4 && op[3] != "" {
		// The output is in the expected format, so pick out the service key.
		serviceKey = op[3]
	} else {
		// The output is not in the expected format, so default the service key.
		serviceKey = defaultServiceKey(serviceInstance)
	}
	return serviceKey, nil
}

func defaultServiceKey(serviceInstance string) string {
	return serviceInstance + "-key"
}

func encrypt(plainText string, serviceURI string, accessToken string, authenticatedClient httpclient.AuthenticatedClient) (string, error) {
	bodyReader, _, err := authenticatedClient.DoAuthenticatedPost(serviceURI+"/encrypt", "text/plain", plainText, accessToken) // No pun intended between "text/plain" and plainText
	if err != nil {
		return "", err
	}

	defer bodyReader.Close()
	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return "", fmt.Errorf("Failed to read encrypted value: %s", err)
	}

	return string(body), nil
}
