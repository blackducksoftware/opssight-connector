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
	"encoding/json"
	"reflect"
	"time"
)

type summaryEntry struct {
	Label         string `json:"label"`
	Score         int    `json:"score"`
	SeverityIndex int    `json:"severityIndex"`
}

// BlackDuckAnnotation create annotations that correspond to the
// Openshift Container Security guide (https://people.redhat.com/aweiteka/docs/preview/20170510/security/container_content.html)
type BlackDuckAnnotation struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Timestamp   int64          `json:"timestamp"`
	Reference   string         `json:"reference"`
	Compliant   bool           `json:"compliant"`
	Summary     []summaryEntry `json:"summary"`
}

// AsString makes a map corresponding to the Openshift
// Container Security guide (https://people.redhat.com/aweiteka/docs/preview/20170510/security/container_content.html)
func (bda *BlackDuckAnnotation) AsString() string {
	mp, _ := json.Marshal(bda)
	return string(mp)
}

// Compare checks if the passed in BlackDuckAnnotation contains the same
// values while ignoring fields that will be different (like timestamp).
// Returns true if the values are the same, false otherwise
func (bda *BlackDuckAnnotation) Compare(newBda *BlackDuckAnnotation) bool {
	if bda.Name != newBda.Name {
		return false
	}
	if bda.Description != newBda.Description {
		return false
	}
	if bda.Reference != newBda.Reference {
		return false
	}
	if bda.Compliant != newBda.Compliant {
		return false
	}
	for pos, summaryEntry := range bda.Summary {
		if !reflect.DeepEqual(summaryEntry, newBda.Summary[pos]) {
			return false
		}
	}

	return true
}

// NewBlackDuckAnnotationFromJSON takes a string that is a marshaled
// BlackDuckAnnotation struct and returns a BlackDuckAnnotation
func NewBlackDuckAnnotationFromJSON(data string) (*BlackDuckAnnotation, error) {
	bda := BlackDuckAnnotation{}
	err := json.Unmarshal([]byte(data), &bda)
	if err != nil {
		return nil, err
	}

	return &bda, nil
}

// CreateBlackDuckVulnerabilityAnnotation returns an annotation containing
// vulnerabilities
func CreateBlackDuckVulnerabilityAnnotation(hasVulns bool, url string, vulnCount int) *BlackDuckAnnotation {
	return &BlackDuckAnnotation{
		"BlackDucksoftware",
		"Vulnerability Info",
		time.Now().Unix(),
		url,
		!hasVulns, // no vunls -> compliant.
		[]summaryEntry{
			{
				Label:         "high",
				Score:         vulnCount,
				SeverityIndex: 1,
			},
		},
	}
}

// CreateBlackDuckPolicyAnnotation returns an annotation containing
// policy violations
func CreateBlackDuckPolicyAnnotation(hasPolicyViolations bool, url string, policyCount int) *BlackDuckAnnotation {
	return &BlackDuckAnnotation{
		"BlackDucksoftware",
		"Policy Info",
		time.Now().Unix(),
		url,
		!hasPolicyViolations, // no violations -> compliant
		[]summaryEntry{
			{
				Label:         "important",
				Score:         policyCount,
				SeverityIndex: 1,
			},
		},
	}
}

// CompareBlackDuckAnnotationJSON takes 2 strings that are marshaled
// BlackDuckAnnotations and compares them.  Returns true if the unmarshaling
// is successful and the values are the same.
func CompareBlackDuckAnnotationJSON(old string, new string) bool {
	bda1, err := NewBlackDuckAnnotationFromJSON(old)
	if err != nil {
		return false
	}

	bda2, err := NewBlackDuckAnnotationFromJSON(new)
	if err != nil {
		return false
	}

	return bda1.Compare(bda2)
}
