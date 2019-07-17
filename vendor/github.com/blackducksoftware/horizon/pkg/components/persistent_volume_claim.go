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
	"github.com/blackducksoftware/horizon/pkg/util"
	"github.com/koki/short/converter/converters"
	"github.com/koki/short/types"

	"k8s.io/apimachinery/pkg/api/resource"
)

// PersistentVolumeClaim defines the persistent volume claim component
type PersistentVolumeClaim struct {
	obj *types.PersistentVolumeClaim
}

// NewPersistentVolumeClaim creates a new PersistentVolumeClaim object
func NewPersistentVolumeClaim(config api.PVCConfig) (*PersistentVolumeClaim, error) {
	_, err := resource.ParseQuantity(config.Size)
	if err != nil {
		return nil, fmt.Errorf("invalid size: %v", err)
	}
	pvc := &types.PersistentVolumeClaim{
		Version:      config.APIVersion,
		Cluster:      config.ClusterName,
		Name:         config.Name,
		Namespace:    config.Namespace,
		StorageClass: config.Class,
		Volume:       config.VolumeName,
		Storage:      config.Size,
	}

	return &PersistentVolumeClaim{obj: pvc}, nil
}

// GetObj returns the PersistentVolumeClaim object in a format the deployer can use
func (p *PersistentVolumeClaim) GetObj() *types.PersistentVolumeClaim {
	return p.obj
}

// GetName returns the name of the PersistentVolumeClaim
func (p *PersistentVolumeClaim) GetName() string {
	return p.obj.Name
}

// AddAnnotations adds annotations to the PersistentVolumeClaim
func (p *PersistentVolumeClaim) AddAnnotations(new map[string]string) {
	p.obj.Annotations = util.MapMerge(p.obj.Annotations, new)
}

// RemoveAnnotations removes annotations from the PersistentVolumeClaim
func (p *PersistentVolumeClaim) RemoveAnnotations(remove []string) {
	for _, k := range remove {
		p.obj.Annotations = util.RemoveElement(p.obj.Annotations, k)
	}
}

// AddLabels adds labels to the PersistentVolumeClaim
func (p *PersistentVolumeClaim) AddLabels(new map[string]string) {
	p.obj.Labels = util.MapMerge(p.obj.Labels, new)
}

// RemoveLabels removes labels from the PersistentVolumeClaim
func (p *PersistentVolumeClaim) RemoveLabels(remove []string) {
	for _, k := range remove {
		p.obj.Labels = util.RemoveElement(p.obj.Labels, k)
	}
}

// AddMatchLabelsSelectors adds the given match label selectors to the PersistentVolumeClaim
func (p *PersistentVolumeClaim) AddMatchLabelsSelectors(new map[string]string) {
	if p.obj.Selector == nil {
		p.obj.Selector = &types.RSSelector{}
	}
	p.obj.Selector.Labels = util.MapMerge(p.obj.Selector.Labels, new)
}

// RemoveMatchLabelsSelectors removes the given match label selectors from the PersistentVolumeClaim
func (p *PersistentVolumeClaim) RemoveMatchLabelsSelectors(remove []string) {
	for _, k := range remove {
		p.obj.Selector.Labels = util.RemoveElement(p.obj.Selector.Labels, k)
	}
}

// AddMatchExpressionsSelector will add match expressions selectors to the PersistentVolumeClaim.
// It takes a string in the following form:
// key <op> <value>
// Where op can be:
// = 	Equal to value ot should be one of the comma separated values
// !=	Key should not be one of the comma separated values
// If no op is provided, then the key should (or should not) exist
// <key>	key should exist
// !<key>	key should not exist
func (p *PersistentVolumeClaim) AddMatchExpressionsSelector(add string) {
	p.obj.Selector.Shorthand = add
}

// RemoveMatchExpressionsSelector removes the match expressions selector from the PersistentVolumeClaim
func (p *PersistentVolumeClaim) RemoveMatchExpressionsSelector() {
	p.obj.Selector.Shorthand = ""
}

// AddAccessMode will add an access mode to the persistent volume claim if the mode
// doesn't already exist
func (p *PersistentVolumeClaim) AddAccessMode(mode api.PVCAccessModeType) {
	shortMode := p.convertType(mode)
	for _, m := range p.obj.AccessModes {
		if m == shortMode {
			return
		}
	}

	p.obj.AccessModes = append(p.obj.AccessModes, shortMode)
}

// RemoveAccessMode will remove an access mode from the persistent volume claim
func (p *PersistentVolumeClaim) RemoveAccessMode(mode api.PVCAccessModeType) {
	shortMode := p.convertType(mode)
	for l, m := range p.obj.AccessModes {
		if m == shortMode {
			p.obj.AccessModes = append(p.obj.AccessModes[:l], p.obj.AccessModes[l+1:]...)
			return
		}
	}
}

func (p *PersistentVolumeClaim) convertType(mode api.PVCAccessModeType) types.PersistentVolumeAccessMode {
	var m types.PersistentVolumeAccessMode

	switch mode {
	case api.ReadWriteOnce:
		m = types.ReadWriteOnce
	case api.ReadOnlyMany:
		m = types.ReadOnlyMany
	case api.ReadWriteMany:
		m = types.ReadWriteMany
	}

	return m
}

// ToKube returns the kubernetes version of the persistent volume claim
func (p *PersistentVolumeClaim) ToKube() (interface{}, error) {
	wrapper := &types.PersistentVolumeClaimWrapper{PersistentVolumeClaim: *p.obj}
	return converters.Convert_Koki_PVC_to_Kube_PVC(wrapper)
}
