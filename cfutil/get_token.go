/*
 * Copyright (C) 2017-Present Pivotal Software, Inc. All rights reserved.
 *
 * This program and the accompanying materials are made available under
 * the terms of the under the Apache License, Version 2.0 (the "License”);
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package cfutil

import (
	"fmt"

	"strings"

	"code.cloudfoundry.org/cli/plugin"
)

func GetToken(cliConnection plugin.CliConnection) (string, error) {
	output, err := cliConnection.AccessToken()
	if err != nil {
		return "", fmt.Errorf("Access token not available: %s", err)
	}

	parsedOutput := strings.Split(strings.Trim(output, "\n"), " ")
	if len(parsedOutput) != 2 || parsedOutput[0] != "bearer" || parsedOutput[1] == "" {
		return "", fmt.Errorf("Access token output invalid: %s", output)
	}

	return parsedOutput[1], nil
}
