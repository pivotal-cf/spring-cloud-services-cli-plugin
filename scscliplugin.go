package main

import (
	"fmt"

	"bitbucket.org/glyn/scscliplugin/eureka"
	"code.cloudfoundry.org/cli/plugin"
	"os"
)

// SCSPlugin is a struct implementing the Plugin interface, defined by the core CLI, which can
// be found in "code.cloudfoundry.org/cli/plugin/plugin.go".
type SCSPlugin struct{}

func (c *SCSPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] != "service-registry" {
		os.Exit(0) // Ignore CLI-MESSAGE-UNINSTALL etc.
	}

	switch getSubcommand(args) {

	case "dump":
		eureka.Dump(cliConnection, getApplicationName(args))

	default:
		diagnose("Invalid subcommand.")

	}
}

func getSubcommand(args []string) string {
	if len(args) < 2 || args[1] == "" {
		diagnose("Missing subcommand.")
	}
	return args[1]
}

func getApplicationName(args []string) string {
	if len(args) < 3 || args[2] == "" {
		diagnose("Application name missing.")
	}
	return args[2]

}

func diagnose(message string) {
	fmt.Printf("%s See 'cf help service-registry.'\n", message)
	os.Exit(1)
}

func (c *SCSPlugin) GetMetadata() plugin.PluginMetadata {
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
				Name:     "service-registry",
				HelpText: "Manage the SCS service registry",
				Alias:    "sr",
				UsageDetails: plugin.Usage{
					Usage: `Dump service registry as JSON:
   cf service-registry dump APP_NAME

   Print table of service registry applications:
   cf service-registry apps APP_NAME

   Override the status of a bound app's instances:
   cf service-registry override-status APP_NAME STATUS

   where STATUS is one of:
      UP - Ready to receive traffic
      DOWN - Not ready to receive traffic because the application failed a health check
      STARTING - Not ready to receive traffic because application is initializing
      OUT_OF_SERVICE - Intentionally not ready to receive traffic
      UNKNOWN - May or may not  be ready to receive traffic, the true status will be determined automatically

   Delete a status override, optionally setting the status of the bound application's
   instances, or allowing it to be determined automatically and print the affected instance records:
   cf service-registry delete-override APP_NAME [STATUS]
`,
				},
			},
		},
	}
}

func main() {
	plugin.Start(new(SCSPlugin))
}
