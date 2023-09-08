package config

import (
	"code.cloudfoundry.org/cli/plugin"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/cfutil"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/serviceutil"
	"io"
	"io/ioutil"
)

type Refresher interface {
	Refresh(string) error
}

type refresher struct {
	cliConnection              plugin.CliConnection
	authenticatedClient        httpclient.AuthenticatedClient
	serviceInstanceUrlResolver serviceutil.ServiceInstanceResolver
}

func checkRefreshStatus(reader io.Reader) error {
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	var jsonMap map[string]map[string]string
	err = json.Unmarshal(body, &jsonMap)
	if err != nil {
		return err
	}

	for _, refresh := range jsonMap {
		status := refresh["status"]
		if status == "FAILED" {
			return errors.New("failed to refresh mirror")
		}
	}
	return nil
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

	bodyReader, status, e := r.authenticatedClient.DoAuthenticatedPost(fmt.Sprintf("%sactuator/refreshmirrors", serviceInstanceUrl), "application/json", "", accessToken)
	defer bodyReader.Close()

	if e != nil {
		return e
	}

	if status != 200 {
		return errors.New("failed to refresh mirror")
	}

	return checkRefreshStatus(bodyReader)
}

func NewRefresher(connection plugin.CliConnection, client httpclient.AuthenticatedClient, resolver serviceutil.ServiceInstanceResolver) Refresher {
	return &refresher{
		cliConnection:              connection,
		authenticatedClient:        client,
		serviceInstanceUrlResolver: resolver,
	}
}
