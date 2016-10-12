package main

import (
	"fmt"

	"code.cloudfoundry.org/cli/plugin"
)

// SCSPlugin is the struct implementing the interface defined by the core CLI. It can
// be found at  "code.cloudfoundry.org/cli/plugin/plugin.go"
type SCSPlugin struct{}

// Run must be implemented by any plugin because it is part of the
// plugin interface defined by the core CLI.
//
// Run(....) is the entry point when the core CLI is invoking a command defined
// by a plugin. The first parameter, plugin.CliConnection, is a struct that can
// be used to invoke cli commands. The second paramter, args, is a slice of
// strings. args[0] will be the name of the command, and will be followed by
// any additional arguments a cli user typed in.
//
// Any error handling should be handled with the plugin itself (this means printing
// user facing errors). The CLI will exit 0 if the plugin exits 0 and will exit
// 1 should the plugin exits nonzero.
func (c *SCSPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	// Ensure that we called the command service-registry
	if args[0] == "service-registry" {
		fmt.Println("Running the service-registry command")
	}
}

// GetMetadata must be implemented as part of the plugin interface
// defined by the core CLI.
//
// GetMetadata() returns a PluginMetadata struct. The first field, Name,
// determines the name of the plugin which should generally be without spaces.
// If there are spaces in the name a user will need to properly quote the name
// during uninstall otherwise the name will be treated as seperate arguments.
// The second value is a slice of Command structs. Our slice only contains one
// Command Struct, but could contain any number of them. The first field Name
// defines the command `cf service-registry` once installed into the CLI. The
// second field, HelpText, is used by the core CLI to display help information
// to the user in the core commands `cf help`, `cf`, or `cf -h`.
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
					Usage:
`Dump service registry as JSON:
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
   instances, or allowing it to be determined automatically:
   cf service-registry delete-override APP_NAME [STATUS]

   The affected instance records are printed.
`,
				},
			},
		},
	}
}

// Unlike most Go programs, the `Main()` function will not be used to run all of the
// commands provided in your plugin. Main will be used to initialize the plugin
// process, as well as any dependencies you might require for your
// plugin.
func main() {
	// Any initialization for your plugin can be handled here
	//
	// Note: to run the plugin.Start method, we pass in a pointer to the struct
	// implementing the interface defined at "code.cloudfoundry.org/cli/plugin/plugin.go"
	//
	// Note: The plugin's main() method is invoked at install time to collect
	// metadata. The plugin will exit 0 and the Run([]string) method will not be
	// invoked.
	plugin.Start(new(SCSPlugin))
	// Plugin code should be written in the Run([]string) method,
	// ensuring the plugin environment is bootstrapped.
}