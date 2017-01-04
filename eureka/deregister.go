package eureka

import (
	"fmt"

	"errors"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
)

func Deregister(cliConnection plugin.CliConnection, srInstanceName string, cfAppName string, authenticatedClient httpclient.AuthenticatedClient) (string, error) {
	return DeregisterWithResolver(cliConnection, srInstanceName, cfAppName, authenticatedClient, EurekaUrlFromDashboardUrl)
}

func DeregisterWithResolver(cliConnection plugin.CliConnection, srInstanceName string, cfAppName string, authClient httpclient.AuthenticatedClient,
	eurekaUrlFromDashboardUrl func(dashboardUrl string, accessToken string, authClient httpclient.AuthenticatedClient) (string, error)) (string, error) {
	serviceModel, err := cliConnection.GetService(srInstanceName)
	if err != nil {
		return "", fmt.Errorf("Service registry instance not found: %s", err)
	}
	accessToken, err := cliConnection.AccessToken()
	if err != nil {
		return "", fmt.Errorf("Access token not available: %s", err)
	}

	eureka, err := eurekaUrlFromDashboardUrl(serviceModel.DashboardUrl, accessToken, authClient)
	if err != nil {
		return "", fmt.Errorf("Error obtaining service registry dashboard URL: %s", err)
	}

	apps, err := getRegisteredAppsWithCfAppName(cliConnection, authClient, accessToken, eureka, cfAppName)
	if err != nil {
		return "", err
	}

	for _, app := range apps {
		err = deregister(authClient, accessToken, eureka, app.eurekaAppName, app.instanceId)
	}

	return "", nil
}

func deregister(authClient httpclient.AuthenticatedClient, accessToken string, eureka string, eurekaAppName string, instanceId string) error {
	return authClient.DoAuthenticatedDelete(eureka+fmt.Sprintf("eureka/apps/%s/%s", eurekaAppName, instanceId), accessToken)
}

func getRegisteredAppsWithCfAppName(cliConnection plugin.CliConnection, authClient httpclient.AuthenticatedClient, accessToken string, eureka string, cfAppName string) ([]eurekaAppRecord, error) {
	registeredAppsWithCfAppName := []eurekaAppRecord{}

	registeredApps, err := getRegisteredApps(cliConnection, authClient, accessToken, eureka)

	if err != nil {
		return registeredAppsWithCfAppName, err
	}

	for _, app := range registeredApps {
		if app.cfAppName == cfAppName {
			registeredAppsWithCfAppName = append(registeredAppsWithCfAppName, app)
		}
	}
	if len(registeredAppsWithCfAppName) == 0 {
		return registeredAppsWithCfAppName, errors.New(fmt.Sprintf("cf app name %s not found", cfAppName))
	}

	return registeredAppsWithCfAppName, nil
}
