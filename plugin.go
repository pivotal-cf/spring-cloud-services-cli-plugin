/*
 * Copyright 2016-2017 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/plugin"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/eureka"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/format"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
)

const skipSslValidationUsage = "Skip verification of the service registry dashboard endpoint. Not recommended!"
const cfInstanceIndexUsage = "Deregister a specific instance in the Eureka registry. The instance index number can be found by using the the service-registry-list command."
const sslValidationFlagName = "skip-ssl-validation"
const instanceIndexFlagName = "cf-instance-index"

var (
	skipSslValidation bool
	cfInstanceIndex   *int
)

// Plugin is a struct implementing the Plugin interface, defined by the core CLI, which can
// be found in "code.cloudfoundry.org/cli/plugin/plugin.go".
type Plugin struct{}

func (c *Plugin) Run(cliConnection plugin.CliConnection, args []string) {

	positionalArgs, err := parseFlags(args)
	if err != nil {
		format.Diagnose(string(err.Error()), os.Stderr, func() {
			os.Exit(1)
		})
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipSslValidation},
	}
	client := &http.Client{Transport: tr}
	authClient := httpclient.NewAuthenticatedClient(client)

	switch args[0] {

	case "service-registry-deregister":
		serviceRegistryInstanceName := getServiceRegistryInstanceName(positionalArgs, args[0])
		cfApplicationName := getCfApplicationName(positionalArgs, args[0])
		runAction(cliConnection, fmt.Sprintf("Deregistering application %s from service registry %s", format.Bold(format.Cyan(cfApplicationName)), format.Bold(format.Cyan(serviceRegistryInstanceName))), func() (string, error) {
			return eureka.Deregister(cliConnection, serviceRegistryInstanceName, cfApplicationName, authClient, cfInstanceIndex)
		})

	case "service-registry-info":
		serviceRegistryInstanceName := getServiceRegistryInstanceName(positionalArgs, args[0])
		runAction(cliConnection, fmt.Sprintf("Getting information for service registry %s", format.Bold(format.Cyan(serviceRegistryInstanceName))), func() (string, error) {
			return eureka.Info(cliConnection, client, serviceRegistryInstanceName, authClient)
		})

	case "service-registry-list":
		serviceRegistryInstanceName := getServiceRegistryInstanceName(positionalArgs, args[0])
		runAction(cliConnection, fmt.Sprintf("Listing service registry %s", format.Bold(format.Cyan(serviceRegistryInstanceName))), func() (string, error) {
			return eureka.List(cliConnection, serviceRegistryInstanceName, authClient)
		})

	default:
		os.Exit(0) // Ignore CLI-MESSAGE-UNINSTALL etc.

	}
}

func getCfApplicationName(args []string, operation string) string {
	if len(args) < 3 || args[2] == "" {
		diagnoseWithHelp("cf application name not specified.", operation)
	}
	return args[2]
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

func parseFlags(args []string) ([]string, error) {

	fc := flags.New()
	fc.NewBoolFlag(sslValidationFlagName, sslValidationFlagName, skipSslValidationUsage) //name, short_name and usage of the string flag
	fc.NewIntFlag(instanceIndexFlagName, "i", skipSslValidationUsage)                    //name, short_name and usage of the string flag
	err := fc.Parse(args...)
	if err != nil {
		return nil, fmt.Errorf("Error parsing arguments: %s", err)
	}
	skipSslValidation = fc.Bool(sslValidationFlagName)

	if fc.IsSet(instanceIndexFlagName) {
		//Use a pointer instead of value because 0 initialized int is a valid instance index
		var idx int
		idx = fc.Int(instanceIndexFlagName)
		cfInstanceIndex = &idx
	}
	return fc.Args(), nil
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
				Name:     "service-registry-deregister",
				HelpText: "Deregister an application registered with a Spring Cloud Services service registry",
				Alias:    "srd",
				UsageDetails: plugin.Usage{
					Usage:   "   cf service-registry-deregister SERVICE_REGISTRY_INSTANCE_NAME CF_APPLICATION_NAME",
					Options: map[string]string{"--skip-ssl-validation": skipSslValidationUsage, "-i/--cf-instance-index": cfInstanceIndexUsage},
				},
			},
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
