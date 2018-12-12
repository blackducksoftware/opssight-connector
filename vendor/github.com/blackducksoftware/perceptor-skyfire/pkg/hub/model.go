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

package hub

import "github.com/blackducksoftware/hub-client-go/hubapi"

// Project .....
type Project struct {
	Name        string
	Versions    []*Version
	Description string
	Source      string
}

// Version .....
type Version struct {
	Name            string
	CodeLocations   []*CodeLocation
	RiskProfile     *RiskProfile
	Distribution    string
	Meta            hubapi.Meta
	ReleasedOn      string
	ReleaseComments string
	Phase           string
	Nickname        string
	PolicyStatus    *PolicyStatus
}

// PolicyStatus .....
type PolicyStatus struct {
	ComponentVersionStatusCounts []*ComponentVersionStatusCount
	Meta                         hubapi.Meta
	OverallStatus                string
	UpdatedAt                    string
}

// ComponentVersionStatusCount .....
type ComponentVersionStatusCount struct {
	Name  string
	Value int
}

// CodeLocation .....
type CodeLocation struct {
	ScanSummaries        []*ScanSummary
	CreatedAt            string
	MappedProjectVersion string
	Meta                 hubapi.Meta
	Name                 string
	Type                 string
	URL                  string
	UpdatedAt            string
}

// RiskProfile .....
type RiskProfile struct {
	BomLastUpdatedAt string
	Categories       map[string]map[string]int
	Meta             hubapi.Meta
}

// ScanSummary .....
type ScanSummary struct {
	CreatedAt string
	Meta      hubapi.Meta
	Status    string
	UpdatedAt string
}
