package eureka

import (
	"fmt"
	"code.cloudfoundry.org/cli/plugin"
)

func Dump(cliConnection plugin.CliConnection, appName string) {
	fmt.Printf("dumping %s\n", appName)
}
