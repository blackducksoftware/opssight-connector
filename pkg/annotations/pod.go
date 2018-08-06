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

	"github.com/blackducksoftware/perceivers/pkg/annotations"
)

// PodAnnotationPrefix is the prefix used for BlackDuckAnnotations in pod annotations
var PodAnnotationPrefix = "quality.pod.openshift.io"

// CreatePodLabels returns a map of labels from a PodAnnotationData object
//func CreatePodLabels(podData *annotations.PodAnnotationData) map[string]string {
func CreatePodLabels(obj interface{}) map[string]string {
	podData := obj.(*annotations.PodAnnotationData)
	labels := make(map[string]string)
	labels["com.blackducksoftware.pod.policy-violations"] = fmt.Sprintf("%d", podData.GetPolicyViolationCount())
	labels["com.blackducksoftware.pod.has-policy-violations"] = fmt.Sprintf("%t", podData.HasPolicyViolations())
	labels["com.blackducksoftware.pod.vulnerabilities"] = fmt.Sprintf("%d", podData.GetVulnerabilityCount())
	labels["com.blackducksoftware.pod.has-vulnerabilities"] = fmt.Sprintf("%t", podData.HasVulnerabilities())
	labels["com.blackducksoftware.pod.overall-status"] = podData.GetOverallStatus()

	return labels
}

// CreatePodAnnotations returns a map of annotations from a PodAnnotationData object
func CreatePodAnnotations(obj interface{}) map[string]string {
	podData := obj.(*annotations.PodAnnotationData)
	newAnnotations := make(map[string]string)
	vulnAnnotations := CreateBlackDuckVulnerabilityAnnotation(podData.HasVulnerabilities() == true, "", podData.GetVulnerabilityCount(), podData.GetScanClientVersion())
	policyAnnotations := CreateBlackDuckPolicyAnnotation(podData.HasPolicyViolations() == true, "", podData.GetPolicyViolationCount(), podData.GetScanClientVersion())

	newAnnotations[fmt.Sprintf("%s/vulnerability.blackduck", PodAnnotationPrefix)] = vulnAnnotations.AsString()
	newAnnotations[fmt.Sprintf("%s/policy.blackduck", PodAnnotationPrefix)] = policyAnnotations.AsString()

	return newAnnotations
}
