package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/eureka"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/format"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
)

// Plugin is a struct implementing the Plugin interface, defined by the core CLI, which can
// be found in "code.cloudfoundry.org/cli/plugin/plugin.go".
type Plugin struct{}

const skipSslValidationUsage = "Skip verification of the service registry dashboard endpoint. Not recommended!"

func (c *Plugin) Run(cliConnection plugin.CliConnection, args []string) {

	skipSslValidation, otherArgs := parseFlags(args)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipSslValidation},
	}
	client := &http.Client{Transport: tr}
	authClient := httpclient.NewAuthenticatedClient(client)

	switch args[0] {

	case "service-registry-info":
		serviceRegistryInstanceName := getServiceRegistryInstanceName(otherArgs, args[0])
		runAction(cliConnection, fmt.Sprintf("Getting information for service registry %s", format.Bold(format.Cyan(serviceRegistryInstanceName))), func() (string, error) {
			return eureka.Info(cliConnection, client, serviceRegistryInstanceName, authClient)
		})

	case "service-registry-list":
		serviceRegistryInstanceName := getServiceRegistryInstanceName(otherArgs, args[0])
		runAction(cliConnection, fmt.Sprintf("Listing service registry %s", format.Bold(format.Cyan(serviceRegistryInstanceName))), func() (string, error) {
			return eureka.List(cliConnection, serviceRegistryInstanceName, authClient)
		})

	default:
		os.Exit(0) // Ignore CLI-MESSAGE-UNINSTALL etc.

	}
}

func getServiceRegistryInstanceName(args []string, operation string) string {
	if len(args) < 2 || args[1] == "" {
		diagnoseWithHelp("Service registry instance name not specified.", operation)
	}
	return args[1]

}

func runAction(cliConnection plugin.CliConnection, message string, action func() (string, error)) {
	format.RunAction(cliConnection, message, action, os.Stdout, func() {
		os.Exit(1)
	})
}

func diagnoseWithHelp(message string, operation string) {
	fmt.Printf("%s See 'cf help %s.'\n", message, operation)
	os.Exit(1)
}

func parseFlags(args []string) (bool, []string) {
	others := []string{}
	found := false
	for _, arg := range args {
		if arg == "--skip-ssl-validation" {
			found = true
		} else {
			others = append(others, arg)
		}
	}
	return found, others
}

func (c *Plugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "SCSPlugin",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 0,
			Build: 1,
		},
		MinCliVersion: plugin.VersionType{
			Major: 6,
			Minor: 7,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "service-registry-info",
				HelpText: "Display Spring Cloud Services service registry instance information",
				Alias:    "sri",
				UsageDetails: plugin.Usage{
					Usage:   "   cf service-registry-info SERVICE_REGISTRY_INSTANCE_NAME",
					Options: map[string]string{"--skip-ssl-validation": skipSslValidationUsage},
				},
			},
			{
				Name:     "service-registry-list",
				HelpText: "Display all applications registered with a Spring Cloud Services service registry",
				Alias:    "srl",
				UsageDetails: plugin.Usage{
					Usage:   "   cf service-registry-list SERVICE_REGISTRY_INSTANCE_NAME",
					Options: map[string]string{"--skip-ssl-validation": skipSslValidationUsage},
				},
			},
		},
	}
}

func main() {
	plugin.Start(new(Plugin))
}
