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

package components

import (
	"fmt"

	"github.com/blackducksoftware/horizon/pkg/api"

	"k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/resource"
)

// Volume defines the volume component
type Volume struct {
	*v1.Volume
}

// NewEmptyDirVolume creates an EmptyDir volume object
func NewEmptyDirVolume(config api.EmptyDirVolumeConfig) (*Volume, error) {
	var size *resource.Quantity

	if len(config.SizeLimit) > 0 {
		s, err := resource.ParseQuantity(config.SizeLimit)
		if err != nil {
			return nil, fmt.Errorf("invalid size: %v", err)
		}
		size = &s
	}

	v := v1.Volume{
		Name: config.VolumeName,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{
				SizeLimit: size,
			},
		},
	}

	switch config.Medium {
	case api.StorageMediumDefault:
		v.VolumeSource.EmptyDir.Medium = v1.StorageMediumDefault
	case api.StorageMediumMemory:
		v.VolumeSource.EmptyDir.Medium = v1.StorageMediumMemory
	case api.StorageMediumHugePages:
		v.VolumeSource.EmptyDir.Medium = v1.StorageMediumHugePages
	}

	return &Volume{&v}, nil
}

// NewHostPathVolume create a HostPath volume object
func NewHostPathVolume(config api.HostPathVolumeConfig) *Volume {
	v := v1.Volume{
		Name: config.VolumeName,
		VolumeSource: v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: config.Path,
			},
		},
	}

	var hpType v1.HostPathType
	switch config.Type {
	case api.HostPathUnset:
		hpType = v1.HostPathUnset
	case api.HostPathDirectoryOrCreate:
		hpType = v1.HostPathDirectoryOrCreate
	case api.HostPathDirectory:
		hpType = v1.HostPathDirectory
	case api.HostPathFileOrCreate:
		hpType = v1.HostPathFileOrCreate
	case api.HostPathFile:
		hpType = v1.HostPathFile
	case api.HostPathSocket:
		hpType = v1.HostPathSocket
	case api.HostPathCharDev:
		hpType = v1.HostPathCharDev
	case api.HostPathBlockDev:
		hpType = v1.HostPathBlockDev
	default:
		hpType = v1.HostPathUnset
	}

	v.VolumeSource.HostPath.Type = &hpType

	return &Volume{&v}
}

// NewConfigMapVolume creates a ConfigMap volume object
func NewConfigMapVolume(config api.ConfigMapOrSecretVolumeConfig) *Volume {
	items := generateKeyToPath(config.Items)

	v := v1.Volume{
		Name: config.VolumeName,
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: config.MapOrSecretName,
				},
				Items:       items,
				DefaultMode: config.DefaultMode,
				Optional:    config.Optional,
			},
		},
	}

	return &Volume{&v}
}

// NewSecretVolume creates a Secret volume object
func NewSecretVolume(config api.ConfigMapOrSecretVolumeConfig) *Volume {
	items := generateKeyToPath(config.Items)

	v := v1.Volume{
		Name: config.VolumeName,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName:  config.MapOrSecretName,
				Items:       items,
				DefaultMode: config.DefaultMode,
				Optional:    config.Optional,
			},
		},
	}

	return &Volume{&v}
}

func generateKeyToPath(items []api.KeyPath) []v1.KeyToPath {
	converted := []v1.KeyToPath{}
	for _, kp := range items {
		ktp := v1.KeyToPath{
			Key:  kp.Key,
			Path: kp.Path,
			Mode: kp.Mode,
		}
		converted = append(converted, ktp)
	}

	return converted
}

// NewGCEPersistentDiskVolume creates a new GCE Persistent Disk volume object
func NewGCEPersistentDiskVolume(config api.GCEPersistentDiskVolumeConfig) *Volume {
	v := v1.Volume{
		Name: config.VolumeName,
		VolumeSource: v1.VolumeSource{
			GCEPersistentDisk: &v1.GCEPersistentDiskVolumeSource{
				PDName:    config.DiskName,
				FSType:    config.FSType,
				Partition: config.Partition,
				ReadOnly:  config.ReadOnly,
			},
		},
	}

	return &Volume{&v}
}

// NewPVCVolume creates a new Persistent Volume Claim volume object
func NewPVCVolume(config api.PVCVolumeConfig) *Volume {
	v := v1.Volume{
		Name: config.VolumeName,
		VolumeSource: v1.VolumeSource{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
				ClaimName: config.PVCName,
				ReadOnly:  config.ReadOnly,
			},
		},
	}

	return &Volume{&v}
}
