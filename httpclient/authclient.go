/*
 * Copyright (C) 2016-Present Pivotal Software, Inc. All rights reserved.
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
package httpclient

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

//go:generate counterfeiter -o httpclientfakes/fake_authenticated_client.go . AuthenticatedClient
type AuthenticatedClient interface {
	DoAuthenticatedGet(url string, accessToken string) (io.ReadCloser, int, error)

	DoAuthenticatedDelete(url string, accessToken string) (int, error)

	DoAuthenticatedPost(url string, bodyType string, body string, accessToken string) (io.ReadCloser, int, error)

	DoAuthenticatedPut(url string, accessToken string) (int, error)
}

type authenticatedClient struct {
	httpClient Client
}

func NewAuthenticatedClient(httpClient Client) *authenticatedClient {
	return &authenticatedClient{httpClient: httpClient}
}

func (c *authenticatedClient) DoAuthenticatedGet(url string, accessToken string) (io.ReadCloser, int, error) {
	statusCode := 0
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, statusCode, fmt.Errorf("Request creation error: %s", err)
	}

	req.Header.Add("Accept", "application/json")
	addAuthorizationHeader(req, accessToken)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("Authenticated get of '%s' failed: %s", url, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("Authenticated get of '%s' failed: %s", url, resp.Status)
	}

	return resp.Body, resp.StatusCode, nil
}

func (c *authenticatedClient) DoAuthenticatedDelete(url string, accessToken string) (int, error) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return 0, fmt.Errorf("Request creation error: %s", err)
	}

	req.Header.Add("Accept", "application/json")
	addAuthorizationHeader(req, accessToken)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("Authenticated delete of '%s' failed: %s", url, err)
	}
	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, fmt.Errorf("Authenticated delete of '%s' failed: %s", url, resp.Status)
	}
	return resp.StatusCode, nil
}

func (c *authenticatedClient) DoAuthenticatedPost(url string, bodyType string, bodyStr string, accessToken string) (io.ReadCloser, int, error) {
	body := strings.NewReader(bodyStr)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, 0, fmt.Errorf("Request creation error: %s", err)
	}
	addAuthorizationHeader(req, accessToken)
	req.Header.Set("Content-Type", bodyType)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("Authenticated post to '%s' failed: %s", url, err)
	}
	if resp.StatusCode != http.StatusOK {
		return resp.Body, resp.StatusCode, fmt.Errorf("Authenticated post to '%s' failed: %s", url, resp.Status)
	}

	return resp.Body, resp.StatusCode, nil
}

func (c *authenticatedClient) DoAuthenticatedPut(url string, accessToken string) (int, error) {
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return 0, fmt.Errorf("Request creation error: %s", err)
	}

	addAuthorizationHeader(req, accessToken)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("Authenticated put of '%s' failed: %s", url, err)
	}
	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, fmt.Errorf("Authenticated put of '%s' failed: %s", url, resp.Status)
	}
	return resp.StatusCode, nil
}

func addAuthorizationHeader(req *http.Request, accessToken string) {
	req.Header.Add("Authorization", fmt.Sprintf("bearer %s", accessToken))
}
