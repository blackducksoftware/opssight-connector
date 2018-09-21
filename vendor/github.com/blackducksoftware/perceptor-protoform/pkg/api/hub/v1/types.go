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

package v1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Hub will be CRD hub definition
// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Hub struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HubSpec   `json:"spec"`
	Status HubStatus `json:"status,omitempty"`
}

// HubSpec will be CRD Hub definition's Spec
type HubSpec struct {
	Namespace       string     `json:"namespace"`
	Flavor          string     `json:"flavor"`
	DockerRegistry  string     `json:"dockerRegistry"`
	DockerRepo      string     `json:"dockerRepo"`
	HubVersion      string     `json:"hubVersion"`
	DbPrototype     string     `json:"dbPrototype"`
	InstanceName    string     `json:"instanceName"`
	BackupInterval  string     `json:"backupInterval"`
	BackupUnit      string     `json:"backupUnit"`
	PVCStorageClass string     `json:"pvcStorageClass"`
	BackupSupport   string     `json:"backupSupport"`
	ScanType        string     `json:"scanType"`
	PVCClaimSize    string     `json:"pvcClaimSize"`
	NFSServer       string     `json:"nfsServer"`
	CertificateName string     `json:"certificateName"`
	Certificate     string     `json:"certificate"`
	CertificateKey  string     `json:"certificateKey"`
	HubType         string     `json:"hubType"`
	State           string     `json:"state"`
	Environs        []Environs `json:"environs"`
}

// Environs will hold the list of Environment variables
type Environs struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// HubStatus will be CRD Hub definition's Status
type HubStatus struct {
	State         string `json:"state"`
	IP            string `json:"ip"`
	PVCVolumeName string `json:"pvcVolumeName"`
	Fqdn          string `json:"fqdn"`
	ErrorMessage  string `json:"errorMessage"`
}

// HubList will store the list of Hubs
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type HubList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`

	Items []Hub `json:"items"`
}
