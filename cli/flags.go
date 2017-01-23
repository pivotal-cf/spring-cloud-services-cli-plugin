package cli

import "fmt"
import "code.cloudfoundry.org/cli/cf/flags"

const (
	SkipSslValidationUsage = "Skip verification of the service registry dashboard endpoint. Not recommended!"
	CfInstanceIndexUsage   = "Deregister a specific instance in the Eureka registry. The instance index number can be found by using the the service-registry-list command."
)

func ParseFlags(args []string) (bool, *int, []string, error) {
	const (
		sslValidationFlagName = "skip-ssl-validation"
		instanceIndexFlagName = "cf-instance-index"
	)

	fc := flags.New()
	//New flag methods take arguments: name, short_name and usage of the string flag
	fc.NewBoolFlag(sslValidationFlagName, sslValidationFlagName, SkipSslValidationUsage)
	fc.NewIntFlag(instanceIndexFlagName, "i", CfInstanceIndexUsage)
	err := fc.Parse(args...)
	if err != nil {
		return false, nil, nil, fmt.Errorf("Error parsing arguments: %s", err)
	}
	skipSslValidation := fc.Bool(sslValidationFlagName)
	//Use a pointer instead of value because 0 initialized int is a valid instance index
	var cfInstanceIndex *int
	if fc.IsSet(instanceIndexFlagName) {
		var idx int
		idx = fc.Int(instanceIndexFlagName)
		cfInstanceIndex = &idx
	}
	return skipSslValidation, cfInstanceIndex, fc.Args(), nil
}
