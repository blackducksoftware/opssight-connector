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

	"k8s.io/api/rbac/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterRoleBinding defines the cluster role binding component
type ClusterRoleBinding struct {
	*v1.ClusterRoleBinding
	MetadataFuncs
}

// NewClusterRoleBinding creates a ClusterRoleBinding object
func NewClusterRoleBinding(config api.ClusterRoleBindingConfig) *ClusterRoleBinding {
	version := "rbac.authorization.k8s.io/v1"
	if len(config.APIVersion) > 0 {
		version = config.APIVersion
	}

	crb := v1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: version,
		},
		ObjectMeta: generateObjectMeta(config.Name, config.Namespace, config.ClusterName),
	}

	return &ClusterRoleBinding{&crb, MetadataFuncs{&crb}}
}

// AddSubject will add a subject to the cluster role binding
func (crb *ClusterRoleBinding) AddSubject(config api.SubjectConfig) {
	s := v1.Subject{
		Kind:      config.Kind,
		APIGroup:  config.APIGroup,
		Name:      config.Name,
		Namespace: config.Namespace,
	}

	crb.Subjects = append(crb.Subjects, s)
}

// RemoveSubject will remove a subject from the cluster role binding
func (crb *ClusterRoleBinding) RemoveSubject(config api.SubjectConfig) {
	s := v1.Subject{
		Kind:      config.Kind,
		APIGroup:  config.APIGroup,
		Name:      config.Name,
		Namespace: config.Namespace,
	}

	for l, sub := range crb.Subjects {
		if reflect.DeepEqual(sub, s) {
			crb.Subjects = append(crb.Subjects[:l], crb.Subjects[l+1:]...)
			break
		}
	}
}

// AddRoleRef will add a role reference to the cluster role binding
func (crb *ClusterRoleBinding) AddRoleRef(config api.RoleRefConfig) {
	crb.RoleRef = v1.RoleRef{
		APIGroup: config.APIGroup,
		Kind:     config.Kind,
		Name:     config.Name,
	}
}

// Deploy will deploy the cluster role binding to the cluster
func (crb *ClusterRoleBinding) Deploy(res api.DeployerResources) error {
	_, err := res.KubeClient.RbacV1().ClusterRoleBindings().Create(crb.ClusterRoleBinding)
	return err
}

// Undeploy will remove the cluster role binding from the cluster
func (crb *ClusterRoleBinding) Undeploy(res api.DeployerResources) error {
	return res.KubeClient.RbacV1().ClusterRoleBindings().Delete(crb.Name, &metav1.DeleteOptions{})
}
