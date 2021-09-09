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

package util

import (
	"fmt"
	"regexp"
)

var imageTagRegexp = regexp.MustCompile(`([0-9a-zA-Z-_:\\.]*)/([0-9a-zA-Z-_:\\.]*):([a-zA-Z0-9-\\._]+)$`)
var imageVersionRegexp = regexp.MustCompile(`([0-9]+).([0-9]+).([0-9]+)$`)

// ValidateImageString takes a docker image string and returns substring submatch if it's valid, and an error if it is not; Example:
// image := "docker.io/blackducksoftware/synopsys-operator:latest"
// tagMatch = [blackducksoftware/synopsys-operator:latest blackducksoftware synopsys-operator latest]
func ValidateImageString(image string) ([]string, error) {
	imageSubstringSubmatch := imageTagRegexp.FindStringSubmatch(image)
	if len(imageSubstringSubmatch) == 4 {
		return imageSubstringSubmatch, nil
	}
	return []string{}, fmt.Errorf("unable to parse the input image %s for regex %+v", image, imageTagRegexp)
}

// ValidateImageVersion takes a docker image version string and returns substring submatch version if it's valid, and an error if it is not; Example:
// version := "2019.4.2"
// versionMatch = [2019.4.2 2019 4 2]
func ValidateImageVersion(version string) ([]string, error) {
	versionSubstringSubmatch := imageVersionRegexp.FindStringSubmatch(version)
	if len(versionSubstringSubmatch) == 4 {
		return versionSubstringSubmatch, nil
	}
	return []string{}, fmt.Errorf("unable to parse the version %s for regex %+v", version, imageVersionRegexp)
}

// GetImageTag takes a docker image string and returns the tag
func GetImageTag(image string) (string, error) {
	imageSubstringSubmatch, err := ValidateImageString(image)
	if err != nil {
		return "", err
	}
	tag := imageSubstringSubmatch[len(imageSubstringSubmatch)-1]
	return tag, nil
}

// GetImageName takes a docker image string and returns the name
func GetImageName(image string) (string, error) {
	imageSubstringSubmatch, err := ValidateImageString(image)
	if err != nil {
		return "", err
	}
	name := imageSubstringSubmatch[len(imageSubstringSubmatch)-2]
	return name, nil
}
