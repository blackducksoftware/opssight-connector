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
	"fmt"
	"strings"
)

// ImageAnnotationData describes the data model for image annotation
type ImageAnnotationData struct {
	policyViolationCount int
	vulnerabilityCount   int
	overallStatus        string
	componentsURL        string
	serverVersion        string
	scanClientVersion    string
}

// NewImageAnnotationData creates a new ImageAnnotationData object
func NewImageAnnotationData(policyViolationCount int, vulnerabilityCount int, overallStatus string, url string, serverVersion string, scVersion string) *ImageAnnotationData {
	return &ImageAnnotationData{
		policyViolationCount: policyViolationCount,
		vulnerabilityCount:   vulnerabilityCount,
		overallStatus:        overallStatus,
		componentsURL:        url,
		serverVersion:        serverVersion,
		scanClientVersion:    scVersion,
	}
}

// HasPolicyViolations returns true if the image has any policy violations
func (iad *ImageAnnotationData) HasPolicyViolations() bool {
	return iad.policyViolationCount > 0
}

// HasVulnerabilities returns true if the image has any vulnerabilities
func (iad *ImageAnnotationData) HasVulnerabilities() bool {
	return iad.vulnerabilityCount > 0
}

// GetVulnerabilityCount returns the number of image vulnerabilities
func (iad *ImageAnnotationData) GetVulnerabilityCount() int {
	return iad.vulnerabilityCount
}

// GetPolicyViolationCount returns the number of image policy violations
func (iad *ImageAnnotationData) GetPolicyViolationCount() int {
	return iad.policyViolationCount
}

// GetComponentsURL returns the image components URL
func (iad *ImageAnnotationData) GetComponentsURL() string {
	return iad.componentsURL
}

// GetOverallStatus returns the image overall status
func (iad *ImageAnnotationData) GetOverallStatus() string {
	return iad.overallStatus
}

// GetServerVersion returns the version of the hub that provided the information
func (iad *ImageAnnotationData) GetServerVersion() string {
	return iad.serverVersion
}

// GetScanClientVersion returns the version of the scan client used to scan the image
func (iad *ImageAnnotationData) GetScanClientVersion() string {
	return iad.scanClientVersion
}

// CreateImageLabels returns a map of labels from a ImageAnnotationData object
func CreateImageLabels(obj interface{}, name string, count int) map[string]string {
	imageData := obj.(*ImageAnnotationData)
	imagePostfix := ""
	labels := make(map[string]string)

	if len(name) > 0 {
		imagePostfix = fmt.Sprintf("%d", count)
		name = strings.Replace(name, "/", ".", -1)
		name = strings.Replace(name, ":", ".", -1)
		labels[fmt.Sprintf("image%d", count)] = name
	}
	labels[fmt.Sprintf("image%s.policy-violations", imagePostfix)] = fmt.Sprintf("%d", imageData.GetPolicyViolationCount())
	labels[fmt.Sprintf("image%s.vulnerabilities", imagePostfix)] = fmt.Sprintf("%d", imageData.GetVulnerabilityCount())
	labels[fmt.Sprintf("image%s.overall-status", imagePostfix)] = imageData.GetOverallStatus()

	return labels
}

// CreateImageAnnotations returns a map of annotations from a ImageAnnotationData object
func CreateImageAnnotations(obj interface{}, name string, count int) map[string]string {
	imageData := obj.(*ImageAnnotationData)
	imagePrefix := ""
	newAnnotations := make(map[string]string)

	if len(name) > 0 {
		imagePrefix = fmt.Sprintf("image%d", count)
		imageName := strings.Replace(name, "/", ".", -1)
		newAnnotations[fmt.Sprintf("%s", imagePrefix)] = imageName
		imagePrefix = imagePrefix + "."
	}
	newAnnotations[fmt.Sprintf("%spolicy-violations", imagePrefix)] = fmt.Sprintf("%d", imageData.GetPolicyViolationCount())
	newAnnotations[fmt.Sprintf("%svulnerabilities", imagePrefix)] = fmt.Sprintf("%d", imageData.GetVulnerabilityCount())
	newAnnotations[fmt.Sprintf("%soverall-status", imagePrefix)] = imageData.GetOverallStatus()
	newAnnotations[fmt.Sprintf("%sscanner-version", imagePrefix)] = imageData.GetScanClientVersion()
	newAnnotations[fmt.Sprintf("%sserver-version", imagePrefix)] = imageData.GetServerVersion()
	newAnnotations[fmt.Sprintf("%sproject-endpoint", imagePrefix)] = imageData.GetComponentsURL()

	return newAnnotations
}
