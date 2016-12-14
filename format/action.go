package format

import (
	"fmt"
	"io"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/fatih/color"
)

var (
	Bold  func(format string, a ...interface{}) string = color.New(color.Bold).SprintfFunc()
	Cyan  func(format string, a ...interface{}) string = color.New(color.FgHiCyan).SprintfFunc()
	Green func(format string, a ...interface{}) string = color.New(color.FgGreen).SprintfFunc()
	Red   func(format string, a ...interface{}) string = color.New(color.FgRed).SprintfFunc()
)

// Run a given action with a given progress message, writing the output to the given writer and invoking a failure closure if an error occurs.
func RunAction(cliConnection plugin.CliConnection, message string, action func() (string, error), writer io.Writer, onFailure func()) {
	printStartAction(cliConnection, message, writer)
	output, err := action()
	if err != nil {
		diagnose(err.Error(), writer, onFailure)
		return
	}
	fmt.Fprintf(writer, "%s\n\n%s", Bold(Green("OK")), output)
}

func printStartAction(cliConnection plugin.CliConnection, message string, writer io.Writer) {
	orgModel, err := cliConnection.GetCurrentOrg()
	if err != nil {
		return
	}

	spaceModel, err := cliConnection.GetCurrentSpace()
	if err != nil {
		return
	}

	user, err := cliConnection.Username()
	if err != nil || user == "" {
		return
	}

	fmt.Fprintf(writer, "%s in org %s / space %s as %s...\n", message, Bold(Cyan(orgModel.Name)), Bold(Cyan(spaceModel.Name)), Bold(Cyan(user)))
}

func diagnose(message string, writer io.Writer, onFailure func()) {
	fmt.Fprintf(writer, "%s\n", Bold(Red("FAILED")))

	hint := ""
	if strings.Contains(message, "unknown authority") {
		hint = "Hint: try --skip-ssl-validation at your own risk.\n"
	}

	fmt.Fprintf(writer, "%s\n%s", message, hint)
	onFailure()
}
