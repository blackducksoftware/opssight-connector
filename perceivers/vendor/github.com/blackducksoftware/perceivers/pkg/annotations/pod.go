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
)

// PodAnnotationData describes the data model for pod annotation
type PodAnnotationData struct {
	policyViolationCount int
	vulnerabilityCount   int
	overallStatus        string
	hubVersion           string
	scanClientVersion    string
}

// NewPodAnnotationData creates a new PodAnnotationData object
func NewPodAnnotationData(policyViolationCount int, vulnerabilityCount int, overallStatus string, hubVersion string, scVersion string) *PodAnnotationData {
	return &PodAnnotationData{
		policyViolationCount: policyViolationCount,
		vulnerabilityCount:   vulnerabilityCount,
		overallStatus:        overallStatus,
		hubVersion:           hubVersion,
		scanClientVersion:    scVersion,
	}
}

// HasPolicyViolations returns true if the pod has any policy violations
func (pad *PodAnnotationData) HasPolicyViolations() bool {
	return pad.policyViolationCount > 0
}

// HasVulnerabilities returns true if the pod has any vulnerabilities
func (pad *PodAnnotationData) HasVulnerabilities() bool {
	return pad.vulnerabilityCount > 0
}

// GetVulnerabilityCount returns the number of pod vulnerabilities
func (pad *PodAnnotationData) GetVulnerabilityCount() int {
	return pad.vulnerabilityCount
}

// GetPolicyViolationCount returns the number of pod policy violations
func (pad *PodAnnotationData) GetPolicyViolationCount() int {
	return pad.policyViolationCount
}

// GetOverallStatus returns the pod overall status
func (pad *PodAnnotationData) GetOverallStatus() string {
	return pad.overallStatus
}

// GetHubVersion returns the version of the hub that provided the information
func (pad *PodAnnotationData) GetHubVersion() string {
	return pad.hubVersion
}

// GetScanClientVersion returns the version of the scan client used to scan the images
func (pad *PodAnnotationData) GetScanClientVersion() string {
	return pad.scanClientVersion
}

// CreatePodLabels returns a map of labels from a PodAnnotationData object
func CreatePodLabels(obj interface{}) map[string]string {
	podData := obj.(*PodAnnotationData)
	labels := make(map[string]string)
	labels["pod.policy-violations"] = fmt.Sprintf("%d", podData.GetPolicyViolationCount())
	labels["pod.vulnerabilities"] = fmt.Sprintf("%d", podData.GetVulnerabilityCount())
	labels["pod.overall-status"] = podData.GetOverallStatus()

	return labels
}

// CreatePodAnnotations returns a map of annotations from a PodAnnotationData object
func CreatePodAnnotations(obj interface{}) map[string]string {
	podData := obj.(*PodAnnotationData)
	newAnnotations := make(map[string]string)

	newAnnotations["pod.scanner-version"] = podData.GetScanClientVersion()
	newAnnotations["pod.server-version"] = podData.GetHubVersion()

	return newAnnotations
}
