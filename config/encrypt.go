package config

import (
	"fmt"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/serviceutil"
	"io/ioutil"
	"net/http"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/cfutil"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
)

type Encrypter interface {
	EncryptString(configServerInstanceName string, plainText string) (string, error)
	EncryptFile(configServerInstanceName string, fileToEncrypt string) (string, error)
}

type encrypter struct {
	cliConnection              plugin.CliConnection
	authenticatedClient        httpclient.AuthenticatedClient
	serviceInstanceUrlResolver serviceutil.ServiceInstanceResolver
}

func NewEncrypter(cliConnection plugin.CliConnection, authenticatedClient httpclient.AuthenticatedClient, serviceInstanceUrlResolver serviceutil.ServiceInstanceResolver) Encrypter {
	return &encrypter{
		cliConnection:              cliConnection,
		authenticatedClient:        authenticatedClient,
		serviceInstanceUrlResolver: serviceInstanceUrlResolver,
	}
}

func (e *encrypter) EncryptFile(configServerInstanceName string, fileToEncrypt string) (string, error) {
	textToEncrypt, err := ReadFileContents(fileToEncrypt)
	if err != nil {
		return "", err
	}

	return e.EncryptString(configServerInstanceName, textToEncrypt)
}

func (e *encrypter) EncryptString(configServerInstanceName string, textToEncrypt string) (string, error) {
	accessToken, err := cfutil.GetToken(e.cliConnection)
	if err != nil {
		return "", err
	}

	configServerUrl, err := e.serviceInstanceUrlResolver.GetServiceInstanceUrl(configServerInstanceName, accessToken)
	if err != nil {
		return "", fmt.Errorf("Error obtaining config server URL: %s", err)
	}

	var bodyHoldsErrorDetails = false
	bodyReader, statusCode, err := e.authenticatedClient.DoAuthenticatedPost(configServerUrl+"encrypt", "text/plain", textToEncrypt, accessToken)
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

	if bodyHoldsErrorDetails || statusCode != http.StatusOK {
		errorDetails := ""
		if len(body) > 0 {
			errorDetails = fmt.Sprintf(": %s", string(body))
		}
		return "", fmt.Errorf("Encryption failed or is not supported by this config server%s", errorDetails)
	}

	return string(body), nil
}
