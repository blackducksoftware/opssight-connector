/*
Copyright (C) 2019 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.
*/

package utils

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// RegistryAuth stores the credentials for a private docker repo
// and is same as common.RegistryAuth in perceptor-scanner repo
type RegistryAuth struct {
	URL      string
	User     string
	Password string
	Token    string
}

// GetResourceOfType takes in the specified URL with credentials and
// tries to decode returning json to specified interface
func GetResourceOfType(url string, cred *RegistryAuth, bearerToken string, target interface{}) error {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("Error in creating get request %e at url %s", err, url)
	}

	if cred != nil {
		req.SetBasicAuth(cred.User, cred.Password)
	}

	if len(bearerToken) > 0 {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}

// PingArtifactoryServer takes in the specified URL with username & password and checks weather
// it's a valid login for artifactory by pinging the server with various options and returns the correct URL
func PingArtifactoryServer(url string, username string, password string) (*RegistryAuth, error) {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr}

	url = fmt.Sprintf("%s/api/system/ping", url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Error in pinging artifactory server %e", err)
	}
	req.SetBasicAuth(username, password)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {

		// The instance may contain /artifactory
		if !strings.Contains(url, "/artifactory") {
			url = strings.Replace(url, "/api/system/ping", "/artifactory", -1)
			return PingArtifactoryServer(url, username, password)
		}

		// Making sure that http and https both return not OK
		if strings.Contains(url, "https://") {
			url = strings.Replace(url, "https://", "http://", -1)
			// Reset to baseURL
			url = strings.Replace(url, "/api/system/ping", "", -1)
			return PingArtifactoryServer(url, username, password)
		}

		return nil, fmt.Errorf("Error in pinging artifactory server supposed to get %d response code got %d", http.StatusOK, resp.StatusCode)
	}

	// Reset to baseURL
	url = strings.Replace(url, "/api/system/ping", "", -1)
	return &RegistryAuth{URL: url, User: username, Password: password}, nil
}
