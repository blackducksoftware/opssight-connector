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

// PodImageLabelKey .....
type PodImageLabelKey int

const (
	PodImageLabelKeyVulnerabilities  PodImageLabelKey = iota
	PodImageLabelKeyPolicyViolations PodImageLabelKey = iota
	PodImageLabelKeyOverallStatus    PodImageLabelKey = iota
	PodImageLabelKeyImage            PodImageLabelKey = iota
)

func (pak PodImageLabelKey) formatString() string {
	switch pak {
	case PodImageLabelKeyVulnerabilities:
		return "image%d.vulnerabilities"
	case PodImageLabelKeyPolicyViolations:
		return "image%d.policy-violations"
	case PodImageLabelKeyOverallStatus:
		return "image%d.overall-status"
	case PodImageLabelKeyImage:
		return "image%d"
	}
	panic(fmt.Errorf("invalid PodImageLabelKey value: %d", pak))
}

var podImageLabelKeys = []PodImageLabelKey{
	PodImageLabelKeyVulnerabilities,
	PodImageLabelKeyPolicyViolations,
	PodImageLabelKeyOverallStatus,
	PodImageLabelKeyImage,
}

// String .....
func (pak PodImageLabelKey) String(index int) string {
	return fmt.Sprintf(pak.formatString(), index)
}

func podImageLabelKeyStrings(index int) []string {
	strs := []string{}
	for _, key := range podImageLabelKeys {
		strs = append(strs, key.String(index))
	}
	return strs
}
