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

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Alert is a specification for a Alert resource
type Alert struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlertSpec   `json:"spec"`
	Status AlertStatus `json:"status,omitempty"`
}

// AlertSpec is the spec for a Alert resource
type AlertSpec struct {
	Namespace         string `json:"namespace,omitempty"`
	Registry          string `json:"registry,omitempty"`
	ImagePath         string `json:"imagePath,omitempty"`
	AlertImageName    string `json:"alertImageName,omitempty"`
	AlertImageVersion string `json:"alertImageVersion,omitempty"`
	CfsslImageName    string `json:"cfsslImageName,omitempty"`
	CfsslImageVersion string `json:"cfsslImageVersion,omitempty"`
	HubHost           string `json:"hubHost,omitempty"`
	HubUser           string `json:"hubUser,omitempty"`
	HubPort           *int   `json:"hubPort,omitempty"`
	Port              *int   `json:"port"`
	StandAlone        *bool  `json:"standAlone"`

	// Should be passed like: e.g "1300Mi"
	AlertMemory string `json:"alertMemory.omitempty"`
	CfsslMemory string `json:"cfsslMemory.omitempty"`
	State       string `json:"state"`
}

// AlertStatus is the status for a Alert resource
type AlertStatus struct {
	State        string `json:"state"`
	ErrorMessage string `json:"errorMessage"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AlertList is a list of Alert resources
type AlertList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`

	Items []Alert `json:"items"`
}
