package config

import (
	"fmt"
	"io/ioutil"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/cfutil"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/serviceutil"
)

func Encrypt(cliConnection plugin.CliConnection, configServerInstanceName string, plainText string, authenticatedClient httpclient.AuthenticatedClient) (string, error) {
	return EncryptWithResolver(cliConnection, configServerInstanceName, plainText, authenticatedClient, serviceutil.ServiceInstanceURL)
}

func EncryptWithResolver(cliConnection plugin.CliConnection, configServerInstanceName string, plainText string, authenticatedClient httpclient.AuthenticatedClient,
	serviceInstanceURL func(cliConnection plugin.CliConnection, serviceInstanceName string, accessToken string, authClient httpclient.AuthenticatedClient) (string, error)) (string, error) {
	accessToken, err := cfutil.GetToken(cliConnection)
	if err != nil {
		return "", err
	}

	configServer, err := serviceInstanceURL(cliConnection, configServerInstanceName, accessToken, authenticatedClient)
	if err != nil {
		return "", fmt.Errorf("Error obtaining config server URL: %s", err)
	}

	return encrypt(plainText, configServer, accessToken, authenticatedClient)
}

func encrypt(plainText string, serviceURI string, accessToken string, authenticatedClient httpclient.AuthenticatedClient) (string, error) {
	bodyReader, _, err := authenticatedClient.DoAuthenticatedPost(serviceURI+"encrypt", "text/plain", plainText, accessToken) // No pun intended between "text/plain" and plainText
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
