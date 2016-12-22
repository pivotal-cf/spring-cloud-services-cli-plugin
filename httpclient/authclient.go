package httpclient

import (
	"bytes"
	"fmt"
	"net/http"
)

//go:generate counterfeiter -o httpclientfakes/fake_authenticated_client.go . AuthenticatedClient
type AuthenticatedClient interface {
	DoAuthenticatedGet(url string, accessToken string) (*bytes.Buffer, error)
	DoAuthenticatedDelete(url string, accessToken string) error
}

func NewAuthenticatedClient(httpClient Client) *authenticatedClient {
	return &authenticatedClient{Httpclient: httpClient}
}

type authenticatedClient struct {
	Httpclient Client
}

func (c *authenticatedClient) DoAuthenticatedGet(url string, accessToken string) (*bytes.Buffer, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// Should never get here
		return nil, fmt.Errorf("Request creation error: %s", err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", accessToken)
	resp, err := c.Httpclient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("authenticated get of '%s' failed: %s", url, err)
	}

	body := resp.Body
	if body == nil {
		return nil, fmt.Errorf("authenticated get of '%s' failed: nil response body", url)
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(body)
	if err != nil {
		return nil, fmt.Errorf("authenticated get of '%s' failed: body cannot be read", url)
	}

	return buf, nil
}

func (c *authenticatedClient) DoAuthenticatedDelete(url string, accessToken string) error {
	req, err := http.NewRequest("DELETE", url, nil)
	fmt.Printf("Url: %s \n", url)
	if err != nil {
		// Should never get here
		return fmt.Errorf("Request creation error: %s", err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", accessToken)
	resp, err := c.Httpclient.Do(req)
	fmt.Printf("Url: %s \n Status: %s\n", url, resp.Status)
	if err != nil {
		return fmt.Errorf("authenticated delete of '%s' failed: %s", url, err)
	}

	return nil
}
