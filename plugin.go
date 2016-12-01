package main

import (
	"fmt"

	"code.cloudfoundry.org/cli/plugin"
	"crypto/tls"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/eureka"
	"net/http"
	"os"
	"strings"
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

	switch args[0] {

	case "service-registry-info":
		info, err := eureka.Info(cliConnection, client, getServiceRegistryInstanceName(otherArgs, "service-registry-info"))
		if err != nil {
			diagnose(err.Error())
		}
		fmt.Printf(info)

	case "service-registry-list":
		list, err := eureka.List(cliConnection, client, getServiceRegistryInstanceName(otherArgs, "service-registry-info"))
		if err != nil {
			diagnose(err.Error())
		}
		fmt.Printf(list)

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

func diagnose(message string) {
	hint := ""
	if strings.Contains(message, "unknown authority") {
		hint = "Hint: try --skip-ssl-validation at your own risk.\n"
	}

	fmt.Printf("%s\n%s", message, hint)
	os.Exit(1)
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
