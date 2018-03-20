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

package annotator

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/blackducksoftware/perceivers/pkg/annotations"
	"github.com/blackducksoftware/perceivers/pkg/utils"

	perceptorapi "github.com/blackducksoftware/perceptor/pkg/api"

	"k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var scannedImages = []perceptorapi.ScannedImage{
	{
		Name:             "image1",
		Sha:              "ASDJ4FSF3FSFK3SF450",
		PolicyViolations: 100,
		Vulnerabilities:  5,
		OverallStatus:    "STATUS3",
		ComponentsURL:    "http://url.com",
	},
	{
		Name:             "image2",
		Sha:              "HAFGW2392FJGNE3FFK04",
		PolicyViolations: 5,
		Vulnerabilities:  15,
		OverallStatus:    "STATUS4",
		ComponentsURL:    "http://new.com/location",
	},
}

var scannedPods = []perceptorapi.ScannedPod{
	{
		Name:             "pod1",
		Namespace:        "ns1",
		PolicyViolations: 10,
		Vulnerabilities:  0,
		OverallStatus:    "STATUS1",
	},
	{
		Name:             "pod2",
		Namespace:        "ns2",
		PolicyViolations: 0,
		Vulnerabilities:  20,
		OverallStatus:    "STATUS2",
	},
}

var results = perceptorapi.ScanResults{
	HubScanClientVersion: "version.1",
	HubVersion:           "version.2",
	Pods:                 scannedPods,
	Images:               scannedImages,
}

func makePodAnnotationObj() *annotations.PodAnnotationData {
	pod := scannedPods[0]
	return annotations.NewPodAnnotationData(pod.PolicyViolations, pod.Vulnerabilities, pod.OverallStatus, results.HubVersion, results.HubScanClientVersion)
}

func makePodWithImage(name string, sha string) *v1.Pod {
	scannedPod := scannedPods[0]
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      scannedPod.Name,
			Namespace: scannedPod.Namespace,
		},
		Status: v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name:    name,
					ImageID: fmt.Sprintf("docker-pullable://%s@sha256:%s", name, sha),
				},
			},
		},
	}
}

func makePod() *v1.Pod {
	scannedImage := scannedImages[0]
	return makePodWithImage(scannedImage.Name, scannedImage.Sha)
}

func TestGetScanResults(t *testing.T) {
	testcases := []struct {
		description   string
		statusCode    int
		body          *perceptorapi.ScanResults
		expectedScans *perceptorapi.ScanResults
		shouldPass    bool
	}{
		{
			description:   "successful GET with actual results",
			statusCode:    200,
			body:          &results,
			expectedScans: &results,
			shouldPass:    true,
		},
		{
			description:   "successful GET with empty results",
			statusCode:    200,
			body:          &perceptorapi.ScanResults{},
			expectedScans: &perceptorapi.ScanResults{},
			shouldPass:    true,
		},
		{
			description:   "bad status code",
			statusCode:    401,
			body:          nil,
			expectedScans: nil,
			shouldPass:    false,
		},
		{
			description:   "nil body on successful GET",
			statusCode:    200,
			body:          nil,
			expectedScans: &perceptorapi.ScanResults{},
			shouldPass:    true,
		},
	}

	endpoint := "RESTEndpoint"
	for _, tc := range testcases {
		bytes, _ := json.Marshal(tc.body)
		handler := utils.FakeHandler{
			StatusCode:  tc.statusCode,
			RespondBody: string(bytes),
			T:           t,
		}
		server := httptest.NewServer(&handler)
		defer server.Close()

		annotator := PodAnnotator{
			scanResultsURL: fmt.Sprintf("%s/%s", server.URL, endpoint),
		}
		scanResults, err := annotator.getScanResults()
		if err != nil && tc.shouldPass {
			t.Fatalf("[%s] unexpected error: %v", tc.description, err)
		}
		if !reflect.DeepEqual(tc.expectedScans, scanResults) {
			t.Errorf("[%s] received %v expected %v", tc.description, scanResults, tc.expectedScans)
		}
	}
}

