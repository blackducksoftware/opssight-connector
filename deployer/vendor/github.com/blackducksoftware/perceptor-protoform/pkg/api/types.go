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

package api

import (
	"github.com/koki/short/types"
)

// ReplicationControllerConfig defines the configuration for a
// replication controller
type ReplicationControllerConfig struct {
	Name           string
	Replicas       int32
	Selector       map[string]string
	Labels         map[string]string
	Vols           map[string]types.Volume
	Containers     []types.Container
	ServiceAccount string
}

// ServiceConfig defines the configuration for a service
type ServiceConfig struct {
	Name     string
	Ports    map[string]int32
	Selector map[string]string
}

// ConfigMapConfig defines the configuration for a config map
type ConfigMapConfig struct {
	Name      string
	Namespace string
	Data      map[string]string
}

// ProtoformDefaults defines default values for Protoform.
// These fields need to be named the same as those in
// protoformConfig in order for defaults to be applied
// properly.  A field that exists in ProtoformDefaults
// but does not exist in protoformConfig will be ignored
type ProtoformDefaults struct {
	PerceptorPort                    int
	PerceiverPort                    int
	ScannerPort                      int
	ImageFacadePort                  int
	AnnotationIntervalSeconds        int
	DumpIntervalMinutes              int
	HubClientTimeoutPerceptorSeconds int
	HubClientTimeoutScannerSeconds   int
	HubHost                          string
	HubUser                          string
	HubUserPassword                  string
	HubPort                          int
	DockerUsername                   string
	DockerPasswordOrToken            string
	ConcurrentScanLimit              int
	InternalDockerRegistries         []string
	DefaultVersion                   string
	Registry                         string
	ImagePath                        string
	PerceptorImageName               string
	ScannerImageName                 string
	ImagePerceiverImageName          string
	PodPerceiverImageName            string
	ImageFacadeImageName             string
	PerceptorContainerVersion        string
	ScannerContainerVersion          string
	PerceiverContainerVersion        string
	ImageFacadeContainerVersion      string
	LogLevel                         string
	Namespace                        string
	DefaultCPU                       string // Should be passed like: e.g. "300m"
	DefaultMem                       string // Should be passed like: e.g "1300Mi"
}
