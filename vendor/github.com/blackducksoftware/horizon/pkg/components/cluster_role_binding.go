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
	"reflect"

	"github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/util"

	"github.com/koki/short/converter/converters"
	"github.com/koki/short/types"

	"k8s.io/apimachinery/pkg/runtime"
)

// ClusterRoleBinding defines the cluster role binding component
type ClusterRoleBinding struct {
	obj *types.ClusterRoleBinding
}

// NewClusterRoleBinding creates a ClusterRoleBinding object
func NewClusterRoleBinding(config api.ClusterRoleBindingConfig) *ClusterRoleBinding {
	version := "rbac.authorization.k8s.io/v1"
	if len(config.APIVersion) > 0 {
		version = config.APIVersion
	}
	cr := &types.ClusterRoleBinding{
		Version:   version,
		Name:      config.Name,
		Cluster:   config.ClusterName,
		Namespace: config.Namespace,
	}

	return &ClusterRoleBinding{obj: cr}
}

// GetObj will return the cluster role binding object in a format the deployer can use
func (crb *ClusterRoleBinding) GetObj() *types.ClusterRoleBinding {
	return crb.obj
}

// GetName will return the name of cluster role binding
func (crb *ClusterRoleBinding) GetName() string {
	return crb.obj.Name
}

// AddAnnotations adds annotations to the cluster role binding
func (crb *ClusterRoleBinding) AddAnnotations(new map[string]string) {
	crb.obj.Annotations = util.MapMerge(crb.obj.Annotations, new)
}

// RemoveAnnotations removes annotations from the cluster role binding
func (crb *ClusterRoleBinding) RemoveAnnotations(remove []string) {
	for _, k := range remove {
		crb.obj.Annotations = util.RemoveElement(crb.obj.Annotations, k)
	}
}

// AddLabels adds labels to the cluster role binding
func (crb *ClusterRoleBinding) AddLabels(new map[string]string) {
	crb.obj.Labels = util.MapMerge(crb.obj.Labels, new)
}

// RemoveLabels removes labels from the cluster role binding
func (crb *ClusterRoleBinding) RemoveLabels(remove []string) {
	for _, k := range remove {
		crb.obj.Labels = util.RemoveElement(crb.obj.Labels, k)
	}
}

// AddSubject will add a subject to the cluster role binding
func (crb *ClusterRoleBinding) AddSubject(config api.SubjectConfig) {
	s := types.Subject{
		Kind:      config.Kind,
		APIGroup:  config.APIGroup,
		Name:      types.Name(config.Name),
		Namespace: config.Namespace,
	}

	crb.obj.Subjects = append(crb.obj.Subjects, s)
}

// RemoveSubject will remove a subject from the cluster role binding
func (crb *ClusterRoleBinding) RemoveSubject(config api.SubjectConfig) {
	s := types.Subject{
		Kind:      config.Kind,
		APIGroup:  config.APIGroup,
		Name:      types.Name(config.Name),
		Namespace: config.Namespace,
	}

	for l, sub := range crb.obj.Subjects {
		if reflect.DeepEqual(sub, s) {
			crb.obj.Subjects = append(crb.obj.Subjects[:l], crb.obj.Subjects[l+1:]...)
			break
		}
	}
}

// AddRoleRef will add a role reference to the cluster role binding
func (crb *ClusterRoleBinding) AddRoleRef(config api.RoleRefConfig) {
	crb.obj.RoleRef = types.RoleRef{
		APIGroup: config.APIGroup,
		Kind:     config.Kind,
		Name:     types.Name(config.Name),
	}
}

// ToKube returns the kubernetes version of the cluster role binding
func (crb *ClusterRoleBinding) ToKube() (runtime.Object, error) {
	wrapper := &types.ClusterRoleBindingWrapper{ClusterRoleBinding: *crb.obj}
	return converters.Convert_Koki_ClusterRoleBinding_to_Kube(wrapper)
}
