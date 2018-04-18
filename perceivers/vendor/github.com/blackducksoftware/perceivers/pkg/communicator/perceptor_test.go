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
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/blackducksoftware/perceivers/pkg/utils"
)

type MockContainer struct {
	Name  string
	Image MockImage
}

type MockImage struct {
	Name        string
	Sha         string
	DockerImage string
}

func TestSendPerceptorAddDeleteEvent(t *testing.T) {
	pod := struct {
		Name       string
		Namespace  string
		Containers []MockContainer
	}{
		Name:      "test",
		Namespace: "testNS",
		Containers: []MockContainer{
			{
				Name: "fakeC1",
				Image: MockImage{
					Name:        "fakeImage1",
					Sha:         "sha1",
					DockerImage: "dockerimage1",
				},
			},
			{
				Name: "fakeC2",
				Image: MockImage{
					Name:        "fakeImage2",
					Sha:         "sha2",
					DockerImage: "dockerimage2",
				},
			},
		},
	}
	testcases := []struct {
		description string
		statusCode  int
		shouldPass  bool
	}{
		{
			description: "valid request",
			statusCode:  200,
			shouldPass:  true,
		},
		{
			description: "server error",
			statusCode:  401,
			shouldPass:  false,
		},
	}

	endpoint := "RESTEndpoint"
	for _, tc := range testcases {
		handler := utils.FakeHandler{
			StatusCode:  tc.statusCode,
			RespondBody: "",
			T:           t,
		}
		server := httptest.NewServer(&handler)
		defer server.Close()

		// Test sending an add event
		err := SendPerceptorAddEvent(fmt.Sprintf("%s/%s", server.URL, endpoint), pod)
		if err != nil && tc.shouldPass {
			t.Errorf("[%s] unexpected error on add: %v", tc.description, err)
		}
		bytes, _ := json.Marshal(pod)
		body := string(bytes)
		err = handler.Validate(fmt.Sprintf("/%s", endpoint), "POST", &body)
		if err != nil && tc.shouldPass {
			t.Errorf("[%s] validate failed on add: %v", tc.description, err)
		}

		// Test sending an add event
		bytes, _ = json.Marshal(pod.Name)
		body = string(bytes)
		err = SendPerceptorDeleteEvent(fmt.Sprintf("%s/%s", server.URL, endpoint), pod.Name)
		if err != nil && tc.shouldPass {
			t.Errorf("[%s] unexpected error on delete: %v", tc.description, err)
		}

		err = handler.Validate(fmt.Sprintf("/%s", endpoint), "DELETE", &body)
		if err != nil && tc.shouldPass {
			t.Errorf("[%s] validate failed on delete: %v", tc.description, err)
		}
	}
}

func TestGetPerceptorScanResults(t *testing.T) {
	testcases := []struct {
		description string
		statusCode  int
		body        []byte
		shouldPass  bool
	}{
		{
			description: "successful GET with actual results",
			statusCode:  200,
			body:        []byte{23, 42, 5, 10},
			shouldPass:  true,
		},
		{
			description: "successful GET with empty results",
			statusCode:  200,
			body:        make([]byte, 0),
			shouldPass:  true,
		},
		{
			description: "bad status code",
			statusCode:  401,
			body:        make([]byte, 0),
			shouldPass:  false,
		},
		{
			description: "nil body on successful GET",
			statusCode:  200,
			body:        make([]byte, 0),
			shouldPass:  false,
		},
	}

	endpoint := "RESTEndpoint"
	for _, tc := range testcases {
		handler := utils.FakeHandler{
			StatusCode:  tc.statusCode,
			RespondBody: string(tc.body),
			T:           t,
		}
		server := httptest.NewServer(&handler)
		defer server.Close()

		gotResults, err := GetPerceptorScanResults(fmt.Sprintf("%s/%s", server.URL, endpoint))
		if err != nil && tc.shouldPass {
			t.Fatalf("[%s] unexpected error: %v", tc.description, err)
		}
		if tc.shouldPass && !reflect.DeepEqual(tc.body, gotResults) {
			t.Errorf("[%s] received %v expected %v", tc.description, gotResults, tc.body)
		}
	}
}

func TestSendPerceptorData(t *testing.T) {
	testcases := []struct {
		description string
		statusCode  int
		body        []byte
		shouldPass  bool
	}{
		{
			description: "successful send",
			statusCode:  200,
			body:        []byte{23, 42, 5, 10},
			shouldPass:  true,
		},
		{
			description: "server error",
			statusCode:  401,
			body:        make([]byte, 0),
			shouldPass:  false,
		},
	}

	endpoint := "endpoint"
	for _, tc := range testcases {
		handler := utils.FakeHandler{
			StatusCode:  tc.statusCode,
			RespondBody: "",
			T:           t,
		}
		server := httptest.NewServer(&handler)
		defer server.Close()

		err := SendPerceptorData(fmt.Sprintf("%s/%s", server.URL, endpoint), tc.body)
		if err != nil && tc.shouldPass {
			t.Fatalf("[%s] unexpected error: %v", tc.description, err)
		}

		body := string(tc.body)
		err = handler.Validate(fmt.Sprintf("/%s", endpoint), "PUT", &body)
		if err != nil && tc.shouldPass {
			t.Errorf("[%s] validate failed: %v", tc.description, err)
		}
	}
}
