/*
 * Copyright (C) 2016-Present Pivotal Software, Inc. All rights reserved.
 *
 * This program and the accompanying materials are made available under
 * the terms of the under the Apache License, Version 2.0 (the "License”);
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
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
		Diagnose(err.Error(), writer, onFailure)
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

func Diagnose(message string, writer io.Writer, onFailure func()) {
	fmt.Fprintf(writer, "%s\n", Bold(Red("FAILED")))

	hint := ""
	if strings.Contains(message, "unknown authority") {
		hint = "Hint: try --skip-ssl-validation at your own risk.\n"
	}

	fmt.Fprintf(writer, "%s\n%s", message, hint)
	onFailure()
}
