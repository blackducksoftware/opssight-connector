/*
Copyright (C) 2018 Black Duck Software, Inc.

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

package kube

import "fmt"

type PodImageAnnotationKey int

const (
	PodImageAnnotationKeyVulnerabilities  PodImageAnnotationKey = iota
	PodImageAnnotationKeyPolicyViolations PodImageAnnotationKey = iota
	PodImageAnnotationKeyOverallStatus    PodImageAnnotationKey = iota
	PodImageAnnotationKeyServerVersion    PodImageAnnotationKey = iota
	PodImageAnnotationKeyScannerVersion   PodImageAnnotationKey = iota
	PodImageAnnotationKeyProjectEndpoint  PodImageAnnotationKey = iota
	PodImageAnnotationKeyImage            PodImageAnnotationKey = iota
)

func (pak PodImageAnnotationKey) formatString() string {
	switch pak {
	case PodImageAnnotationKeyVulnerabilities:
		return "image%d.vulnerabilities"
	case PodImageAnnotationKeyPolicyViolations:
		return "image%d.policy-violations"
	case PodImageAnnotationKeyOverallStatus:
		return "image%d.overall-status"
	case PodImageAnnotationKeyServerVersion:
		return "image%d.server-version"
	case PodImageAnnotationKeyScannerVersion:
		return "image%d.scanner-version"
	case PodImageAnnotationKeyProjectEndpoint:
		return "image%d.project-endpoint"
	case PodImageAnnotationKeyImage:
		return "image%d"
	}
	panic(fmt.Errorf("invalid PodImageAnnotationKey value: %d", pak))
}

var podImageAnnotationKeys = []PodImageAnnotationKey{
	PodImageAnnotationKeyVulnerabilities,
	PodImageAnnotationKeyPolicyViolations,
	PodImageAnnotationKeyOverallStatus,
	PodImageAnnotationKeyServerVersion,
	PodImageAnnotationKeyScannerVersion,
	PodImageAnnotationKeyProjectEndpoint,
	PodImageAnnotationKeyImage,
}

func (pak PodImageAnnotationKey) String(index int) string {
	return fmt.Sprintf(pak.formatString(), index)
}

func podImageAnnotationKeyStrings(index int) []string {
	strs := []string{}
	for _, key := range podImageAnnotationKeys {
		strs = append(strs, key.String(index))
	}
	return strs
}
