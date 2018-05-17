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

import (
	"fmt"
	"time"

	"github.com/blackducksoftware/hub-client-go/hubapi"
	"github.com/blackducksoftware/hub-client-go/hubclient"
	log "github.com/sirupsen/logrus"
)

type HubDumper struct {
	HubClient   *hubclient.Client
	HubUsername string
	HubPassword string
}

func NewHubDumper(hubHost string, username string, password string) (*HubDumper, error) {
	var baseURL = fmt.Sprintf("https://%s", hubHost)
	hubClient, err := hubclient.NewWithSession(baseURL, hubclient.HubClientDebugTimings, 5000*time.Second)
	if err != nil {
		log.Errorf("unable to get hub client: %s", err.Error())
		return nil, err
	}
	dumper := &HubDumper{HubClient: hubClient, HubUsername: username, HubPassword: password}
	err = dumper.Login()
	if err != nil {
		log.Errorf("unable to log in to hub: %s", err.Error())
		return nil, err
	}
	return dumper, nil
}

func (hd *HubDumper) Login() error {
	return hd.HubClient.Login(hd.HubUsername, hd.HubPassword)
}

func (hd *HubDumper) Dump() (*Dump, error) {
	hubProjects, err := hd.DumpAllProjects()
	if err != nil {
		return nil, err
	}
	hubVersion, err := hd.Version()
	if err != nil {
		return nil, err
	}
	return NewDump(hubVersion, hubProjects), nil
}

func (hd *HubDumper) Version() (string, error) {
	version, err := hd.HubClient.CurrentVersion()
	if err != nil {
		return "", err
	}
	return version.Version, nil
}

func (hd *HubDumper) DumpAllProjects() ([]*Project, error) {
	limit := 20000 // totally arbitrary number, just needs to be higher than the
	// number of projects in the hub.  20000 is so high as to be effectively
	// infinite, due to the amount of time it takes to issue 20000 * 5 http requests
	projectList, err := hd.HubClient.ListProjects(&hubapi.GetListOptions{Limit: &limit})
	if err != nil {
		return nil, err
	}
	projects := []*Project{}
	for _, hubProject := range projectList.Items {
		var project *Project
		for {
			project, err = hd.DumpProject(&hubProject)
			if err == nil {
				break
			}
			log.Errorf("unable to dump project %+v: %s", hubProject, err.Error())
			time.Sleep(2 * time.Second)
		}
		projects = append(projects, project)
		time.Sleep(1 * time.Second)
	}
	return projects, nil
}

func (hd *HubDumper) DumpProject(hubProject *hubapi.Project) (*Project, error) {
	log.Debugf("looking for project %s at url %s", hubProject.Name, hubProject.Meta.Href)
	versions := []*Version{}
	versionsLink, err := hubProject.GetProjectVersionsLink()
	if err != nil {
		return nil, err
	}
	versionsList, err := hd.HubClient.ListProjectVersions(*versionsLink, nil)
	if err != nil {
		return nil, err
	}
	for _, hubVersion := range versionsList.Items {
		version, err := hd.DumpVersion(&hubVersion)
		if err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}
	log.Debugf("successfully dumped project %s at url %s", hubProject.Name, hubProject.Meta.Href)
	project := &Project{
		Name:        hubProject.Name,
		Versions:    versions,
		Description: hubProject.Description,
		Source:      hubProject.Source}
	return project, nil
}

