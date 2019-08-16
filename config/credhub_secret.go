package config

import (
	"code.cloudfoundry.org/cli/plugin"
	"errors"
	"fmt"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/cfutil"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/serviceutil"
)

type CredHubSecret interface {
	Add(configServerInstanceName string, credHubPath string, credHubSecret string) error
	Remove(configServerInstanceName string, credHubPath string) error
}

type credHubSecret struct {
	cliConnection              plugin.CliConnection
	authenticatedClient        httpclient.AuthenticatedClient
	serviceInstanceUrlResolver serviceutil.ServiceInstanceResolver
}

func NewCredHubSecret(connection plugin.CliConnection, client httpclient.AuthenticatedClient, resolver serviceutil.ServiceInstanceResolver) CredHubSecret {
	return &credHubSecret{
		cliConnection:              connection,
		authenticatedClient:        client,
		serviceInstanceUrlResolver: resolver,
	}
}

func (r *credHubSecret) Add(configServerInstanceName string, credHubPath string, credHubSecret string) error {
	accessToken, err := cfutil.GetToken(r.cliConnection)
	if err != nil {
		return err
	}

	serviceInstanceUrl, err := r.serviceInstanceUrlResolver.GetServiceInstanceUrl(configServerInstanceName, accessToken)
	if err != nil {
		return fmt.Errorf("error obtaining config server URL: %s", err)
	}

	path := fmt.Sprintf("%ssecrets/%s", serviceInstanceUrl, credHubPath)
	status, e := r.authenticatedClient.DoAuthenticatedPut(path, "application/json", credHubSecret, accessToken)

	if status != 200 {
		return errors.New("failed to add secret or is not supported by this config server")
	}

	if e != nil {
		return e
	}

	return nil
}

func (r *credHubSecret) Remove(configServerInstanceName string, credHubPath string) error {
	accessToken, err := cfutil.GetToken(r.cliConnection)
	if err != nil {
		return err
	}

	serviceInstanceUrl, err := r.serviceInstanceUrlResolver.GetServiceInstanceUrl(configServerInstanceName, accessToken)
	if err != nil {
		return fmt.Errorf("error obtaining config server URL: %s", err)
	}

	path := fmt.Sprintf("%ssecrets/%s", serviceInstanceUrl, credHubPath)
	status, e := r.authenticatedClient.DoAuthenticatedDelete(path, accessToken)

	if status != 200 {
		return errors.New("failed to remove secret or is not supported by this config server")
	}

	if e != nil {
		return e
	}

	return nil
}
