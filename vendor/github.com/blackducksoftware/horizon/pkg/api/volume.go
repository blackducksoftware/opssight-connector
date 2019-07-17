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

// EmptyDirVolumeConfig defines the configuration for an
// EmptyDir volume
type EmptyDirVolumeConfig struct {
	VolumeName string
	Medium     StorageMediumType
	SizeLimit  string
}

// StorageMediumType defines the storage medium
type StorageMediumType int

const (
	StorageMediumDefault StorageMediumType = iota
	StorageMediumMemory
	StorageMediumHugePages
)

// ConfigMapOrSecretVolumeConfig defines the configuration for
// a config map or secret
type ConfigMapOrSecretVolumeConfig struct {
	VolumeName      string
	MapOrSecretName string
	Items           map[string]KeyAndMode
	DefaultMode     *int32
	Required        *bool
}

// KeyAndMode defines the key and file mode
type KeyAndMode struct {
	KeyOrPath string
	Mode      *int32
}

// HostPathVolumeConfig defines the configuration for a
// HostPath volume
type HostPathVolumeConfig struct {
	VolumeName string
	Path       string
	Type       HostPathType
}

// HostPathType defines the type of HostPath
type HostPathType int

const (
	HostPathUnset HostPathType = iota
	HostPathDirectoryOrCreate
	HostPathDirectory
	HostPathFileOrCreate
	HostPathFile
	HostPathSocket
	HostPathCharDev
	HostPathBlockDev
)

// GCEPersistentDiskVolumeConfig defines the configuraton for a
// GCE Persistent Disk volume
type GCEPersistentDiskVolumeConfig struct {
	VolumeName string
	DiskName   string
	FSType     string
	Partition  int32
	ReadOnly   bool
}

// PVCVolumeConfig defines the configuration for a persistent volume claim
type PVCVolumeConfig struct {
	VolumeName string
	PVCName    string
	ReadOnly   bool
}
