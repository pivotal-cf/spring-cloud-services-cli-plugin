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
package httpclient

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

//go:generate counterfeiter -o httpclientfakes/fake_authenticated_client.go . AuthenticatedClient
type AuthenticatedClient interface {
	GetClientCredentialsAccessToken(accessTokenURI string, clientId string, clientSecret string) (string, error)

	// The access token must include the "bearer" authorisation scheme.
	// See https://www.pivotaltracker.com/story/show/143022981
	DoAuthenticatedGet(url string, accessToken string) (io.ReadCloser, int, error)

	// The access token must include the "bearer" authorisation scheme.
	// See https://www.pivotaltracker.com/story/show/143022981
	DoAuthenticatedDelete(url string, accessToken string) (int, error)

	DoAuthenticatedPost(url string, bodyType string, body string, accessToken string) (io.ReadCloser, int, error)
}

type authenticatedClient struct {
	Httpclient Client
}

func NewAuthenticatedClient(httpClient Client) *authenticatedClient {
	return &authenticatedClient{Httpclient: httpClient}
}

type accessTokenInfo struct {
	AccessToken string `json:"access_token"`
}

func (c *authenticatedClient) GetClientCredentialsAccessToken(accessTokenURI string, clientId string, clientSecret string) (string, error) {
	body := strings.NewReader("grant_type=client_credentials")
	req, err := http.NewRequest("POST", accessTokenURI, body)
	if err != nil {
		return "", fmt.Errorf("Failed to create access token request: %s", err)
	}
	req.SetBasicAuth(clientId, clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.Httpclient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Failed to obtain access token: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Failed to obtain access token: %s", resp.Status)
	}

	defer resp.Body.Close()
	accessToken, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read access token: %s", err)
	}
	accessTokenInfo := &accessTokenInfo{}
	err = json.Unmarshal(accessToken, accessTokenInfo)
	if err != nil {
		return "", fmt.Errorf("Failed to unmarshal access token: %s", err)
	}
	return accessTokenInfo.AccessToken, nil
}

func (c *authenticatedClient) DoAuthenticatedGet(url string, accessToken string) (io.ReadCloser, int, error) {
	statusCode := 0
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, statusCode, fmt.Errorf("Request creation error: %s", err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", accessToken)
	resp, err := c.Httpclient.Do(req)
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
	req.Header.Add("Authorization", accessToken)
	resp, err := c.Httpclient.Do(req)
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
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", bodyType)
	resp, err := c.Httpclient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("Authenticated post to '%s' failed: %s", url, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("Authenticated post to '%s' failed: %s", url, resp.Status)
	}

	return resp.Body, resp.StatusCode, nil
}
