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
package cli

import "fmt"
import "code.cloudfoundry.org/cli/cf/flags"

const CfInstanceIndexUsage = "Operate on a specific instance in the Eureka registry. The instance index number can be found by using the service-registry-list command."

func ParseFlags(args []string) (*int, []string, error) {
	const instanceIndexFlagName = "cf-instance-index"

	fc := flags.New()
	//New flag methods take arguments: name, short_name and usage of the string flag
	fc.NewIntFlag(instanceIndexFlagName, "i", CfInstanceIndexUsage)
	err := fc.Parse(args...)
	if err != nil {
		return nil, nil, fmt.Errorf("Error parsing arguments: %s", err)
	}
	//Use a pointer instead of value because 0 initialized int is a valid instance index
	var cfInstanceIndex *int
	if fc.IsSet(instanceIndexFlagName) {
		var idx int
		idx = fc.Int(instanceIndexFlagName)
		cfInstanceIndex = &idx
	}
	return cfInstanceIndex, fc.Args(), nil
}

func ParseNoFlags(args []string) ([]string, error) {
	fc := flags.New()
	fc.SkipFlagParsing(true)
	err := fc.Parse(args...)
	if err != nil {
		return nil, fmt.Errorf("Error parsing arguments: %s", err)
	}
	return fc.Args(), nil
}
