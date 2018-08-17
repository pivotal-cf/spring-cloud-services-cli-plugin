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

	"io"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/cli"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/config"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/eureka"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/format"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/instance"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/pluginutil"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/serviceutil"
)

// Plugin version. Substitute "<major>.<minor>.<build>" at build time, e.g. using -ldflags='-X main.pluginVersion=1.2.3'
var pluginVersion = "invalid version - plugin was not built correctly"

// Plugin is a struct implementing the Plugin interface, defined by the core CLI, which can
// be found in "code.cloudfoundry.org/cli/plugin/plugin.go".
type Plugin struct{}

func (c *Plugin) Run(cliConnection plugin.CliConnection, args []string) {
	var cfInstanceIndex *int = nil
	var fileToEncrypt string
	var positionalArgs []string
	var err error
	if args[0] == "config-server-encrypt-value" {
		// Enable encryption of a value starting with "-".
		fileToEncrypt, positionalArgs, err = cli.ParseStringFlags(args)
	} else {
		cfInstanceIndex, positionalArgs, err = cli.ParseFlags(args)
	}
	if err != nil {
		format.Diagnose(string(err.Error()), os.Stderr, func() {
			os.Exit(1)
		})
	}

	skipSslValidation, err := cliConnection.IsSSLDisabled()
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

	argsConsumer := cli.NewArgConsumer(positionalArgs, diagnoseWithHelp)

	switch args[0] {

	case "config-server-encrypt-value":
		configServerInstanceName := getConfigServerInstanceName(argsConsumer)
		plainText := getPlainText(argsConsumer)

		if (plainText == "" && fileToEncrypt == "") || (plainText != "" && fileToEncrypt != "") {
			diagnoseWithHelp(fmt.Sprintf("Provide either VALUE_TO_ENCRYPT or the --file-to-encrypt flag, but not both."), "config-server-encrypt-value")
		}

		runActionQuietly(argsConsumer, cliConnection, func() (string, error) {
			return config.Encrypt(cliConnection, configServerInstanceName, plainText, fileToEncrypt, authClient)
		})

	case "config-server-delete-git":
		gitRepoURI := getGitRepoURI(argsConsumer)
		configServerInstanceName := getConfigServerInstanceName(argsConsumer)
		runAction(argsConsumer, cliConnection, fmt.Sprintf("Deleting Git repository %s from config server service instance %s", format.Bold(format.Cyan(gitRepoURI)), format.Bold(format.Cyan(configServerInstanceName))), func(progressWriter io.Writer) (string, error) {
			return config.DeleteGitRepo(cliConnection, authClient, configServerInstanceName, gitRepoURI)
		})

	case "spring-cloud-service-stop":
		serviceInstanceName := getServiceInstanceName(argsConsumer)
		runAction(argsConsumer, cliConnection, fmt.Sprintf("Stopping service instance %s", format.Bold(format.Cyan(serviceInstanceName))), func(progressWriter io.Writer) (string, error) {
			return instance.RunOperation(cliConnection, authClient, serviceInstanceName, instance.Stop)
		})

	case "spring-cloud-service-start":
		serviceInstanceName := getServiceInstanceName(argsConsumer)
		runAction(argsConsumer, cliConnection, fmt.Sprintf("Starting service instance %s", format.Bold(format.Cyan(serviceInstanceName))), func(progressWriter io.Writer) (string, error) {
			return instance.RunOperation(cliConnection, authClient, serviceInstanceName, instance.Start)
		})

	case "spring-cloud-service-restart":
		serviceInstanceName := getServiceInstanceName(argsConsumer)
		runAction(argsConsumer, cliConnection, fmt.Sprintf("Restarting service instance %s", format.Bold(format.Cyan(serviceInstanceName))), func(progressWriter io.Writer) (string, error) {
			return instance.RunOperation(cliConnection, authClient, serviceInstanceName, instance.Restart)
		})

	case "spring-cloud-service-restage":
		serviceInstanceName := getServiceInstanceName(argsConsumer)
		runAction(argsConsumer, cliConnection, fmt.Sprintf("Restaging service instance %s", format.Bold(format.Cyan(serviceInstanceName))), func(progressWriter io.Writer) (string, error) {
			return instance.RunOperation(cliConnection, authClient, serviceInstanceName, instance.Restage)
		})

	case "spring-cloud-service-view":
		serviceInstanceName := getServiceInstanceName(argsConsumer)
		runAction(argsConsumer, cliConnection, fmt.Sprintf("Viewing service instance %s", format.Bold(format.Cyan(serviceInstanceName))), func(progressWriter io.Writer) (string, error) {
			return instance.RunOperation(cliConnection, authClient, serviceInstanceName, instance.View)
		})

	case "spring-cloud-service-configuration":
		serviceInstanceName := getServiceInstanceName(argsConsumer)
		runActionQuietly(argsConsumer, cliConnection, func() (string, error) {
			return instance.RunOperation(cliConnection, authClient, serviceInstanceName, instance.Parameters)
		})

	case "service-registry-enable":
		serviceRegistryInstanceName := getServiceRegistryInstanceName(argsConsumer)
		cfApplicationName := getCfApplicationName(argsConsumer)
		runAction(argsConsumer, cliConnection, fmt.Sprintf("Enabling application %s in service registry %s", format.Bold(format.Cyan(cfApplicationName)), format.Bold(format.Cyan(serviceRegistryInstanceName))), func(progressWriter io.Writer) (string, error) {
			return eureka.OperateOnApplication(cliConnection, serviceRegistryInstanceName, cfApplicationName, authClient, cfInstanceIndex, progressWriter, serviceutil.ServiceInstanceURL, eureka.Enable)
		})

	case "service-registry-deregister":
		serviceRegistryInstanceName := getServiceRegistryInstanceName(argsConsumer)
		cfApplicationName := getCfApplicationName(argsConsumer)
		runAction(argsConsumer, cliConnection, fmt.Sprintf("Deregistering application %s from service registry %s", format.Bold(format.Cyan(cfApplicationName)), format.Bold(format.Cyan(serviceRegistryInstanceName))), func(progressWriter io.Writer) (string, error) {
			return eureka.OperateOnApplication(cliConnection, serviceRegistryInstanceName, cfApplicationName, authClient, cfInstanceIndex, progressWriter, serviceutil.ServiceInstanceURL, eureka.Deregister)
		})

	case "service-registry-disable":
		serviceRegistryInstanceName := getServiceRegistryInstanceName(argsConsumer)
		cfApplicationName := getCfApplicationName(argsConsumer)
		runAction(argsConsumer, cliConnection, fmt.Sprintf("Disabling application %s in service registry %s", format.Bold(format.Cyan(cfApplicationName)), format.Bold(format.Cyan(serviceRegistryInstanceName))), func(progressWriter io.Writer) (string, error) {
			return eureka.OperateOnApplication(cliConnection, serviceRegistryInstanceName, cfApplicationName, authClient, cfInstanceIndex, progressWriter, serviceutil.ServiceInstanceURL, eureka.Disable)
		})

	case "service-registry-info":
		serviceRegistryInstanceName := getServiceRegistryInstanceName(argsConsumer)
		runAction(argsConsumer, cliConnection, fmt.Sprintf("Getting information for service registry %s", format.Bold(format.Cyan(serviceRegistryInstanceName))), func(progressWriter io.Writer) (string, error) {
			return eureka.Info(cliConnection, client, serviceRegistryInstanceName, authClient)
		})

	case "service-registry-list":
		serviceRegistryInstanceName := getServiceRegistryInstanceName(argsConsumer)
		runAction(argsConsumer, cliConnection, fmt.Sprintf("Listing service registry %s", format.Bold(format.Cyan(serviceRegistryInstanceName))), func(progressWriter io.Writer) (string, error) {
			return eureka.List(cliConnection, serviceRegistryInstanceName, authClient)
		})

	default:
		os.Exit(0) // Ignore CLI-MESSAGE-UNINSTALL etc.
	}
}
func getGitRepoURI(ac *cli.ArgConsumer) string {
	return ac.Consume(2, "git repository URI")
}

