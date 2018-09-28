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

// OpsSight is a specification for a OpsSight resource
type OpsSight struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpsSightSpec   `json:"spec"`
	Status OpsSightStatus `json:"status,omitempty"`
}

// RegistryAuth will store the Openshift Internal Registries
type RegistryAuth struct {
	URL      string `json:"Url"`
	User     string
	Password string
}

// OpsSightSpec is the spec for a OpsSight resource
type OpsSightSpec struct {
	Namespace string `json:"namespace,omitempty"`
	State     string `json:"state"`
	// CONTAINER CONFIGS
	// These are sed replaced into the config maps for the containers.
	PerceptorPort                         *int           `json:"perceptorPort,omitempty"`
	ScannerPort                           *int           `json:"scannerPort,omitempty"`
	PerceiverPort                         *int           `json:"perceiverPort,omitempty"`
	ImageFacadePort                       *int           `json:"imageFacadePort,omitempty"`
	SkyfirePort                           *int           `json:"skyfirePort,omitempty"`
	InternalRegistries                    []RegistryAuth `json:"internalRegistries,omitempty"`
	AnnotationIntervalSeconds             *int           `json:"annotationIntervalSeconds,omitempty"`
	DumpIntervalMinutes                   *int           `json:"dumpIntervalMinutes,omitempty"`
	HubHost                               string         `json:"hubHost,omitempty"`
	HubUser                               string         `json:"hubUser,omitempty"`
	HubPort                               *int           `json:"hubPort,omitempty"`
	HubUserPassword                       string         `json:"hubUserPassword,omitempty"`
	HubClientTimeoutPerceptorMilliseconds *int           `json:"hubClientTimeoutPerceptorMilliseconds,omitempty"`
	HubClientTimeoutScannerSeconds        *int           `json:"hubClientTimeoutScannerSeconds,omitempty"`
	ConcurrentScanLimit                   *int           `json:"concurrentScanLimit,omitempty"`
	TotalScanLimit                        *int           `json:"totalScanLimit,omitempty"`
	CheckForStalledScansPauseHours        *int           `json:"checkForStalledScansPauseHours"`
	StalledScanClientTimeoutHours         *int           `json:"stalledScanClientTimeoutHours"`
	ModelMetricsPauseSeconds              *int           `json:"modelMetricsPauseSeconds"`
	UnknownImagePauseMilliseconds         *int           `json:"unknownImagePauseMilliseconds"`

	// CONTAINER PULL CONFIG
	// These are for defining docker registry and image location and versions
	DefaultVersion string `json:"defaultVersion,omitempty"`
	Registry       string `json:"registry,omitempty"`
	ImagePath      string `json:"imagePath,omitempty"`

	PerceptorImageName      string `json:"perceptorImageName,omitempty"`
	ScannerImageName        string `json:"scannerImageName,omitempty"`
	PodPerceiverImageName   string `json:"podPerceiverImageName,omitempty"`
	ImagePerceiverImageName string `json:"imagePerceiverImageName,omitempty"`
	ImageFacadeImageName    string `json:"imageFacadeImageName,omitempty"`
	SkyfireImageName        string `json:"skyfireImageName,omitempty"`

	PerceptorImageVersion   string `json:"perceptorImageVersion,omitempty"`
	ScannerImageVersion     string `json:"scannerImageVersion,omitempty"`
	PerceiverImageVersion   string `json:"perceiverImageVersion,omitempty"`
	ImageFacadeImageVersion string `json:"imageFacadeImageVersion,omitempty"`
	SkyfireImageVersion     string `json:"skyfireImageVersion,omitempty"`

	ServiceAccounts  map[string]string `json:"serviceAccounts,omitempty"`
	ContainerNames   map[string]string `json:"names,omitempty"`
	ImagePerceiver   *bool             `json:"imagePerceiver,omitempty"`
	PodPerceiver     *bool             `json:"podPerceiver,omitempty"`
	Metrics          *bool             `json:"metrics,omitempty"`
	PerceptorSkyfire *bool             `json:"perceptorSkyfire,omitempty"`
	NamespaceFilter  string            `json:"namespaceFilter,omitempty"`

	// CPU and memory configurations
	// Should be passed like: e.g. "300m"
	DefaultCPU string `json:"defaultCpu,omitempty"`
	// Should be passed like: e.g "1300Mi"
	DefaultMem string `json:"defaultMem,omitempty"`

	// Log level
	LogLevel string `json:"logLevel,omitempty"`

	// Environment Variables
	HubUserPasswordEnvVar string `json:"hubuserPasswordEnvVar"`

	// Configuration secret
	SecretName  string `json:"secretName"`
	UseMockMode *bool  `json:"useMockMode"`
}

// OpsSightStatus is the status for a OpsSight resource
type OpsSightStatus struct {
	State        string `json:"state"`
	ErrorMessage string `json:"errorMessage"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OpsSightList is a list of OpsSight resources
type OpsSightList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`

	Items []OpsSight `json:"items"`
}
