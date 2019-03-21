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
package httpclient

import (
	"bytes"
	"fmt"
	"net/http"
)

//go:generate counterfeiter -o httpclientfakes/fake_authenticated_client.go . AuthenticatedClient
type AuthenticatedClient interface {
	DoAuthenticatedGet(url string, accessToken string) (*bytes.Buffer, int, error)
	DoAuthenticatedDelete(url string, accessToken string) error
}

type authenticatedClient struct {
	Httpclient Client
}

func NewAuthenticatedClient(httpClient Client) *authenticatedClient {
	return &authenticatedClient{Httpclient: httpClient}
}

func (c *authenticatedClient) DoAuthenticatedGet(url string, accessToken string) (*bytes.Buffer, int, error) {
	statusCode := 0
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// Should never get here
		return nil, statusCode, fmt.Errorf("Request creation error: %s", err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", accessToken)
	resp, err := c.Httpclient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("authenticated get of '%s' failed: %s", url, err)
	}

	statusCode = resp.StatusCode
	body := resp.Body
	if body == nil {
		return nil, statusCode, fmt.Errorf("authenticated get of '%s' failed: nil response body", url)
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(body)
	if err != nil {
		return nil, statusCode, fmt.Errorf("authenticated get of '%s' failed: body cannot be read", url)
	}

	return buf, statusCode, nil
}

func (c *authenticatedClient) DoAuthenticatedDelete(url string, accessToken string) error {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		// Should never get here
		return fmt.Errorf("Request creation error: %s", err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", accessToken)
	resp, err := c.Httpclient.Do(req)
	if err != nil {
		return fmt.Errorf("authenticated delete of '%s' failed: %s", url, err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("authenticated delete of '%s' returned incorrect status code: %s\n", url, resp.StatusCode)
	}
	return nil
}
