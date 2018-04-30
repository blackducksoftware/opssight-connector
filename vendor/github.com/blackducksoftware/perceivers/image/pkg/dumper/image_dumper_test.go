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

package dumper

import (
	"reflect"
	"testing"

	perceptorapi "github.com/blackducksoftware/perceptor/pkg/api"

	"github.com/openshift/client-go/image/clientset/versioned/fake"

	"github.com/openshift/api/image/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetAllImagesAsPerceptorImages(t *testing.T) {
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
		osImages    v1.ImageList
		expected    []perceptorapi.Image
		shouldPass  bool
	}{
		{
			description: "valid image",
			osImages:    v1.ImageList{Items: []v1.Image{validImage}},
			expected:    []perceptorapi.Image{validPerceptorImage},
			shouldPass:  true,
		},
		{
			description: "invalid image",
			osImages:    v1.ImageList{Items: []v1.Image{invalidImage}},
			expected:    make([]perceptorapi.Image, 0),
			shouldPass:  false,
		},
		{
			description: "invalid and valid images",
			osImages:    v1.ImageList{Items: []v1.Image{invalidImage, validImage}},
			expected:    []perceptorapi.Image{validPerceptorImage},
			shouldPass:  false,
		},
	}

	for _, tc := range testcases {
		client := fake.NewSimpleClientset(&tc.osImages)
		id := ImageDumper{
			client: client.ImageV1(),
		}
		images, err := id.getAllImagesAsPerceptorImages()
		if err != nil && tc.shouldPass {
			t.Errorf("[%s] unexpected error: %v", tc.description, err)
		}
		for cnt, image := range images {
			if !reflect.DeepEqual(image, tc.expected[cnt]) {
				t.Errorf("[%s] expected pod %v, got %v", tc.description, tc.expected[cnt], image)
			}
		}
	}
}
