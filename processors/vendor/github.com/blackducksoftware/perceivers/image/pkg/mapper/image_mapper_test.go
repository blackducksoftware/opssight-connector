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

package mapper

import (
	"reflect"
	"testing"

	perceptorapi "github.com/blackducksoftware/perceptor/pkg/api"

	"github.com/openshift/api/image/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewPerceptorImageFromOSImage(t *testing.T) {
	invalidImage := v1.Image{
		ObjectMeta: metav1.ObjectMeta{
			Name: "invalidImage",
		},
		DockerImageReference: "imageName",
	}
	validImage := v1.Image{
		ObjectMeta: metav1.ObjectMeta{
			Name: "sha256:235n348g24",
		},
		DockerImageReference: "imageName@sha256:235n348g24",
	}
	validPerceptorImage := perceptorapi.Image{
		Name:        "imageName",
		Sha:         "235n348g24",
		DockerImage: "imageName@sha256:235n348g24",
	}

	testcases := []struct {
		description string
		image       *v1.Image
		expected    *perceptorapi.Image
		shouldPass  bool
	}{
		{
			description: "valid image",
			image:       &validImage,
			expected:    &validPerceptorImage,
			shouldPass:  true,
		},
		{
			description: "invalid image",
			image:       &invalidImage,
			expected:    nil,
			shouldPass:  false,
		},
	}

	for _, tc := range testcases {
		result, err := NewPerceptorImageFromOSImage(tc.image)
		if err != nil && tc.shouldPass {
			t.Fatalf("[%s] unexpected error: %v", tc.description, err)
		}
		if result != tc.expected && !reflect.DeepEqual(result, tc.expected) {
			t.Errorf("[%s] expected %v, got %v", tc.description, tc.expected, result)
		}
	}
}