func TestAddPodAnnotations(t *testing.T) {
	fullAnnotationSet := func() map[string]string {
		pod := makePod()
		annotationObj := makePodAnnotationObj()
		pa := PodAnnotator{}
		return pa.createNewAnnotations(pod, annotationObj, scannedImages)
	}

	partialPodAnnotationSet := func() map[string]string {
		annotations := fullAnnotationSet()
		for k := range annotations {
			if strings.Contains(k, fmt.Sprintf("%s/vulnerability", bdannotations.BDPodAnnotationPrefix)) {
				delete(annotations, k)
			}
		}
		return annotations
	}

	partialImageAnnotationSet := func() map[string]string {
		annotations := fullAnnotationSet()
		for k := range annotations {
			if strings.Contains(k, "blackducksoftware") {
				delete(annotations, k)
			}
		}
		return annotations
	}

	testcases := []struct {
		description         string
		pod                 *v1.Pod
		existingAnnotations map[string]string
		shouldAdd           bool
	}{
		{
			description:         "pod with no annotations",
			pod:                 makePod(),
			existingAnnotations: make(map[string]string),
			shouldAdd:           true,
		},
		{
			description:         "pod with existing annotations, no overlap",
			pod:                 makePod(),
			existingAnnotations: map[string]string{"key1": "value1", "key2": "value2"},
			shouldAdd:           true,
		},
		{
			description:         "pod with existing annotations, some pod overlap",
			pod:                 makePod(),
			existingAnnotations: partialPodAnnotationSet(),
			shouldAdd:           true,
		},
		{
			description:         "pod with existing annotations, some image overlap",
			pod:                 makePod(),
			existingAnnotations: partialImageAnnotationSet(),
			shouldAdd:           true,
		},
		{
			description:         "pod with exact existing annotations",
			pod:                 makePod(),
			existingAnnotations: fullAnnotationSet(),
			shouldAdd:           false,
		},
		{
			description:         "pod with no scanned images",
			pod:                 makePodWithImage("imageName", "234F8sdgj235jsdf923"),
			existingAnnotations: make(map[string]string),
			shouldAdd:           true,
		},
		{
			description:         "pod with no scanned images, existing pod annotations",
			pod:                 makePodWithImage("imageName", "234F8sdgj235jsdf923"),
			existingAnnotations: bdannotations.CreatePodAnnotations(makePodAnnotationObj()),
			shouldAdd:           false,
		},
	}

	for _, tc := range testcases {
		annotationObj := makePodAnnotationObj()
		tc.pod.SetAnnotations(tc.existingAnnotations)
		pa := PodAnnotator{}
		result := pa.addPodAnnotations(tc.pod, annotationObj, scannedImages)
		if result != tc.shouldAdd {
			t.Fatalf("[%s] expected %t, got %t", tc.description, tc.shouldAdd, result)
		}
		new := pa.createNewAnnotations(tc.pod, annotationObj, scannedImages)
		updated := tc.pod.GetAnnotations()
		for k, v := range new {
			if val, ok := updated[k]; !ok {
				t.Errorf("[%s] key %s doesn't exist in pod annotations %v", tc.description, k, updated)
			} else if val != v {
				t.Errorf("[%s] key %s has wrong value in pod annotation.  Expected %s got %s", tc.description, k, new[k], updated[k])
			}
		}
	}
}

func TestAddPodLabels(t *testing.T) {
	fullLabelSet := func() map[string]string {
		pod := makePod()
		annotationObj := makePodAnnotationObj()
		pa := PodAnnotator{}
		return pa.createNewLabels(pod, annotationObj, scannedImages)
	}

	partialPodLabelSet := func() map[string]string {
		labels := fullLabelSet()
		for k := range labels {
			if strings.Contains(k, "violations") {
				delete(labels, k)
			}
		}
		return labels
	}

	partialImageLabelSet := func() map[string]string {
		labels := fullLabelSet()
		for k := range labels {
			if strings.Contains(k, "violations") {
				delete(labels, k)
			}
		}
		return labels
	}

	testcases := []struct {
		description    string
		pod            *v1.Pod
		existingLabels map[string]string
		shouldAdd      bool
	}{
		{
			description:    "pod with no labels",
			pod:            makePod(),
			existingLabels: make(map[string]string),
			shouldAdd:      true,
		},
		{
			description:    "pod with existing labels, no overlap",
			pod:            makePod(),
			existingLabels: map[string]string{"key1": "value1", "key2": "value2"},
			shouldAdd:      true,
		},
		{
			description:    "pod with existing labels, some pod overlap",
			pod:            makePod(),
			existingLabels: partialPodLabelSet(),
			shouldAdd:      true,
		},
		{
			description:    "pod with existing labels, some image overlap",
			pod:            makePod(),
			existingLabels: partialImageLabelSet(),
			shouldAdd:      true,
		},
		{
			description:    "pod with exact existing labels",
			pod:            makePod(),
			existingLabels: fullLabelSet(),
			shouldAdd:      false,
		},
		{
			description:    "pod with no scanned images",
			pod:            makePodWithImage("imageName", "234F8sdgj235jsdf923"),
			existingLabels: make(map[string]string),
			shouldAdd:      true,
		},
		{
			description:    "pod with no scanned images, existing pod labels",
			pod:            makePodWithImage("imageName", "234F8sdgj235jsdf923"),
			existingLabels: bdannotations.CreatePodLabels(makePodAnnotationObj()),
			shouldAdd:      false,
		},
	}

	for _, tc := range testcases {
		annotationObj := makePodAnnotationObj()
		tc.pod.SetLabels(tc.existingLabels)
		pa := PodAnnotator{}
		result := pa.addPodLabels(tc.pod, annotationObj, scannedImages)
		if result != tc.shouldAdd {
			t.Fatalf("[%s] expected %t, got %t", tc.description, tc.shouldAdd, result)
		}
		new := pa.createNewLabels(tc.pod, annotationObj, scannedImages)
		updated := tc.pod.GetLabels()
		for k, v := range new {
			if val, ok := updated[k]; !ok {
				t.Errorf("[%s] key %s doesn't exist in pod labels %v", tc.description, k, updated)
			} else if val != v {
				t.Errorf("[%s] key %s has wrong value in pod label.  Expected %s got %s", tc.description, k, new[k], updated[k])
			}
		}
	}
}

