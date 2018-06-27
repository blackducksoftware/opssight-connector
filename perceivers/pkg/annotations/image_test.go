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
	"strings"
	"testing"

	"github.com/blackducksoftware/perceivers/pkg/annotations"
)

func TestCreateImageLabels(t *testing.T) {
	registry := "registry.name.com"
	registryWithPort := "registry:1000"
	shortImageName := "short/imagename"
	longImageName := "this/isareally/long/imagename/that/islonger/than/allowed/63characters"

	testcases := []struct {
		description string
		imageName   string
		expected    map[string]string
	}{
		{
			description: "image name includes a registry, but name is under than 63 characters",
			imageName:   registry + "/" + shortImageName,
			expected:    map[string]string{"com.blackducksoftware.image0": RemoveRegistryInfo(shortImageName)},
		},
		{
			description: "image name includes a registry with a port, but name is under than 63 characters",
			imageName:   registryWithPort + "/" + shortImageName,
			expected:    map[string]string{"com.blackducksoftware.image0": RemoveRegistryInfo(shortImageName)},
		},
		{
			description: "image name includes a registry, and name is longer than 63 characters",
			imageName:   registry + "/" + longImageName,
			expected:    map[string]string{"com.blackducksoftware.image0": RemoveRegistryInfo(longImageName[0:63])},
		},
		{
			description: "image name is longer than 63 characters",
			imageName:   longImageName,
			expected:    map[string]string{"com.blackducksoftware.image0": RemoveRegistryInfo(longImageName[0:63])},
		},
		{
			description: "image name is shorter than 63 characters",
			imageName:   shortImageName,
			expected:    map[string]string{"com.blackducksoftware.image0": RemoveRegistryInfo(shortImageName)},
		},
	}

	for _, tc := range testcases {
		obj := annotations.NewImageAnnotationData(2, 10, "NOT_IN_VIOLATION", "http://url/ofthe/hub/scan", "1.1.1", "1.1.1")
		result := CreateImageLabels(obj, tc.imageName, 0)
		for k, v := range tc.expected {
			if val, ok := result[k]; !ok {
				t.Errorf("[%s] created labels missing key %s, labels %v", tc.description, k, result)
			} else {
				if strings.Compare(v, val) != 0 {
					t.Errorf("[%s] created label value %s differs from expected %s", tc.description, val, v)
				}
				if len(val) > 63 {
					t.Errorf("[%s] key %s has value %s, which is longer than 62 characters", tc.description, k, val)
				}
			}
		}
	}
}
