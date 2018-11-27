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
	"fmt"

	"github.com/blackducksoftware/perceptor-skyfire/pkg/kube"
)

// MetaReport .....
type MetaReport struct {
	KubeMeta    *kube.Meta
	HubVersions map[string]string
	//	HubScanClientVersion string // TODO we don't need this, do we?
}

// NewMetaReport .....
func NewMetaReport(dump *Dump) *MetaReport {
	hubVersions := map[string]string{}
	for host, dump := range dump.Hubs {
		hubVersions[host] = dump.Version
	}
	return &MetaReport{
		KubeMeta:    dump.Kube.Meta,
		HubVersions: hubVersions,
	}
}

// HumanReadableString .....
func (m *MetaReport) HumanReadableString() string {
	return fmt.Sprintf(`
Overview:
 - Hub versions %+v
 - Kubernetes version %s with build date %s
 - the cluster had %d nodes
`,
		m.HubVersions,
		m.KubeMeta.GitVersion,
		m.KubeMeta.BuildDate,
		m.KubeMeta.NodeCount)
}
