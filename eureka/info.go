package eureka

import (
	"bytes"
	"code.cloudfoundry.org/cli/plugin"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pivotal-cf/spring-cloud-services-cli-plugin/httpclient"
	"net/http"
	"net/url"
	"strings"
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

func Info(cliConnection plugin.CliConnection, client httpclient.Client, srInstanceName string) (string, error) {
	serviceModel, err := cliConnection.GetService(srInstanceName)
	if err != nil {
		return "", fmt.Errorf("Service registry instance not found: %s", err)
	}

	dashboardUrl := serviceModel.DashboardUrl
	eureka, err := eurekaFromDashboard(dashboardUrl)
	if err != nil {
		return "", fmt.Errorf("Invalid service registry dashboard URL: %s", err)
	}

	req, err := http.NewRequest("GET", eureka+"info", nil)
	if err != nil {
		// Should never get here
		return "", fmt.Errorf("Unexpected error: %s", err)
	}
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Service registry unavailable: %s", err)
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

func eurekaFromDashboard(dashboardUrl string) (string, error) {
	url, err := url.Parse(dashboardUrl)
	if err != nil {
		return "", err
	}
	hostname, path := url.Host, url.Path

	labels := strings.Split(hostname, ".")
	if len(labels) < 2 {
		return "", fmt.Errorf("hostname of %s has less than two labels", dashboardUrl)
	}

	segments := strings.Split(path, "/")
	if len(segments) == 0 || (len(segments) == 1 && segments[0] == "") {
		return "", fmt.Errorf("path of %s has no segments", dashboardUrl)
	}
	guid := segments[len(segments)-1]

	url.Host = "eureka-" + guid + "." + strings.Join(labels[1:], ".")
	url.Path = "/"
	return url.String(), nil
}

func peersToStrings(peers []Peer) []string {
	p := []string{}
	for _, peer := range peers {
		p = append(p, peer.Uri)
	}
	return p
}
