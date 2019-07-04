package config

import (
	"code.cloudfoundry.org/cli/plugin"
	"errors"
	"fmt"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/cfutil"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/serviceutil"
)

type Refresher interface {
	Refresh(string) error
}

type refresher struct {
	cliConnection              plugin.CliConnection
	authenticatedClient        httpclient.AuthenticatedClient
	serviceInstanceUrlResolver serviceutil.ServiceInstanceUrlResolver
}

func (r *refresher) Refresh(configServerInstanceName string) error {
	accessToken, err := cfutil.GetToken(r.cliConnection)
	if err != nil {
		return err
	}

	serviceInstanceUrl, err := r.serviceInstanceUrlResolver.GetServiceInstanceUrl(configServerInstanceName, accessToken)
	if err != nil {
		return fmt.Errorf("error obtaining config server URL: %s", err)
	}

	_, status, e := r.authenticatedClient.DoAuthenticatedPost(fmt.Sprintf("%sactuator/refreshmirrors", serviceInstanceUrl), "application/json", "", accessToken)
	if e != nil {
		return e
	}

	if status != 200 {
		return errors.New("failed to refresh mirror")
	}

	return nil
}

func NewRefresher(connection plugin.CliConnection, client httpclient.AuthenticatedClient, resolver serviceutil.ServiceInstanceUrlResolver) Refresher {
	return &refresher{
		cliConnection:              connection,
		authenticatedClient:        client,
		serviceInstanceUrlResolver: resolver,
	}
}