func TestGetPodContainerMap(t *testing.T) {
	generator := func(obj *annotations.ImageAnnotationData, name string, count int) map[string]string {
		return map[string]string{fmt.Sprintf("key%d", count): fmt.Sprintf("%s%d", name, count)}
	}
	imageWithoutPrefix := v1.ContainerStatus{
		Name:    "notscanned",
		ImageID: "repository.com/notscanned@sha256:34545ngelkj235knegr",
	}

	imageWithPrefix := v1.ContainerStatus{
		Name:    "notscanned",
		ImageID: "docker-pullable://repository.com/notscanned@sha256:j2345msdf9235nb834",
	}

	testcases := []struct {
		description      string
		pod              *v1.Pod
		additionalImages []v1.ContainerStatus
		resultMap        map[string]string
	}{
		{
			description:      "all containers scanned",
			pod:              makePod(),
			additionalImages: make([]v1.ContainerStatus, 0),
			resultMap:        map[string]string{"key0": scannedImages[0].Name + "0"},
		},
		{
			description:      "one container scanned, one not scanned",
			pod:              makePod(),
			additionalImages: []v1.ContainerStatus{imageWithPrefix},
			resultMap:        map[string]string{"key0": scannedImages[0].Name + "0"},
		},
		{
			description:      "2 images without scans",
			pod:              &v1.Pod{},
			additionalImages: []v1.ContainerStatus{imageWithPrefix, imageWithoutPrefix},
			resultMap:        make(map[string]string),
		},
	}

	for _, tc := range testcases {
		pa := PodAnnotator{}
		for _, image := range tc.additionalImages {
			tc.pod.Status.ContainerStatuses = append(tc.pod.Status.ContainerStatuses, image)
		}
		new := pa.getPodContainerMap(tc.pod, scannedImages, "hub version", "scan client version", generator)
		if !reflect.DeepEqual(new, tc.resultMap) {
			t.Errorf("[%s] container maps are different.  Expected %v got %v", tc.description, tc.resultMap, new)
		}
	}
}

func TestFindImageAnnotations(t *testing.T) {
	testcases := []struct {
		description string
		name        string
		sha         string
		result      *perceptorapi.ScannedImage
	}{
		{
			description: "finds name and sha in scanned images",
			name:        "image1",
			sha:         "ASDJ4FSF3FSFK3SF450",
			result:      &scannedImages[0],
		},
		{
			description: "correct name, wrong sha",
			name:        "image1",
			sha:         "asj23gadgk234",
			result:      nil,
		},
		{
			description: "correct sha, wrong name",
			name:        "notfound",
			sha:         "ASDJ4FSF3FSFK3SF450",
			result:      nil,
		},
		{
			description: "wrong name and sha",
			name:        "notfound",
			sha:         "asj23gadgk234",
			result:      nil,
		},
	}

	for _, tc := range testcases {
		pa := PodAnnotator{}
		result := pa.findImageAnnotations(tc.name, tc.sha, scannedImages)
		if result != tc.result && !reflect.DeepEqual(*result, *tc.result) {
			t.Errorf("[%s] expected %v got %v: name %s, sha %s", tc.description, tc.result, result, tc.name, tc.sha)
		}
	}
}

func TestAnnotate(t *testing.T) {
	testcases := []struct {
		description string
		statusCode  int
		body        *perceptorapi.ScanResults
		shouldPass  bool
	}{
		{
			description: "successful GET with empty results",
			statusCode:  200,
			body:        &perceptorapi.ScanResults{},
			shouldPass:  true,
		},
		{
			description: "failed to annotate",
			statusCode:  401,
			body:        nil,
			shouldPass:  false,
		},
	}
	endpoint := "RESTEndpoint"
	for _, tc := range testcases {
		bytes, _ := json.Marshal(tc.body)
		handler := utils.FakeHandler{
			StatusCode:  tc.statusCode,
			RespondBody: string(bytes),
			T:           t,
		}
		server := httptest.NewServer(&handler)
		defer server.Close()

		annotator := PodAnnotator{
			scanResultsURL: fmt.Sprintf("%s/%s", server.URL, endpoint),
		}
		err := annotator.annotate()
		if err != nil && tc.shouldPass {
			t.Fatalf("[%s] unexpected error: %v", tc.description, err)
		}
		if err == nil && !tc.shouldPass {
			t.Errorf("[%s] expected error but didn't receive one", tc.description)
		}
	}
}
