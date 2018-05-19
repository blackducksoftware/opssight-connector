/*
Copyright (C) 2018 Black Duck Software, Inc.

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

package basicskyfire

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	skyfire "github.com/blackducksoftware/perceptor-skyfire/pkg/report"
)

func FetchSkyfireReport(skyfireURL string) (*skyfire.Report, error) {
	bodyBytes, err := getHttpResponse(skyfireURL, 200)

	if err != nil {
		panic(fmt.Sprintf("Unable to get the response for %s due to %+v", skyfireURL, err.Error()))
	}

	var report *skyfire.Report
	err = json.Unmarshal(bodyBytes, &report)
	if err != nil {
		return nil, err
	}

	return report, nil
}

func getHttpResponseBody(url string, responseCode int) (io.ReadCloser, error) {
	httpClient := http.Client{Timeout: 5 * time.Second}
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp.Body, err
}

func getHttpResponse(url string, responseCode int) ([]byte, error) {
	httpClient := http.Client{Timeout: 5 * time.Second}
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != responseCode {
		return nil, fmt.Errorf("invalid status code %d, expected 200", resp.StatusCode)
	}

	return bodyBytes, nil
}

func PrettyPrint(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "  ")
	println(string(b))
}