func (hd *HubDumper) DumpVersion(hubVersion *hubapi.ProjectVersion) (*Version, error) {
	riskProfileLink, err := hubVersion.GetProjectVersionRiskProfileLink()
	if err != nil {
		return nil, err
	}
	hubRiskProfile, err := hd.HubClient.GetProjectVersionRiskProfile(*riskProfileLink)
	if err != nil {
		return nil, err
	}
	riskProfile, err := hd.DumpRiskProfile(hubRiskProfile)

	codeLocations := []*CodeLocation{}
	codeLocationsLink, err := hubVersion.GetCodeLocationsLink()
	if err != nil {
		return nil, err
	}
	hubCodeLocations, err := hd.HubClient.ListCodeLocations(*codeLocationsLink, nil)
	if err != nil {
		return nil, err
	}
	for _, hubCodeLocation := range hubCodeLocations.Items {
		codeLocation, err := hd.DumpCodeLocation(&hubCodeLocation)
		if err != nil {
			return nil, err
		}
		codeLocations = append(codeLocations, codeLocation)
	}

	policyStatusLink, err := hubVersion.GetProjectVersionPolicyStatusLink()
	if err != nil {
		return nil, err
	}
	hubPolicyStatus, err := hd.HubClient.GetProjectVersionPolicyStatus(*policyStatusLink)
	if err != nil {
		return nil, err
	}
	policyStatus, err := hd.DumpPolicyStatus(hubPolicyStatus)

	version := &Version{
		Name:            hubVersion.VersionName,
		CodeLocations:   codeLocations,
		RiskProfile:     riskProfile,
		Distribution:    hubVersion.Distribution,
		Meta:            hubVersion.Meta,
		Nickname:        hubVersion.Nickname,
		ReleasedOn:      hubVersion.ReleasedOn,
		ReleaseComments: hubVersion.ReleaseComments,
		PolicyStatus:    policyStatus,
		Phase:           hubVersion.Phase,
	}
	return version, nil
}

func (hd *HubDumper) DumpPolicyStatus(hubPolicyStatus *hubapi.ProjectVersionPolicyStatus) (*PolicyStatus, error) {
	statusCounts := []*ComponentVersionStatusCount{}
	for _, hubStatusCount := range hubPolicyStatus.ComponentVersionStatusCounts {
		statusCount := &ComponentVersionStatusCount{
			Name:  hubStatusCount.Name,
			Value: hubStatusCount.Value,
		}
		statusCounts = append(statusCounts, statusCount)
	}
	policyStatus := &PolicyStatus{
		ComponentVersionStatusCounts: statusCounts,
		Meta:          hubPolicyStatus.Meta,
		OverallStatus: hubPolicyStatus.OverallStatus,
		UpdatedAt:     hubPolicyStatus.UpdatedAt,
	}
	return policyStatus, nil
}

func (hd *HubDumper) DumpCodeLocation(hubCodeLocation *hubapi.CodeLocation) (*CodeLocation, error) {
	scanSummaries := []*ScanSummary{}
	link, err := hubCodeLocation.GetScanSummariesLink()
	if err != nil {
		return nil, err
	}
	hubScanSummaries, err := hd.HubClient.ListScanSummaries(*link)
	if err != nil {
		return nil, err
	}
	for _, hubScanSummary := range hubScanSummaries.Items {
		scanSummary, err := hd.DumpScanSummary(&hubScanSummary)
		if err != nil {
			return nil, err
		}
		scanSummaries = append(scanSummaries, scanSummary)
	}
	codeLocation := &CodeLocation{
		CreatedAt:            hubCodeLocation.CreatedAt,
		MappedProjectVersion: hubCodeLocation.MappedProjectVersion,
		Meta:                 hubCodeLocation.Meta,
		Name:                 hubCodeLocation.Name,
		ScanSummaries:        scanSummaries,
		Type:                 hubCodeLocation.Type,
		URL:                  hubCodeLocation.URL,
		UpdatedAt:            hubCodeLocation.UpdatedAt,
	}
	return codeLocation, nil
}

func (hd *HubDumper) DumpRiskProfile(hubRiskProfile *hubapi.ProjectVersionRiskProfile) (*RiskProfile, error) {
	riskProfile := &RiskProfile{
		BomLastUpdatedAt: hubRiskProfile.BomLastUpdatedAt,
		Categories:       hubRiskProfile.Categories,
		Meta:             hubRiskProfile.Meta,
	}
	return riskProfile, nil
}

func (hd *HubDumper) DumpScanSummary(hubScanSummary *hubapi.ScanSummary) (*ScanSummary, error) {
	scanSummary := &ScanSummary{
		CreatedAt: hubScanSummary.CreatedAt,
		Meta:      hubScanSummary.Meta,
		Status:    hubScanSummary.Status,
		UpdatedAt: hubScanSummary.UpdatedAt,
	}
	return scanSummary, nil
}

//func dumpComponent(hubComponent *hubapi.)
