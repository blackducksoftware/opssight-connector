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

package freeway

import (
	"fmt"
	"regexp"
)

type LinkType int

const (
	LinkTypeProjs                LinkType = iota
	LinkTypeVersions             LinkType = iota
	LinkTypeVersion              LinkType = iota
	LinkTypeAssignableUsers      LinkType = iota
	LinkTypeAssignableUserGroups LinkType = iota
	LinkTypeUsers                LinkType = iota
	LinkTypeUserGroups           LinkType = iota
	LinkTypeTags                 LinkType = iota
	LinkTypeVersionReport        LinkType = iota
	LinkTypeLicenseReports       LinkType = iota
	LinkTypeRiskProfile          LinkType = iota
	LinkTypeComponents           LinkType = iota
	LinkTypeVulnerableComponents LinkType = iota
	LinkTypePolicyStatus         LinkType = iota
	LinkTypeCodeLocations        LinkType = iota
	LinkTypeScanSummaries        LinkType = iota
)

func (l LinkType) MarshalJSON() ([]byte, error) {
	jsonString := fmt.Sprintf(`"%s"`, l.String())
	return []byte(jsonString), nil
}

func (l LinkType) MarshalText() (text []byte, err error) {
	return []byte(l.String()), nil
}

var allLinkTypes = []LinkType{
	LinkTypeProjs,
	LinkTypeVersions,
	LinkTypeVersion,
	LinkTypeAssignableUsers,
	LinkTypeAssignableUserGroups,
	LinkTypeUsers,
	LinkTypeUserGroups,
	LinkTypeTags,
	LinkTypeVersionReport,
	LinkTypeLicenseReports,
	LinkTypeRiskProfile,
	LinkTypeComponents,
	LinkTypeVulnerableComponents,
	LinkTypePolicyStatus,
	LinkTypeCodeLocations,
	LinkTypeScanSummaries,
}

func (l LinkType) String() string {
	switch l {
	case LinkTypeProjs:
		return "LinkTypeProjs"
	case LinkTypeVersions:
		return "LinkTypeVersions"
	case LinkTypeVersion:
		return "LinkTypeVersion"
	case LinkTypeAssignableUsers:
		return "LinkTypeAssignableUsers"
	case LinkTypeAssignableUserGroups:
		return "LinkTypeAssignableUserGroups"
	case LinkTypeUsers:
		return "LinkTypeUsers"
	case LinkTypeUserGroups:
		return "LinkTypeUserGroups"
	case LinkTypeTags:
		return "LinkTypeTags"
	case LinkTypeVersionReport:
		return "LinkTypeVersionReport"
	case LinkTypeLicenseReports:
		return "LinkTypeLicenseReports"
	case LinkTypeRiskProfile:
		return "LinkTypeRiskProfile"
	case LinkTypeComponents:
		return "LinkTypeComponents"
	case LinkTypeVulnerableComponents:
		return "LinkTypeVulnerableComponents"
	case LinkTypePolicyStatus:
		return "LinkTypePolicyStatus"
	case LinkTypeCodeLocations:
		return "LinkTypeCodeLocations"
	case LinkTypeScanSummaries:
		return "LinkTypeScanSummaries"
	}
	panic(fmt.Errorf("invalid link type: %d", l))
}

func (l LinkType) Regex() *regexp.Regexp {
	switch l {
	case LinkTypeProjs:
		return regexp.MustCompile("^.*/projects$")
	case LinkTypeVersions:
		return regexp.MustCompile("^.*/projects/.*/versions$")
	case LinkTypeVersion:
		return regexp.MustCompile("^.*/projects/.*/versions/[^/]*$")
	case LinkTypeAssignableUsers:
		return regexp.MustCompile("^.*/projects/.*/assignable-users$")
	case LinkTypeAssignableUserGroups:
		return regexp.MustCompile("^.*/projects/.*/assignable-usergroups$")
	case LinkTypeUsers:
		return regexp.MustCompile("^.*/projects/.*/users$")
	case LinkTypeUserGroups:
		return regexp.MustCompile("^.*/projects/.*/usergroups$")
	case LinkTypeTags:
		return regexp.MustCompile("^.*/projects/.*/tags$")
	case LinkTypeVersionReport:
		return regexp.MustCompile("^.*/versions/.*/reports$")
	case LinkTypeLicenseReports:
		return regexp.MustCompile("^.*/versions/.*/license-reports$")
	case LinkTypeRiskProfile:
		return regexp.MustCompile("^.*/projects/.*/versions/.*/risk-profile$")
	case LinkTypeComponents:
		return regexp.MustCompile("^.*/projects/.*/versions/.*/components$")
	case LinkTypeVulnerableComponents:
		return regexp.MustCompile("^.*/projects/.*/versions/.*/vulnerable-bom-components$")
	case LinkTypePolicyStatus:
		return regexp.MustCompile("^.*/projects/.*/versions/.*/policy-status$")
	case LinkTypeCodeLocations:
		return regexp.MustCompile("^.*/projects/.*/versions/.*/codelocations$")
	case LinkTypeScanSummaries:
		return regexp.MustCompile("^.*/codelocations/.*/scan-summaries$")
	}
	panic(fmt.Errorf("invalid link type: %d", l))
}

func AnalyzeLink(link string) (*LinkType, error) {
	for _, linkType := range allLinkTypes {
		match := linkType.Regex().FindAllString(link, -1)
		if len(match) == 1 {
			return &linkType, nil
		}
	}
	return nil, fmt.Errorf("unable to match link type for %s", link)
}
