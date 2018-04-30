/*
Copyright (C) 2018 Synopsys, Inc.

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

package communicator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// SendPerceptorAddEvent sends an add event to perceptor at the dest endpoint
func SendPerceptorAddEvent(dest string, obj interface{}) error {
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("unable to serialize %v: %v", obj, err)
	}
	resp, err := http.Post(dest, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return fmt.Errorf("unable to POST to %s: %v", dest, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("http POST request to %s failed with status code %d", dest, resp.StatusCode)
	}
	return nil
}

// SendPerceptorDeleteEvent sends a delete event to perceptor at the dest endpoint
func SendPerceptorDeleteEvent(dest string, name string) error {
	jsonBytes, err := json.Marshal(name)
	if err != nil {
		return fmt.Errorf("unable to serialize %s: %v", name, err)
	}
	req, err := http.NewRequest("DELETE", dest, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return fmt.Errorf("unable to create DELETE request for %s: %v", dest, err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to DELETE to %s: %v", dest, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("http DELETE request to %s failed with status code %d", dest, resp.StatusCode)
	}
	return nil
}

// GetPerceptorScanResults will get scan results from the perceptor located at
// the provided url
func GetPerceptorScanResults(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("unable to GET %s for pod annotation: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to GET %s.  Got %d instead of 200", url, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read resp body from %s: %v", url, err)
	}

	return body, nil
}

// SendPerceptorData will send the given data to the provided url
func SendPerceptorData(url string, data []byte) error {
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("unable to create PUT request for %s: %v", url, err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to PUT to %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("http POST request to %s failed with status code %d", url, resp.StatusCode)
	}
	return nil
}
