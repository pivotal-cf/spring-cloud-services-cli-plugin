package config

import (
	"fmt"
	"io/ioutil"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/cfutil"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/serviceutil"
)

var DefaultResolver = serviceutil.ServiceInstanceURL

func Encrypt(cliConnection plugin.CliConnection, configServerInstanceName string, plainText string, fileToEncrypt string, authenticatedClient httpclient.AuthenticatedClient) (string, error) {
	textToEncrypt := plainText
	var err error

	if fileToEncrypt != "" {
		textToEncrypt, err = ReadFileContents(fileToEncrypt)
		if err != nil {
			return "", err
		}
	}
	return EncryptWithResolver(cliConnection, configServerInstanceName, textToEncrypt, authenticatedClient, DefaultResolver)
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
	var bodyHoldsErrorDetails = false
	bodyReader, _, err := authenticatedClient.DoAuthenticatedPost(serviceURI+"encrypt", "text/plain", plainText, accessToken) // No pun intended between "text/plain" and plainText
	if err != nil {
		if bodyReader == nil {
			return "", err
		}
		bodyHoldsErrorDetails = true
	}

	defer bodyReader.Close()
	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return "", fmt.Errorf("Failed to read encrypted value: %s", err)
	}

	if bodyHoldsErrorDetails {
		return "", fmt.Errorf("Encryption failed: %v", string(body))
	}

	return string(body), nil
}
