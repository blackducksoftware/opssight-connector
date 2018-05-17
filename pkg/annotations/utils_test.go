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

package annotations

import (
	"testing"
)

func TestIsBlackDuckEntry(t *testing.T) {
	testcases := []struct {
		description string
		key         string
		retval      bool
	}{
		{
			description: "blackducksoftare.com key",
			key:         "blackducksoftware.com",
			retval:      true,
		},
		{
			description: "blackducksoftare.com key with sub-data",
			key:         "blackducksoftware.com/sub-data",
			retval:      true,
		},
		{
			description: "quality.image.openshift.io key",
			key:         "quality.image.openshift.io",
			retval:      true,
		},
		{
			description: "quality.image.openshift.io key with sub-data",
			key:         "quality.image.openshift.io/sub-data",
			retval:      true,
		},
		{
			description: "non-blackduck related key",
			key:         "shouldbefalse",
			retval:      false,
		},
	}

	for _, tc := range testcases {
		result := isBlackDuckEntry(tc.key)
		if result != tc.retval {
			t.Errorf("[%s] expected %t got %t: key %s", tc.description, tc.retval, result, tc.key)
		}
	}
}

func TestIsBlackDuckAnnotation(t *testing.T) {
	testcases := []struct {
		description string
		key         string
		retval      bool
	}{
		{
			description: "quality.image.openshift.io key",
			key:         "quality.image.openshift.io",
			retval:      true,
		},
		{
			description: "quality.image.openshift.io key with sub-data",
			key:         "quality.image.openshift.io/sub-data",
			retval:      true,
		},
		{
			description: "quality.pod.openshift.io key",
			key:         "quality.pod.openshift.io",
			retval:      true,
		},
		{
			description: "quality.pod.openshift.io key with sub-data",
			key:         "quality.pod.openshift.io/sub-data",
			retval:      true,
		},
		{
			description: "non-blackduck related key",
			key:         "shouldbefalse",
			retval:      false,
		},
	}

	for _, tc := range testcases {
		result := isBlackDuckAnnotation(tc.key)
		if result != tc.retval {
			t.Errorf("[%s] expected %t got %t: key %s", tc.description, tc.retval, result, tc.key)
		}
	}
}

func TestMapContainsBlackDuckEntries(t *testing.T) {
	testcases := []struct {
		description string
		orig        map[string]string
		new         map[string]string
		retval      bool
	}{
		{
			description: "image annotation prefix with same non-json value",
			orig:        map[string]string{ImageAnnotationPrefix: "value", "key1": "value1"},
			new:         map[string]string{ImageAnnotationPrefix: "value"},
			retval:      true,
		},
		{
			description: "image annotation prefix with different non-json value",
			orig:        map[string]string{ImageAnnotationPrefix: "value", "key1": "value1"},
			new:         map[string]string{ImageAnnotationPrefix: "newValue"},
			retval:      false,
		},
		{
			description: "image annotation prefix with same json value",
			orig:        map[string]string{"otherkey": "othervalue", ImageAnnotationPrefix: CreateBlackDuckVulnerabilityAnnotation(true, "url", 20, "1.2.3").AsString()},
			new:         map[string]string{ImageAnnotationPrefix: CreateBlackDuckVulnerabilityAnnotation(true, "url", 20, "1.2.3").AsString()},
			retval:      true,
		},
		{
			description: "image annotation prefix with different json value",
			orig:        map[string]string{"otherkey": "othervalue", ImageAnnotationPrefix: CreateBlackDuckVulnerabilityAnnotation(true, "url", 20, "1.2.3").AsString()},
			new:         map[string]string{ImageAnnotationPrefix: CreateBlackDuckVulnerabilityAnnotation(true, "url", 10, "1.2.3").AsString()},
			retval:      false,
		},
		{
			description: "pod annotation prefix with same non-json value",
			orig:        map[string]string{PodAnnotationPrefix: "value", "key1": "value1"},
			new:         map[string]string{PodAnnotationPrefix: "value"},
			retval:      true,
		},
		{
			description: "pod annotation prefix with different non-json value",
			orig:        map[string]string{PodAnnotationPrefix: "value", "key1": "value1"},
			new:         map[string]string{PodAnnotationPrefix: "newValue"},
			retval:      false,
		},
		{
			description: "pod annotation prefix with same json value",
			orig:        map[string]string{"otherkey": "othervalue", PodAnnotationPrefix: CreateBlackDuckVulnerabilityAnnotation(true, "url", 20, "1.2.3").AsString()},
			new:         map[string]string{PodAnnotationPrefix: CreateBlackDuckVulnerabilityAnnotation(true, "url", 20, "1.2.3").AsString()},
			retval:      true,
		},
		{
			description: "pod annotation prefix with different json value",
			orig:        map[string]string{"otherkey": "othervalue", PodAnnotationPrefix: CreateBlackDuckVulnerabilityAnnotation(true, "url", 20, "1.2.3").AsString()},
			new:         map[string]string{PodAnnotationPrefix: CreateBlackDuckVulnerabilityAnnotation(true, "url", 10, "1.2.3").AsString()},
			retval:      false,
		},
	}

	for _, tc := range testcases {
		result := MapContainsBlackDuckEntries(tc.orig, tc.new)
		if result != tc.retval {
			t.Errorf("[%s] expected %t got %t: orig %v, new %v", tc.description, tc.retval, result, tc.orig, tc.new)
		}
	}
}
