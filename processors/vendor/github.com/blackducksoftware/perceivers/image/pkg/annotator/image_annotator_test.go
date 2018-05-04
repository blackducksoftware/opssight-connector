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

	"github.com/openshift/api/image/v1"

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

var results = perceptorapi.ScanResults{
	HubScanClientVersion: "version.1",
	HubVersion:           "version.2",
	Pods:                 []perceptorapi.ScannedPod{},
	Images:               scannedImages,
}

var imageAnnotations = map[string]string{"imageannotationkey1": "imagevalue1", "imageannotationkey2": "imagevalue2", "imageannotationkey3": "imagevalue3"}
var imageLabels = map[string]string{"imagelabelkey1": "imagevalue1", "imagelabelkey2": "imagevalue2", "imagelabelkey3": "imagevalue3"}

func makeImageAnnotationObj() *annotations.ImageAnnotationData {
	image := scannedImages[0]
	return annotations.NewImageAnnotationData(image.PolicyViolations, image.Vulnerabilities, image.OverallStatus, image.ComponentsURL, results.HubVersion, results.HubScanClientVersion)
}

func makeImage(name string, sha string) *v1.Image {
	return &v1.Image{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("sha256:%s", sha),
		},
		DockerImageReference: fmt.Sprintf("%s@sha256:%s", name, sha),
	}
}

func imageLabelGenerator(obj interface{}, name string, count int) map[string]string {
	return imageLabels
}

func imageAnnotationGenerator(obj interface{}, name string, count int) map[string]string {
	return imageAnnotations
}

func createIA() *ImageAnnotator {
	return &ImageAnnotator{h: annotations.ImageAnnotatorHandlerFuncs{
		ImageLabelCreationFunc:      imageLabelGenerator,
		ImageAnnotationCreationFunc: imageAnnotationGenerator,
		MapCompareHandlerFuncs: annotations.MapCompareHandlerFuncs{
			MapCompareFunc: annotations.StringMapContains,
		},
	}}
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

		annotator := ImageAnnotator{
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

func TestAddImageAnnotations(t *testing.T) {
	partialAnnotationSet := func() map[string]string {
		annotations := make(map[string]string)
		for k, v := range imageAnnotations {
			if !strings.Contains(k, "imageannotationkey2") {
				annotations[k] = v
			}
		}
		return annotations
	}

	otherAnnotations := map[string]string{"key1": "value1", "key2": "value2"}

	testcases := []struct {
		description         string
		image               *v1.Image
		existingAnnotations map[string]string
		expectedAnnotations map[string]string
		shouldAdd           bool
	}{
		{
			description:         "image with no annotations",
			image:               makeImage(scannedImages[0].Name, scannedImages[0].Sha),
			existingAnnotations: make(map[string]string),
			expectedAnnotations: imageAnnotations,
			shouldAdd:           true,
		},
		{
			description:         "image with existing annotations, no overlap",
			image:               makeImage(scannedImages[0].Name, scannedImages[0].Sha),
			existingAnnotations: otherAnnotations,
			expectedAnnotations: utils.MapMerge(otherAnnotations, imageAnnotations),
			shouldAdd:           true,
		},
		{
			description:         "pod with existing annotations, some overlap",
			image:               makeImage(scannedImages[0].Name, scannedImages[0].Sha),
			existingAnnotations: partialAnnotationSet(),
			expectedAnnotations: imageAnnotations,
			shouldAdd:           true,
		},
		{
			description:         "image with exact existing annotations",
			image:               makeImage(scannedImages[0].Name, scannedImages[0].Sha),
			existingAnnotations: imageAnnotations,
			expectedAnnotations: imageAnnotations,
			shouldAdd:           false,
		},
	}

	for _, tc := range testcases {
		annotationObj := makeImageAnnotationObj()
		fullName := fmt.Sprintf("%s@sha256:%s", scannedImages[0].Name, scannedImages[0].Sha)
		tc.image.SetAnnotations(tc.existingAnnotations)
		result := createIA().addImageAnnotations(fullName, tc.image, annotationObj)
		if result != tc.shouldAdd {
			t.Fatalf("[%s] expected %t, got %t", tc.description, tc.shouldAdd, result)
		}
		updated := tc.image.GetAnnotations()
		for k, v := range tc.expectedAnnotations {
			if val, ok := updated[k]; !ok {
				t.Errorf("[%s] key %s doesn't exist in image annotations %v", tc.description, k, updated)
			} else if val != v {
				t.Errorf("[%s] key %s has wrong value in image annotation.  Expected %s got %s", tc.description, k, tc.expectedAnnotations[k], updated[k])
			}
		}
	}
}

func TestAddPodLabels(t *testing.T) {
	partialLabelSet := func() map[string]string {
		labels := make(map[string]string)
		for k, v := range imageLabels {
			if !strings.Contains(k, "imagelabelkey2") {
				labels[k] = v
			}
		}
		return labels
	}

	otherLabels := map[string]string{"key1": "value1", "key2": "value2"}

	testcases := []struct {
		description    string
		image          *v1.Image
		existingLabels map[string]string
		expectedLabels map[string]string
		shouldAdd      bool
	}{
		{
			description:    "image with no labels",
			image:          makeImage(scannedImages[0].Name, scannedImages[0].Sha),
			existingLabels: make(map[string]string),
			expectedLabels: imageLabels,
			shouldAdd:      true,
		},
		{
			description:    "image with existing labels, no overlap",
			image:          makeImage(scannedImages[0].Name, scannedImages[0].Sha),
			existingLabels: otherLabels,
			expectedLabels: utils.MapMerge(otherLabels, imageLabels),
			shouldAdd:      true,
		},
		{
			description:    "image with existing labels, some overlap",
			image:          makeImage(scannedImages[0].Name, scannedImages[0].Sha),
			existingLabels: partialLabelSet(),
			expectedLabels: imageLabels,
			shouldAdd:      true,
		},
		{
			description:    "image with exact existing labels",
			image:          makeImage(scannedImages[0].Name, scannedImages[0].Sha),
			existingLabels: imageLabels,
			expectedLabels: imageLabels,
			shouldAdd:      false,
		},
	}

	for _, tc := range testcases {
		annotationObj := makeImageAnnotationObj()
		fullName := fmt.Sprintf("%s@sha256:%s", scannedImages[0].Name, scannedImages[0].Sha)
		tc.image.SetLabels(tc.existingLabels)
		result := createIA().addImageLabels(fullName, tc.image, annotationObj)
		if result != tc.shouldAdd {
			t.Fatalf("[%s] expected %t, got %t", tc.description, tc.shouldAdd, result)
		}

		updated := tc.image.GetLabels()
		for k, v := range tc.expectedLabels {
			if val, ok := updated[k]; !ok {
				t.Errorf("[%s] key %s doesn't exist in image labels %v", tc.description, k, updated)
			} else if val != v {
				t.Errorf("[%s] key %s has wrong value in image label.  Expected %s got %s", tc.description, k, tc.expectedLabels[k], updated[k])
			}
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

		annotator := ImageAnnotator{
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
