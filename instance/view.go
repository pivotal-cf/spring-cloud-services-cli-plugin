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
package instance

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"strings"

	"time"

	"code.cloudfoundry.org/bytefmt"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
)

type ViewInstanceResp struct {
	BackingApps []BackingApp `json:"backing_apps"`
}

type BackingApp struct {
	Name             string
	Buildpack        string
	LastUploaded     int64 `json:"last_uploaded"`
	Stack            string
	Memory           int
	NumInstances     int    `json:"num_instances"`
	RunningInstances int    `json:"running_instances"`
	RequestedState   string `json:"requested_state"`
	Routes           []string
	Instances        []BackingAppInstance
}

type BackingAppInstance struct {
	Index       int
	State       string
	Since       int64
	CPU         float64
	MemoryUsage int64 `json:"memory_usage"`
	MemoryQuota int64 `json:"memory_quota"`
	DiskUsage   int64 `json:"disk_usage"`
	DiskQuota   int64 `json:"disk_quota"`
	Details     string
}

type viewOperation struct{}

func (so *viewOperation) Run(authenticatedClient httpclient.AuthenticatedClient, serviceInstanceAdminURL string, accessToken string) (string, error) {
	bodyReader, statusCode, err := authenticatedClient.DoAuthenticatedGet(serviceInstanceAdminURL, accessToken)
	if err != nil {
		return "", err
	}
	if statusCode != http.StatusOK {
		return "", fmt.Errorf("Service broker view instance failed: %d", statusCode)
	}

	if bodyReader == nil {
		return "", errors.New("Service broker view instance response body missing")
	}
	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return "", fmt.Errorf("Cannot read service broker view instance response body: %s", err)
	}

	var viewInstanceResp ViewInstanceResp
	err = json.Unmarshal(body, &viewInstanceResp)
	if err != nil {
		return "", fmt.Errorf("Invalid service broker view instance response JSON: %s, response body: '%s'", err, string(body))
	}

	return RenderView(&viewInstanceResp)
}

func (so *viewOperation) IsLifecycleOperation() bool {
	return false
}

func NewViewOperation() Operation {
	return &viewOperation{}
}

func RenderView(viewInstanceResp *ViewInstanceResp) (string, error) {
	const maxWidth = 150

	var buffer bytes.Buffer
	for _, backingApp := range viewInstanceResp.BackingApps {
		buffer.WriteString(fmt.Sprintf(`
backing app name: %s
requested state:  %s
instances:        %d/%d
usage:            %s x %d instances
routes:           %s
last uploaded:    %s
stack:            %s
buildpack:        %s

     state     since                  cpu    memory       disk           details
`,
			backingApp.Name,
			strings.ToLower(backingApp.RequestedState),
			backingApp.RunningInstances, backingApp.NumInstances,
			byteSize(int64(backingApp.Memory*bytefmt.MEGABYTE)), backingApp.NumInstances,
			strings.Join(backingApp.Routes, ", "),
			iso8601LocalDate(backingApp.LastUploaded),
			backingApp.Stack,
			justifyBuildpack(backingApp.Buildpack, maxWidth, len("buildpack:        "))))

		for _, backingAI := range backingApp.Instances {
			buffer.WriteString(fmt.Sprintf(`#%-3d %-9s %-22s %-6s %-12s %-14s %s
`,
				backingAI.Index,
				strings.ToLower(backingAI.State),
				rfc3339UtcDate(backingAI.Since),
				fmt.Sprintf("%.1f%%", 100*backingAI.CPU),
				fmt.Sprintf("%s of %s", byteSize(backingAI.MemoryUsage), byteSize(backingAI.MemoryQuota)),
				fmt.Sprintf("%s of %s", byteSize(backingAI.DiskUsage), byteSize(backingAI.DiskQuota)),
				backingAI.Details))
		}
	}
	return buffer.String(), nil
}

func iso8601LocalDate(input int64) string {
	return toTime(input).Local().Format("Mon 02 Jan 15:04:05 MST 2006")
}

func rfc3339UtcDate(input int64) string {
	return toTime(input).UTC().Format(time.RFC3339)
}

func toTime(input int64) time.Time {
	const millis int64 = 1000
	return time.Unix(input/millis, 0)
}

func justifyBuildpack(buildpack string, width int, indent int) string {
	if len(buildpack) > width {
		return buildpack[:width] + "\n" + strings.Repeat(" ", indent) + justifyBuildpack(buildpack[width:], width, indent)
	}
	return buildpack
}

func byteSize(size int64) string {
	return bytefmt.ByteSize(uint64(size))
}