func getCfApplicationName(ac *cli.ArgConsumer) string {
	return ac.Consume(2, "cf application name")
}

func getConfigServerInstanceName(ac *cli.ArgConsumer) string {
	return ac.Consume(1, "configuration server instance name")
}

func getServiceInstanceName(ac *cli.ArgConsumer) string {
	return ac.Consume(1, "service instance name")
}

func getPlainText(ac *cli.ArgConsumer) string {
	return ac.ConsumeOptional(2, "string to encrypt")
}

func getServiceRegistryInstanceName(ac *cli.ArgConsumer) string {
	return ac.Consume(1, "service registry instance name")
}

func runAction(argsConsumer *cli.ArgConsumer, cliConnection plugin.CliConnection, message string, action func(progressWriter io.Writer) (string, error)) {
	argsConsumer.CheckAllConsumed()

	format.RunAction(cliConnection, message, action, os.Stdout, func() {
		os.Exit(1)
	})
}

func runActionQuietly(argsConsumer *cli.ArgConsumer, cliConnection plugin.CliConnection, action func() (string, error)) {
	argsConsumer.CheckAllConsumed()

	format.RunActionQuietly(cliConnection, action, os.Stdout, func() {
		os.Exit(1)
	})
}

func diagnoseWithHelp(message string, command string) {
	fmt.Printf("%s See 'cf help %s'.\n", message, command)
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
		Name:    "spring-cloud-services",
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
					Usage: `   cf config-server-encrypt-value CONFIG_SERVER_INSTANCE_NAME [VALUE_TO_ENCRYPT]

      NOTE: Either VALUE_TO_ENCRYPT or --file-to-encrypt flag is required, but not both.`,
					Options: map[string]string{"-f/--file-to-encrypt": cli.FileNameUsage},
				},
			},
			{
				Name:     "config-server-delete-git",
				HelpText: "Delete a Git repository from a Spring Cloud Services configuration server service instance",
				Alias:    "csdg",
				UsageDetails: plugin.Usage{
					Usage: "   cf config-server-delete-git CONFIG_SERVER_INSTANCE_NAME GIT_REPO_URI",
				},
			},
			{
				Name:     "spring-cloud-service-stop",
				HelpText: "Stop a Spring Cloud Services service instance",
				Alias:    "scs-stop",
				UsageDetails: plugin.Usage{
					Usage: "   cf scs-stop SERVICE_INSTANCE_NAME",
				},
			},
			{
				Name:     "spring-cloud-service-start",
				HelpText: "Start a Spring Cloud Services service instance",
				Alias:    "scs-start",
				UsageDetails: plugin.Usage{
					Usage: "   cf scs-start SERVICE_INSTANCE_NAME",
				},
			},
			{
				Name:     "spring-cloud-service-restart",
				HelpText: "Restart a Spring Cloud Services service instance",
				Alias:    "scs-restart",
				UsageDetails: plugin.Usage{
					Usage: "   cf scs-restart SERVICE_INSTANCE_NAME",
				},
			},
			{
				Name:     "spring-cloud-service-restage",
				HelpText: "Restage a Spring Cloud Services service instance",
				Alias:    "scs-restage",
				UsageDetails: plugin.Usage{
					Usage: "   cf scs-restage SERVICE_INSTANCE_NAME",
				},
			},
			{
				Name:     "spring-cloud-service-view",
				HelpText: "Display health and status for a Spring Cloud Services service instance",
				Alias:    "scs-view",
				UsageDetails: plugin.Usage{
					Usage: "   cf scs-view SERVICE_INSTANCE_NAME",
				},
			},
			{
				Name:     "spring-cloud-service-configuration",
				HelpText: "Display configuration parameters for a Spring Cloud Services service instance",
				Alias:    "scs-config",
				UsageDetails: plugin.Usage{
					Usage: "   cf scs-config SERVICE_INSTANCE_NAME",
				},
			},
			{
				Name:     "service-registry-deregister",
				HelpText: "Deregister an application registered with a Spring Cloud Services service registry",
				Alias:    "srdr",
				UsageDetails: plugin.Usage{
					Usage:   "   cf service-registry-deregister SERVICE_REGISTRY_INSTANCE_NAME CF_APPLICATION_NAME",
					Options: map[string]string{"-i/--cf-instance-index": cli.CfInstanceIndexUsage},
				},
			},
			{
				Name:     "service-registry-disable",
				HelpText: "Disable an application registered with a Spring Cloud Services service registry so that it is unavailable for traffic",
				Alias:    "srda",
				UsageDetails: plugin.Usage{
					Usage:   "   cf service-registry-disable SERVICE_REGISTRY_INSTANCE_NAME CF_APPLICATION_NAME",
					Options: map[string]string{"-i/--cf-instance-index": cli.CfInstanceIndexUsage},
				},
			},
			{
				Name:     "service-registry-enable",
				HelpText: "Enable an application registered with a Spring Cloud Services service registry so that it is available for traffic",
				Alias:    "sren",
				UsageDetails: plugin.Usage{
					Usage:   "   cf service-registry-enable SERVICE_REGISTRY_INSTANCE_NAME CF_APPLICATION_NAME",
					Options: map[string]string{"-i/--cf-instance-index": cli.CfInstanceIndexUsage},
				},
			},
			{
				Name:     "service-registry-info",
				HelpText: "Display Spring Cloud Services service registry instance information",
				Alias:    "sri",
				UsageDetails: plugin.Usage{
					Usage: "   cf service-registry-info SERVICE_REGISTRY_INSTANCE_NAME",
				},
			},
			{
				Name:     "service-registry-list",
				HelpText: "Display all applications registered with a Spring Cloud Services service registry",
				Alias:    "srl",
				UsageDetails: plugin.Usage{
					Usage: "   cf service-registry-list SERVICE_REGISTRY_INSTANCE_NAME",
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
