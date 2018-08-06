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

package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// LogInterface is a simple interface providing Errorf an Logf
type LogInterface interface {
	Errorf(format string, args ...interface{})
	Logf(format string, args ...interface{})
}

// FakeHandler is to assist in testing HTTP requests
type FakeHandler struct {
	RequestReceived *http.Request
	RequestBody     string
	StatusCode      int
	RespondBody     string
	T               LogInterface
}

func (f *FakeHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	f.RequestReceived = request
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(f.StatusCode)
	response.Write([]byte(f.RespondBody))
	bodyReceived, err := ioutil.ReadAll(request.Body)
	if err != nil && f.T != nil {
		f.T.Logf("Received read error: %v", err)
	}
	f.RequestBody = string(bodyReceived)
}

// Validate verifies that FakeHandler received a request with expected path, method, and body.
func (f *FakeHandler) Validate(expectedPath string, expectedMethod string, expectedBody *string) error {
	expectURL, err := url.Parse(expectedPath)
	if err != nil {
		return fmt.Errorf("couldn't parse %v as a URL", expectedPath)
	}
	if f.RequestReceived == nil {
		return fmt.Errorf("unexpected nil request received for %s", expectedPath)
	}
	if f.RequestReceived.URL.Path != expectURL.Path {
		return fmt.Errorf("unexpected request path for request %#v, received: %q, expected: %q", f.RequestReceived, f.RequestReceived.URL.Path, expectURL.Path)
	}
	if f.RequestReceived.Method != expectedMethod {
		return fmt.Errorf("unexpected method: %q, expected: %q", f.RequestReceived.Method, expectedMethod)
	}
	if expectedBody != nil {
		if *expectedBody != f.RequestBody {
			return fmt.Errorf("received body:\n%s\n Doesn't match expected body:\n%s", f.RequestBody, *expectedBody)
		}
	}
	return nil
}
