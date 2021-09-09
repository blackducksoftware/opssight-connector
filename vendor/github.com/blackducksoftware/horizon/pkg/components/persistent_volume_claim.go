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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PersistentVolumeClaim defines the persistent volume claim component
type PersistentVolumeClaim struct {
	*v1.PersistentVolumeClaim
	MetadataFuncs
	LabelSelectorFuncs
}

// NewPersistentVolumeClaim creates a new PersistentVolumeClaim object
func NewPersistentVolumeClaim(config api.PVCConfig) (*PersistentVolumeClaim, error) {
	version := "v1"
	if len(config.APIVersion) > 0 {
		version = config.APIVersion
	}

	size, err := resource.ParseQuantity(config.Size)
	if err != nil {
		return nil, fmt.Errorf("invalid size: %v", err)
	}

	pvc := v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: version,
		},
		ObjectMeta: generateObjectMeta(config.Name, config.Namespace, config.ClusterName),
		Spec: v1.PersistentVolumeClaimSpec{
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: size,
				},
			},
			VolumeName:       config.VolumeName,
			StorageClassName: config.Class,
		},
	}

	var volumeMode v1.PersistentVolumeMode
	switch config.Mode {
	case api.PVCModeBLock:
		volumeMode = v1.PersistentVolumeBlock
		pvc.Spec.VolumeMode = &volumeMode
	case api.PVCModeFilesystem:
		volumeMode = v1.PersistentVolumeFilesystem
		pvc.Spec.VolumeMode = &volumeMode
	}

	if config.DataSourceAPIGroup != nil || len(config.DataSourceKind) != 0 || len(config.DataSourceName) != 0 {
		pvc.Spec.DataSource = &v1.TypedLocalObjectReference{
			APIGroup: config.DataSourceAPIGroup,
			Kind:     config.DataSourceKind,
			Name:     config.DataSourceName,
		}
	}

	return &PersistentVolumeClaim{&pvc, MetadataFuncs{&pvc}, LabelSelectorFuncs{&pvc}}, nil
}

// AddAccessMode will add an access mode to the persistent volume claim if the mode
// doesn't already exist
func (p *PersistentVolumeClaim) AddAccessMode(mode api.PVCAccessModeType) {
	newMode := p.convertType(mode)
	for _, m := range p.Spec.AccessModes {
		if m == newMode {
			return
		}
	}

	p.Spec.AccessModes = append(p.Spec.AccessModes, newMode)
}

// RemoveAccessMode will remove an access mode from the persistent volume claim
func (p *PersistentVolumeClaim) RemoveAccessMode(mode api.PVCAccessModeType) {
	newMode := p.convertType(mode)
	for l, m := range p.Spec.AccessModes {
		if m == newMode {
			p.Spec.AccessModes = append(p.Spec.AccessModes[:l], p.Spec.AccessModes[l+1:]...)
			return
		}
	}
}

func (p *PersistentVolumeClaim) convertType(mode api.PVCAccessModeType) v1.PersistentVolumeAccessMode {
	var m v1.PersistentVolumeAccessMode

	switch mode {
	case api.ReadWriteOnce:
		m = v1.ReadWriteOnce
	case api.ReadOnlyMany:
		m = v1.ReadOnlyMany
	case api.ReadWriteMany:
		m = v1.ReadWriteMany
	}

	return m
}

// Deploy will deploy the persistent volume claim to the cluster
func (p *PersistentVolumeClaim) Deploy(res api.DeployerResources) error {
	_, err := res.KubeClient.CoreV1().PersistentVolumeClaims(p.Namespace).Create(p.PersistentVolumeClaim)
	return err
}

// Undeploy will remove the persistent volume claim from the cluster
func (p *PersistentVolumeClaim) Undeploy(res api.DeployerResources) error {
	return res.KubeClient.CoreV1().PersistentVolumeClaims(p.Namespace).Delete(p.Name, &metav1.DeleteOptions{})
}
