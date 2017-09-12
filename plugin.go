/*
 * Copyright 2016-Present the original author or authors.
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

	"code.cloudfoundry.org/cli/plugin"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/cli"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/config"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/eureka"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/format"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/pluginutil"
)

// Plugin version. Substitute "<major>.<minor>.<build>" at build time, e.g. using -ldflags='-X main.pluginVersion=1.2.3'
var pluginVersion = "invalid version - plugin was not built correctly"

// Plugin is a struct implementing the Plugin interface, defined by the core CLI, which can
// be found in "code.cloudfoundry.org/cli/plugin/plugin.go".
type Plugin struct{}

func (c *Plugin) Run(cliConnection plugin.CliConnection, args []string) {
	skipSslValidation, cfInstanceIndex, positionalArgs, err := cli.ParseFlags(args)
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

	case "config-server-encrypt-value":
		configServerInstanceName := getConfigServerInstanceName(positionalArgs, args[0])
		plainText := getPlainText(positionalArgs, args[0])
		runActionQuietly(cliConnection, func() (string, error) {
			return config.Encrypt(cliConnection, configServerInstanceName, plainText, authClient)
		})

	case "service-registry-enable":
		serviceRegistryInstanceName := getServiceRegistryInstanceName(positionalArgs, args[0])
		cfApplicationName := getCfApplicationName(positionalArgs, args[0])
		runAction(cliConnection, fmt.Sprintf("Enabling application %s in service registry %s", format.Bold(format.Cyan(cfApplicationName)), format.Bold(format.Cyan(serviceRegistryInstanceName))), func() (string, error) {
			return eureka.Enable(cliConnection, serviceRegistryInstanceName, cfApplicationName, authClient, cfInstanceIndex)
		})

	case "service-registry-deregister":
		serviceRegistryInstanceName := getServiceRegistryInstanceName(positionalArgs, args[0])
		cfApplicationName := getCfApplicationName(positionalArgs, args[0])
		runAction(cliConnection, fmt.Sprintf("Deregistering application %s from service registry %s", format.Bold(format.Cyan(cfApplicationName)), format.Bold(format.Cyan(serviceRegistryInstanceName))), func() (string, error) {
			return eureka.Deregister(cliConnection, serviceRegistryInstanceName, cfApplicationName, authClient, cfInstanceIndex)
		})

	case "service-registry-disable":
		serviceRegistryInstanceName := getServiceRegistryInstanceName(positionalArgs, args[0])
		cfApplicationName := getCfApplicationName(positionalArgs, args[0])
		runAction(cliConnection, fmt.Sprintf("Disabling application %s in service registry %s", format.Bold(format.Cyan(cfApplicationName)), format.Bold(format.Cyan(serviceRegistryInstanceName))), func() (string, error) {
			return eureka.Disable(cliConnection, serviceRegistryInstanceName, cfApplicationName, authClient, cfInstanceIndex)
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

func getConfigServerInstanceName(args []string, operation string) string {
	if len(args) < 2 || args[1] == "" {
		diagnoseWithHelp("Configuration server instance name not specified.", operation)
	}
	return args[1]

}

func getPlainText(args []string, operation string) string {
	if len(args) < 3 || args[2] == "" {
		diagnoseWithHelp("string to encrypt not specified.", operation)
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

func runActionQuietly(cliConnection plugin.CliConnection, action func() (string, error)) {
	format.RunActionQuietly(cliConnection, action, os.Stdout, func() {
		os.Exit(1)
	})
}

func diagnoseWithHelp(message string, operation string) {
	fmt.Printf("%s See 'cf help %s.'\n", message, operation)
	os.Exit(1)
}

func failInstallation(format string, inserts ...interface{}) {
	// There is currently no way to emit the message to the command line during plugin installation. Standard output and error are swallowed.
	fmt.Printf(format, inserts...)
	fmt.Println("")

	// Fail the installation
	os.Exit(64)
}

func (c *Plugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name:    "SCSPlugin",
		Version: pluginutil.ParsePluginVersion(pluginVersion, failInstallation),
		MinCliVersion: plugin.VersionType{
			Major: 6,
			Minor: 7,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "config-server-encrypt-value",
				HelpText: "Encrypt a string using a Spring Cloud Services configuration server",
				Alias:    "csev",
				UsageDetails: plugin.Usage{
					Usage:   "   cf config-server-encrypt-value CONFIG_SERVER_INSTANCE_NAME VALUE_TO_ENCRYPT",
					Options: map[string]string{"--skip-ssl-validation": cli.SkipSslValidationUsage},
				},
			},
			{
				Name:     "service-registry-deregister",
				HelpText: "Deregister an application registered with a Spring Cloud Services service registry",
				Alias:    "srdr",
				UsageDetails: plugin.Usage{
					Usage:   "   cf service-registry-deregister SERVICE_REGISTRY_INSTANCE_NAME CF_APPLICATION_NAME",
					Options: map[string]string{"--skip-ssl-validation": cli.SkipSslValidationUsage, "-i/--cf-instance-index": cli.CfInstanceIndexUsage},
				},
			},
			{
				Name:     "service-registry-disable",
				HelpText: "Disable an application registered with a Spring Cloud Services service registry so that it is unavailable for traffic",
				Alias:    "srda",
				UsageDetails: plugin.Usage{
					Usage:   "   cf service-registry-disable SERVICE_REGISTRY_INSTANCE_NAME CF_APPLICATION_NAME",
					Options: map[string]string{"--skip-ssl-validation": cli.SkipSslValidationUsage, "-i/--cf-instance-index": cli.CfInstanceIndexUsage},
				},
			},
			{
				Name:     "service-registry-enable",
				HelpText: "Enable an application registered with a Spring Cloud Services service registry so that it is available for traffic",
				Alias:    "sren",
				UsageDetails: plugin.Usage{
					Usage:   "   cf service-registry-enable SERVICE_REGISTRY_INSTANCE_NAME CF_APPLICATION_NAME",
					Options: map[string]string{"--skip-ssl-validation": cli.SkipSslValidationUsage, "-i/--cf-instance-index": cli.CfInstanceIndexUsage},
				},
			},
			{
				Name:     "service-registry-info",
				HelpText: "Display Spring Cloud Services service registry instance information",
				Alias:    "sri",
				UsageDetails: plugin.Usage{
					Usage:   "   cf service-registry-info SERVICE_REGISTRY_INSTANCE_NAME",
					Options: map[string]string{"--skip-ssl-validation": cli.SkipSslValidationUsage},
				},
			},
			{
				Name:     "service-registry-list",
				HelpText: "Display all applications registered with a Spring Cloud Services service registry",
				Alias:    "srl",
				UsageDetails: plugin.Usage{
					Usage:   "   cf service-registry-list SERVICE_REGISTRY_INSTANCE_NAME",
					Options: map[string]string{"--skip-ssl-validation": cli.SkipSslValidationUsage},
				},
			},
		},
	}
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println("This program is a plugin which expects to be installed into the cf CLI. It is not intended to be run stand-alone.")
		pv := pluginutil.ParsePluginVersion(pluginVersion, failInstallation)
		fmt.Printf("Plugin version: %d.%d.%d\n", pv.Major, pv.Minor, pv.Build)
		os.Exit(0)
	}
	plugin.Start(new(Plugin))
}
