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

package scannertester

import (
	"k8s.io/api/core/v1"
	// v1beta1 "k8s.io/api/extensions/v1beta1"
)

type Openshift struct {
	Config *Config

	Kube *Kube

	ReplicationControllers []*v1.ReplicationController
	ConfigMaps             []*v1.ConfigMap
	Services               []*v1.Service
	Secrets                []*v1.Secret
}

func NewOpenshift(config *Config) *Openshift {
	os := &Openshift{
		Config: config,
		Kube:   NewKube(config)}
	return os
}

func (os *Openshift) GetConfigMaps() []*v1.ConfigMap {
	return os.Kube.GetConfigMaps()
}

func (os *Openshift) GetServices() []*v1.Service {
	return os.Kube.GetServices()
}

func (os *Openshift) GetSecrets() []*v1.Secret {
	return os.Kube.GetSecrets()
}

func (os *Openshift) GetReplicationControllers() []*v1.ReplicationController {
	return os.Kube.GetReplicationControllers()
}
