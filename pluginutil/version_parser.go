package pluginutil

import (
	"strconv"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
)

const numComponents = 3

// ParsePluginVersion parses the given plugin version and return its parsed form. If the given plugin
// version is invalid, calls the given fail function with a suitable message. The fail function will
// typically, except in testing, exit the process or panic.
func ParsePluginVersion(pv string, fail func(format string, inserts ...interface{})) plugin.VersionType {
	v := getPluginVersionComponents(pv, fail)

	return plugin.VersionType{
		Major: v[0],
		Minor: v[1],
		Build: v[2],
	}
}

func getPluginVersionComponents(pv string, fail func(format string, inserts ...interface{})) []int {
	intComponents := make([]int, numComponents)
	components := strings.Split(pv, ".")
	if len(components) != numComponents {
		fail(`pluginVersion %q has invalid format. Expected %d dot-separated integer components.`, pv, numComponents)
		return intComponents
	}

	for c := 0; c < numComponents; c++ {
		v, err := strconv.Atoi(components[c])
		if err != nil {
			fail(`pluginVersion %q has invalid format. Expected integer components.`, pv)
			return intComponents
		}
		intComponents[c] = v
	}
	return intComponents
}
