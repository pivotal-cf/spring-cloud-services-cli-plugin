/*
 * Copyright 2016-2017 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package eureka

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
)

type Peer struct {
	Uri               string
	Issuer            string
	SkipSslValidation bool
}

type InfoResp struct {
	NodeCount string
	Peers     []Peer
}

func Info(cliConnection plugin.CliConnection, client httpclient.Client, srInstanceName string, authClient httpclient.AuthenticatedClient) (string, error) {
	return InfoWithResolver(cliConnection, client, srInstanceName, authClient, EurekaUrlFromDashboardUrl)
}

func InfoWithResolver(cliConnection plugin.CliConnection, client httpclient.Client, srInstanceName string, authClient httpclient.AuthenticatedClient,
	eurekaUrlFromDashboardUrl func(dashboardUrl string, accessToken string, authClient httpclient.AuthenticatedClient) (string, error)) (string, error) {
	serviceModel, err := cliConnection.GetService(srInstanceName)
	if err != nil {
		return "", fmt.Errorf("Service registry instance not found: %s", err)
	}
	accessToken, err := cliConnection.AccessToken()
	if err != nil {
		return "", fmt.Errorf("Access token not available: %s", err)
	}
	dashboardUrl := serviceModel.DashboardUrl
	eureka, err := eurekaUrlFromDashboardUrl(dashboardUrl, accessToken, authClient)
	if err != nil {
		return "", fmt.Errorf("Error obtaining service registry dashboard URL: %s", err)
	}

	req, err := http.NewRequest("GET", eureka+"info", nil)
	if err != nil {
		// Should never get here
		return "", fmt.Errorf("Unexpected error: %s", err)
	}
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Service registry error: %s", err)
	}

	buf := new(bytes.Buffer)
	body := resp.Body
	if body == nil {
		return "", errors.New("Invalid service registry response: missing body")
	}
	buf.ReadFrom(resp.Body)

	var infoResp InfoResp
	err = json.Unmarshal(buf.Bytes(), &infoResp)
	if err != nil {
		return "", fmt.Errorf("Invalid service registry response JSON: %s", err)
	}

	return fmt.Sprintf(`Service instance: %s
Server URL: %s
High availability count: %s
Peers: %s
`, srInstanceName, eureka, infoResp.NodeCount, strings.Join(peersToStrings(infoResp.Peers), ", ")), nil
}

type ServiceDefinitionResp struct {
	Credentials struct {
		Uri string
	}
}

func peersToStrings(peers []Peer) []string {
	p := []string{}
	for _, peer := range peers {
		p = append(p, peer.Uri)
	}
	return p
}
