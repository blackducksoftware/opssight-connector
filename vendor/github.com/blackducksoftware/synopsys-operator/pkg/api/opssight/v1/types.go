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
	"github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
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
	User     string `json:"user"`
	Password string `json:"password"`
}

// Blackduck ...
type Blackduck struct {
	Hosts               []string `json:"hosts"`
	User                string   `json:"user"`
	Port                int      `json:"port"`
	ConcurrentScanLimit int      `json:"concurrentScanLimit"`
	TotalScanLimit      int      `json:"totalScanLimit"`
	PasswordEnvVar      string   `json:"passwordEnvVar"`
	// Auto scaling parameters
	InitialCount                 int               `json:"initialCount"`
	MaxCount                     int               `json:"maxCount"`
	DeleteHubThresholdPercentage int               `json:"deleteHubThresholdPercentage"`
	BlackduckSpec                *v1.BlackduckSpec `json:"blackduckSpec"`
}

// Perceptor ...
type Perceptor struct {
	Name                           string `json:"name"`
	Image                          string `json:"image"`
	Port                           int    `json:"port"`
	CheckForStalledScansPauseHours int    `json:"checkForStalledScansPauseHours"`
	StalledScanClientTimeoutHours  int    `json:"stalledScanClientTimeoutHours"`
	ModelMetricsPauseSeconds       int    `json:"modelMetricsPauseSeconds"`
	UnknownImagePauseMilliseconds  int    `json:"unknownImagePauseMilliseconds"`
	ClientTimeoutMilliseconds      int    `json:"clientTimeoutMilliseconds"`
}

// ScannerPod ...
type ScannerPod struct {
	Name           string       `json:"name"`
	Scanner        *Scanner     `json:"scanner"`
	ImageFacade    *ImageFacade `json:"imageFacade"`
	ReplicaCount   int          `json:"scannerReplicaCount"`
	ImageDirectory string       `json:"imageDirectory"`
}

// Scanner ...
type Scanner struct {
	Name                 string `json:"name"`
	Image                string `json:"image"`
	Port                 int    `json:"port"`
	ClientTimeoutSeconds int    `json:"clientTimeoutSeconds"`
}

// ImageFacade ...
type ImageFacade struct {
	Name               string         `json:"name"`
	Image              string         `json:"image"`
	Port               int            `json:"port"`
	InternalRegistries []RegistryAuth `json:"internalRegistries"`
	ImagePullerType    string         `json:"imagePullerType"`
	ServiceAccount     string         `json:"serviceAccount"`
}

// ImagePerceiver ...
type ImagePerceiver struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

// PodPerceiver ...
type PodPerceiver struct {
	Name            string `json:"name"`
	Image           string `json:"image"`
	NamespaceFilter string `json:"namespaceFilter,omitempty"`
}

// Perceiver ...
type Perceiver struct {
	EnableImagePerceiver      bool            `json:"enableImagePerceiver"`
	EnablePodPerceiver        bool            `json:"enablePodPerceiver"`
	ImagePerceiver            *ImagePerceiver `json:"imagePerceiver,omitempty"`
	PodPerceiver              *PodPerceiver   `json:"podPerceiver,omitempty"`
	AnnotationIntervalSeconds int             `json:"annotationIntervalSeconds"`
	DumpIntervalMinutes       int             `json:"dumpIntervalMinutes"`
	ServiceAccount            string          `json:"serviceAccount"`
	Port                      int             `json:"port"`
}

// Skyfire ...
type Skyfire struct {
	Name           string `json:"name"`
	Image          string `json:"image"`
	Port           int    `json:"port"`
	PrometheusPort int    `json:"prometheusPort"`
	ServiceAccount string `json:"serviceAccount"`

	HubClientTimeoutSeconds      int `json:"hubClientTimeoutSeconds"`
	HubDumpPauseSeconds          int `json:"hubDumpPauseSeconds"`
	KubeDumpIntervalSeconds      int `json:"kubeDumpIntervalSeconds"`
	PerceptorDumpIntervalSeconds int `json:"perceptorDumpIntervalSeconds"`
}

// Prometheus container definition
type Prometheus struct {
	Name  string `json:"name"`
	Image string `json:"image"`
	Port  int    `json:"port"`
}

// OpsSightSpec is the spec for a OpsSight resource
type OpsSightSpec struct {
	Namespace string `json:"namespace"`
	State     string `json:"state"`

	Perceptor     *Perceptor  `json:"perceptor"`
	ScannerPod    *ScannerPod `json:"scannerPod"`
	Perceiver     *Perceiver  `json:"perceiver"`
	Prometheus    *Prometheus `json:"prometheus"`
	EnableSkyfire bool        `json:"enableSkyfire"`
	Skyfire       *Skyfire    `json:"skyfire"`

	Blackduck *Blackduck `json:"blackduck"`

	EnableMetrics bool `json:"enableMetrics"`

	// CPU and memory configurations
	// Example: "300m"
	DefaultCPU string `json:"defaultCpu,omitempty"`
	// Example: "1300Mi"
	DefaultMem string `json:"defaultMem,omitempty"`

	// Log level
	LogLevel string `json:"logLevel,omitempty"`

	ConfigMapName string `json:"configMapName"`

	// Configuration secret
	SecretName string `json:"secretName"`
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
