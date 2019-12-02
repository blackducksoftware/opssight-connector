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

// PVCConfig defines the configuration for a persistent volume claim
type PVCConfig struct {
	APIVersion         string
	ClusterName        string
	Name               string
	Namespace          string
	Class              *string
	VolumeName         string
	Size               string
	Mode               PVCMode
	DataSourceAPIGroup *string
	DataSourceKind     string
	DataSourceName     string
}

// PVCAccessModeType defines the access mode for the persistent volume claim
type PVCAccessModeType int

const (
	ReadWriteOnce PVCAccessModeType = iota + 1
	ReadOnlyMany
	ReadWriteMany
)

type PVCMode int

const (
	PVCModeBLock PVCMode = iota + 1
	PVCModeFilesystem
)
