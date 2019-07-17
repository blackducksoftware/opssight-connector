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
	"github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/util"

	"github.com/koki/short/converter/converters"
	"github.com/koki/short/types"

	"k8s.io/apimachinery/pkg/runtime"
)

// CustomResourceDefinition defines a custom resource
type CustomResourceDefinition struct {
	obj *types.CustomResourceDefinition
}

// NewCustomResourceDefintion returns a CustomerResrouceDefinition object
func NewCustomResourceDefintion(config api.CRDConfig) *CustomResourceDefinition {
	name := types.CRDName{
		Plural:     config.Plural,
		Singular:   config.Singular,
		ShortNames: config.ShortNames,
		Kind:       config.Kind,
		ListKind:   config.ListKind,
	}
	meta := types.CRDMeta{
		Group:   config.Group,
		Version: config.CRDVersion,
		CRDName: name,
	}
	crd := &types.CustomResourceDefinition{
		Version:    config.APIVersion,
		Cluster:    config.ClusterName,
		Name:       config.Name,
		Namespace:  config.Namespace,
		Validation: config.Validation,
		CRDMeta:    meta,
	}

	switch config.Scope {
	case api.CRDClusterScoped:
		crd.Scope = types.CRDClusterScoped
	case api.CRDNamespaceScoped:
		crd.Scope = types.CRDNamespaceScoped
	}

	return &CustomResourceDefinition{obj: crd}
}

// GetObj returns the custom resource definition object in a format the deployer can use
func (crd *CustomResourceDefinition) GetObj() *types.CustomResourceDefinition {
	return crd.obj
}

// GetName returns the name of the custom resource definition
func (crd *CustomResourceDefinition) GetName() string {
	return crd.obj.Name
}

// AddAnnotations adds annotations to the custom resource definition
func (crd *CustomResourceDefinition) AddAnnotations(new map[string]string) {
	crd.obj.Annotations = util.MapMerge(crd.obj.Annotations, new)
}

// RemoveAnnotations removes annotations from the custom resource definition
func (crd *CustomResourceDefinition) RemoveAnnotations(remove []string) {
	for _, k := range remove {
		crd.obj.Annotations = util.RemoveElement(crd.obj.Annotations, k)
	}
}

// AddLabels adds labels to the custom resource definition
func (crd *CustomResourceDefinition) AddLabels(new map[string]string) {
	crd.obj.Labels = util.MapMerge(crd.obj.Labels, new)
}

// RemoveLabels removes labels from the custom resource definition
func (crd *CustomResourceDefinition) RemoveLabels(remove []string) {
	for _, k := range remove {
		crd.obj.Labels = util.RemoveElement(crd.obj.Labels, k)
	}
}

// ToKube returns the kubernetes version of the custom resource definition
func (crd *CustomResourceDefinition) ToKube() (runtime.Object, error) {
	wrapper := &types.CRDWrapper{CRD: *crd.obj}
	return converters.Convert_Koki_CRD_to_Kube(wrapper)
}
