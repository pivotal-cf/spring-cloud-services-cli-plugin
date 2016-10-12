package eureka

import (
	"fmt"
	"code.cloudfoundry.org/cli/plugin"
	"os"
	"strings"
	"github.com/elgs/gojq"
	"os/exec"
	"encoding/json"
	"bytes"
)

func Dump(cliConnection plugin.CliConnection, appName string) {
	model, err := cliConnection.GetApp(appName)
	if err != nil {
		handleError("Application not found", err)
	}
	appGuid := model.Guid;

	output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v2/apps/%s/env", appGuid))
	if err != nil {
		handleError("Application environment error", err)
	}
	envJson := strings.Join(output, "")

	parser, err := gojq.NewStringQuery(envJson)
	if err != nil {
		handleError("Failed to parse application environment JSON", err)
	}
	creds, err := parser.Query("system_env_json.VCAP_SERVICES.p-service-registry.[0].credentials")
	if err != nil {
		handleError("Failed to find credentials in application environment JSON", err)
	}

	credentials, ok := creds.(map[string]interface{})
	if !ok {
		handleError("Credentials have wrong type", nil)
	}
	accessTokenUri := getString(credentials, "access_token_uri")
	clientId := getString(credentials, "client_id")
	clientSecret := getString(credentials, "client_secret")
	eurekaUri := getString(credentials, "uri")

	// FIXME: replace the following with a proper HTTP request. Using curl for spike only.
	cmd := exec.Command("curl", "-s" ,"-k", accessTokenUri, "-u", clientId + ":" + clientSecret, "-d", "grant_type=client_credentials")
	op, err := cmd.Output()
	if err != nil {
		handleError("curl failed", err)
	}

	parser, err = gojq.NewStringQuery(string(op))
	if err != nil {
		handleError("Failed to parse curl response JSON", err)
	}
	accessToken, err := parser.Query("access_token")
	if err != nil {
		handleError("Failed to find access token in curl response JSON", err)
	}
	registry := request(eurekaUri, accessToken.(string), "GET", "/eureka/apps")

	var pretty bytes.Buffer
	err = json.Indent(&pretty, []byte(registry), "", "  ")
	if err != nil {
		handleError("json.Indent failed", err)
	}
	fmt.Println(string(pretty.Bytes()))
}

func request(uri string, token string, method string, path string) string {
	// FIXME: replace the following with a proper HTTP request. Using curl for spike only.
	cmd := exec.Command("curl", "-s" ,"-k", "-H", fmt.Sprintf("Authorization: bearer %s", token), "-H", "Accept: application/json", "-X", method, uri + path)
	op, err := cmd.Output()
	if err != nil {
		handleError("curl failed", err)
	}
	return string(op)
}

func getString(credentials map[string]interface{}, key string) string {
	value, ok := credentials[key].(string)
	if !ok {
		handleError(fmt.Sprintf("%s has wrong type", key), nil)
	}
	return value
}

func handleError(message string, err error) {
	fmt.Printf("%s: %s\n", message, err)
	os.Exit(1)
}