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

// Blackduck will be CRD blackduck definition
// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Blackduck struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata,omitempty"`
	View               BlackduckView   `json:"view"`
	Spec               BlackduckSpec   `json:"spec"`
	Status             BlackduckStatus `json:"status,omitempty"`
}

// BlackduckView will be used to populate information for the Blackduck UI.
type BlackduckView struct {
	Clones            map[string]string `json:"clones"`
	StorageClasses    map[string]string `json:"storageClasses"`
	CertificateNames  []string          `json:"certificateNames"`
	Environs          []string          `json:"environs"`
	ContainerTags     []string          `json:"containerTags"`
	Version           string            `json:"version"`
	SupportedVersions []string          `json:"supportedVersions"`
}

// BlackduckSpec will be CRD Blackduck definition's Spec
type BlackduckSpec struct {
	Namespace         string                    `json:"namespace"`
	Size              string                    `json:"size"`
	Version           string                    `json:"version"`
	ExposeService     string                    `json:"exposeService"`
	DbPrototype       string                    `json:"dbPrototype,omitempty"`
	ExternalPostgres  *PostgresExternalDBConfig `json:"externalPostgres,omitempty"`
	PVCStorageClass   string                    `json:"pvcStorageClass,omitempty"`
	LivenessProbes    bool                      `json:"livenessProbes"`
	ScanType          string                    `json:"scanType,omitempty"`
	PersistentStorage bool                      `json:"persistentStorage"`
	PVC               []PVC                     `json:"pvc,omitempty"`
	CertificateName   string                    `json:"certificateName"`
	Certificate       string                    `json:"certificate,omitempty"`
	CertificateKey    string                    `json:"certificateKey,omitempty"`
	ProxyCertificate  string                    `json:"proxyCertificate,omitempty"`
	AuthCustomCA      string                    `json:"authCustomCa"`
	Type              string                    `json:"type,omitempty"`
	DesiredState      string                    `json:"desiredState"`
	Environs          []string                  `json:"environs,omitempty"`
	ImageRegistries   []string                  `json:"imageRegistries,omitempty"`
	ImageUIDMap       map[string]int64          `json:"imageUidMap,omitempty"`
	LicenseKey        string                    `json:"licenseKey,omitempty"`
}

// Environs will hold the list of Environment variables
type Environs struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PVC will contain the specifications of the different PVC.
// This will overwrite the default claim configuration
type PVC struct {
	Name         string `json:"name"`
	Size         string `json:"size,omitempty"`
	StorageClass string `json:"storageClass,omitempty"`
}

// PostgresExternalDBConfig contain the external database configuration
type PostgresExternalDBConfig struct {
	PostgresHost          string `json:"postgresHost"`
	PostgresPort          int    `json:"postgresPort"`
	PostgresAdmin         string `json:"postgresAdmin"`
	PostgresUser          string `json:"postgresUser"`
	PostgresSsl           bool   `json:"postgresSsl"`
	PostgresAdminPassword string `json:"postgresAdminPassword"`
	PostgresUserPassword  string `json:"postgresUserPassword"`
}

// BlackduckStatus will be CRD Blackduck definition's Status
type BlackduckStatus struct {
	State         string            `json:"state"`
	IP            string            `json:"ip"`
	PVCVolumeName map[string]string `json:"pvcVolumeName,omitempty"`
	Fqdn          string            `json:"fqdn,omitempty"`
	ErrorMessage  string            `json:"errorMessage,omitempty"`
}

// BlackduckList will store the list of Blackducks
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type BlackduckList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`

	Items []Blackduck `json:"items"`
}
