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

	"k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewPerceptorPodFromKubePod(t *testing.T) {
	invalidPod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "invalidPod",
			Namespace: "ns",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				v1.Container{},
			},
		},
		Status: v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name:    "invalid",
					ImageID: "invalid ID",
				},
			},
		},
	}
	validPod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "podName",
			Namespace: "ns",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				v1.Container{},
				v1.Container{},
			},
		},
		Status: v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name:    "image1",
					ImageID: "docker-pullable://imageName@sha256:23f2sdf23",
					Image:   "imageName@sha256:23f2sdf23",
				},
				{
					Name:    "image2",
					ImageID: "docker-pullable://imageName2@sha256:4823nv823rn",
					Image:   "imageName2@sha256:4823nv823rn",
				},
			},
		},
	}
	validPerceptorPod := perceptorapi.Pod{
		Name:      "podName",
		Namespace: "ns",
		Containers: []perceptorapi.Container{
			{
				Name: "image1",
				Image: perceptorapi.Image{
					Name:        "imageName",
					Sha:         "23f2sdf23",
					DockerImage: "imageName@sha256:23f2sdf23",
				},
			},
			{
				Name: "image2",
				Image: perceptorapi.Image{
					Name:        "imageName2",
					Sha:         "4823nv823rn",
					DockerImage: "imageName2@sha256:4823nv823rn",
				},
			},
		},
	}

	missingImageIDPod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "podName",
			Namespace: "ns",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				v1.Container{},
			},
		},
		Status: v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name:  "image1",
					Image: "imageName@sha256:23f2sdf23",
				},
			},
		},
	}

	noContainerStatuses := v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				v1.Container{},
			},
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "podName",
			Namespace: "ns",
		},
		Status: v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{},
		},
	}

	testcases := []struct {
		description string
		pod         *v1.Pod
		expected    *perceptorapi.Pod
		shouldPass  bool
	}{
		{
			description: "valid pod with multiple containers",
			pod:         &validPod,
			expected:    &validPerceptorPod,
			shouldPass:  true,
		},
		{
			description: "invalid pod",
			pod:         &invalidPod,
			expected:    nil,
			shouldPass:  false,
		},
		{
			description: "pod with no ImageID",
			pod:         &missingImageIDPod,
			expected:    nil,
			shouldPass:  false,
		},
		{
			description: "pod with no container statuses",
			pod:         &noContainerStatuses,
			expected:    nil,
			shouldPass:  false,
		},
	}

	for _, tc := range testcases {
		result, err := NewPerceptorPodFromKubePod(tc.pod)
		if err != nil && tc.shouldPass {
			t.Errorf("[%s] unexpected error: %v", tc.description, err)
		}
		if result != tc.expected && !reflect.DeepEqual(result, tc.expected) {
			t.Errorf("[%s] expected %v, got %v", tc.description, tc.expected, result)
		}
	}
}
