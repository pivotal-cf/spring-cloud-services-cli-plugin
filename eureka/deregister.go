package eureka

import (
	"code.cloudfoundry.org/cli/plugin"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
	"fmt"
	"encoding/json"
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
	fmt.Printf("App to be deregistered: %s\n", apps)

	for _, app := range apps {
		err = deregister(authClient, accessToken, eureka, app.eurekaAppName, app.instanceId)
	}

	return "", nil
}

func deregister(authClient httpclient.AuthenticatedClient, accessToken string, eureka string, eurekaAppName string, instanceId string) error {
	return authClient.DoAuthenticatedDelete(eureka+fmt.Sprintf("eureka/apps/%s/%s", eurekaAppName, instanceId), accessToken)
}

type eurekaAppRecord struct {
	cfAppName string
	eurekaAppName string
	instanceId string
}

func getRegisteredAppsWithCfAppName(cliConnection plugin.CliConnection,authClient httpclient.AuthenticatedClient, accessToken string, eureka string, cfAppName string) ([]eurekaAppRecord, error) {
	registeredAppsWithCfAppName := []eurekaAppRecord{}

	registeredApps, err := getRegisteredApps(cliConnection, authClient, accessToken, eureka)
	if err != nil {
		return registeredAppsWithCfAppName, err
	}

	for _, app := range registeredApps {
		fmt.Printf("cfAppName is: %s \nregistered app struct is: %s\n", cfAppName, app)
		if app.cfAppName == cfAppName {
			registeredAppsWithCfAppName = append(registeredAppsWithCfAppName, app)
		}
	}

	return registeredAppsWithCfAppName, nil
}

func getRegisteredApps(cliConnection plugin.CliConnection, authClient httpclient.AuthenticatedClient, accessToken string, eureka string) ([]eurekaAppRecord, error) {
	registeredApps := []eurekaAppRecord{}
	buf, err := authClient.DoAuthenticatedGet(eureka+"eureka/apps", accessToken)
	if err != nil {
		return registeredApps, fmt.Errorf("Service registry error: %s", err)
	}

	var listResp ListResp
	err = json.Unmarshal(buf.Bytes(), &listResp)
	if err != nil {
		return registeredApps, fmt.Errorf("Invalid service registry response JSON: %s, response body: '%s'", err, string(buf.Bytes()))
	}

	apps := listResp.Applications.Application
	for _, app := range apps {
		instances := app.Instance
		for _, instance := range instances {
			metadata := instance.Metadata
			var cfAppNm string
			if metadata.CfAppGuid == "" {
				fmt.Printf("cf app GUID not present in metadata of eureka app %s. Perhaps the app was built with an old version of Spring Cloud Services starters.\n", instance.App)
				return registeredApps, fmt.Errorf("add details later")
			} else {
				cfAppNm, err = cfAppName(cliConnection, metadata.CfAppGuid)
				if err != nil {
					return registeredApps, fmt.Errorf("Failed to determine cf app name corresponding to cf app GUID '%s': %s", metadata.CfAppGuid, err)
				}
			}
			fmt.Printf("CF App name is: %s", cfAppNm)
			registeredApps = append(registeredApps, eurekaAppRecord{
				cfAppName: cfAppNm,
				eurekaAppName: instance.App,
				instanceId: instance.InstanceId,
			})
			fmt.Printf("Registered apps: %s", registeredApps)
		}
		return registeredApps, nil
	}

	return []eurekaAppRecord{}, nil
}


