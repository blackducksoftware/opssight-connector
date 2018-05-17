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

package report

import (
	"strings"
)

type Report struct {
	Dump          *Dump
	Meta          *MetaReport
	Kube          *KubeReport
	KubePerceptor *KubePerceptorReport
	PerceptorHub  *PerceptorHubReport
	Hub           *HubReport
}

func NewReport(dump *Dump) *Report {
	return &Report{
		dump,
		NewMetaReport(dump),
		NewKubeReport(dump.Kube),
		NewKubePerceptorReport(dump),
		NewPerceptorHubReport(dump),
		NewHubReport(dump.Hub),
	}
}

func (r *Report) HumanReadableString() string {
	chunks := []string{
		r.Meta.HumanReadableString(),
		r.Kube.HumanReadableString(),
		r.KubePerceptor.HumanReadableString(),
		r.PerceptorHub.HumanReadableString(),
		r.Hub.HumanReadableString(),
	}
	return strings.Join(chunks, "\n\n")
}

// In perceptor but not in hub:
// - completed images
// - completed pods

// In hub but not in perceptor:
// - completed image

// Extra hub stuff:
// - multiple projects matching a sha (?)
// - multiple project versions in a project
// - multiple scan summaries in a project version
// - multiple code locations in a scan summary
