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

	"k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes/fake"
)

func TestGetAllPodsAsPerceptorPods(t *testing.T) {
	invalidPod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "invalidPod",
			Namespace: "ns",
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
		Status: v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name:    "image1",
					ImageID: "docker-pullable://imageName@sha256:23f2sdf23",
					Image:   "imageName@sha256:23f2sdf23",
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
		},
	}

	testcases := []struct {
		description string
		kubePods    v1.PodList
		expected    []perceptorapi.Pod
		shouldPass  bool
	}{
		{
			description: "valid pods",
			kubePods:    v1.PodList{Items: []v1.Pod{validPod}},
			expected:    []perceptorapi.Pod{validPerceptorPod},
			shouldPass:  true,
		},
		{
			description: "invalid pod",
			kubePods:    v1.PodList{Items: []v1.Pod{invalidPod}},
			expected:    make([]perceptorapi.Pod, 0),
			shouldPass:  false,
		},
		{
			description: "invalid and valid pods",
			kubePods:    v1.PodList{Items: []v1.Pod{invalidPod, validPod}},
			expected:    []perceptorapi.Pod{validPerceptorPod},
			shouldPass:  false,
		},
	}

	for _, tc := range testcases {
		client := fake.NewSimpleClientset(&tc.kubePods)
		pd := PodDumper{
			coreV1: client.CoreV1(),
		}
		pods, err := pd.getAllPodsAsPerceptorPods()
		if err != nil && tc.shouldPass {
			t.Errorf("[%s] unexpected error: %v", tc.description, err)
		}
		for cnt, pod := range pods {
			if !reflect.DeepEqual(pod, tc.expected[cnt]) {
				t.Errorf("[%s] expected pod %v, got %v", tc.description, tc.expected[cnt], pod)
			}
		}
	}
}
