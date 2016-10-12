package eureka

import (
	"fmt"
	"code.cloudfoundry.org/cli/plugin"
	"strings"
	"github.com/elgs/gojq"
	"os/exec"
	"encoding/json"
	"bytes"
)

func Dump(cliConnection plugin.CliConnection, appName string) (string, error) {
	model, err := cliConnection.GetApp(appName)
	if err != nil {
		return "", NewSubcommandError("Application not found", err)
	}
	appGuid := model.Guid;

	output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v2/apps/%s/env", appGuid))
	if err != nil {
		return "", NewSubcommandError("Application environment error", err)
	}
	envJson := strings.Join(output, "")

	parser, err := gojq.NewStringQuery(envJson)
	if err != nil {
		return "", NewSubcommandError("Failed to parse application environment JSON", err)
	}
	creds, err := parser.Query("system_env_json.VCAP_SERVICES.p-service-registry.[0].credentials")
	if err != nil {
		return "", NewSubcommandError("Failed to find credentials in application environment JSON", err)
	}

	credentials, ok := creds.(map[string]interface{})
	if !ok {
		return "", NewSubcommandError("Credentials have wrong type", nil)
	}
	accessTokenUri, err := getString(credentials, "access_token_uri")
	if err != nil {
		return "", err
	}
	clientId, err := getString(credentials, "client_id")
	if err != nil {
		return "", err
	}
	clientSecret, err := getString(credentials, "client_secret")
	if err != nil {
		return "", err
	}
	eurekaUri, err := getString(credentials, "uri")
	if err != nil {
		return "", err
	}

	// FIXME: replace the following with a proper HTTP request. Using curl for spike only.
	cmd := exec.Command("curl", "-s" ,"-k", accessTokenUri, "-u", clientId + ":" + clientSecret, "-d", "grant_type=client_credentials")
	op, err := cmd.Output()
	if err != nil {
		return "", NewSubcommandError("curl failed", err)
	}

	parser, err = gojq.NewStringQuery(string(op))
	if err != nil {
		return "", NewSubcommandError("Failed to parse curl response JSON", err)
	}
	accessToken, err := parser.Query("access_token")
	if err != nil {
		return "", NewSubcommandError("Failed to find access token in curl response JSON", err)
	}
	registry, err := request(eurekaUri, accessToken.(string), "GET", "/eureka/apps")
	if err != nil {
		return "", err
	}

	var pretty bytes.Buffer
	err = json.Indent(&pretty, []byte(registry), "", "  ")
	if err != nil {
		return "", NewSubcommandError("json.Indent failed", err)
	}
	return string(pretty.Bytes()), nil
}

func request(uri string, token string, method string, path string) (string, error) {
	// FIXME: replace the following with a proper HTTP request. Using curl for spike only.
	cmd := exec.Command("curl", "-s" ,"-k", "-H", fmt.Sprintf("Authorization: bearer %s", token), "-H", "Accept: application/json", "-X", method, uri + path)
	op, err := cmd.Output()
	if err != nil {
		return "", NewSubcommandError("curl failed", err)
	}
	return string(op), nil
}

func getString(credentials map[string]interface{}, key string) (string, error) {
	value, ok := credentials[key].(string)
	if !ok {
		return "", NewSubcommandError(fmt.Sprintf("%s has wrong type", key), nil)
	}
	return value, nil
}

type SubcommandError struct {
	message  string
	cause error
}

func (se *SubcommandError) Error() string {
	return fmt.Sprintf("%s: %s", se.message, se.cause)
}

func NewSubcommandError(message string, cause error) error {
	return &SubcommandError{
		message: message,
		cause: cause,
	}
}