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

// Role defines the cluster role component
type Role struct {
	*v1.Role
	MetadataFuncs
}

// NewRole creates a Role object
func NewRole(config api.RoleConfig) *Role {
	version := "rbac.authorization.k8s.io/v1"
	if len(config.APIVersion) > 0 {
		version = config.APIVersion
	}

	cr := v1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        config.Name,
			Namespace:   config.Namespace,
			ClusterName: config.ClusterName,
		},
	}

	return &Role{&cr, MetadataFuncs{&cr}}
}

// AddPolicyRule will add a policy rule to the cluster role
func (r *Role) AddPolicyRule(config api.PolicyRuleConfig) {
	pr := v1.PolicyRule{
		Verbs:           config.Verbs,
		APIGroups:       config.APIGroups,
		Resources:       config.Resources,
		ResourceNames:   config.ResourceNames,
		NonResourceURLs: config.NonResourceURLs,
	}

	r.Rules = append(r.Rules, pr)
}

// RemovePolicyRule will remove a policy rule from the cluster role
func (r *Role) RemovePolicyRule(config api.PolicyRuleConfig) {
	pr := v1.PolicyRule{
		Verbs:           config.Verbs,
		APIGroups:       config.APIGroups,
		Resources:       config.Resources,
		ResourceNames:   config.ResourceNames,
		NonResourceURLs: config.NonResourceURLs,
	}

	for l, rule := range r.Rules {
		if reflect.DeepEqual(rule, pr) {
			r.Rules = append(r.Rules[:l], r.Rules[l+1:]...)
			break
		}
	}
}

// Deploy will deploy the cluster role to the cluster
func (r *Role) Deploy(res api.DeployerResources) error {
	_, err := res.KubeClient.RbacV1().Roles(r.Namespace).Create(r.Role)
	return err
}

// Undeploy will remove the cluster role from the cluster
func (r *Role) Undeploy(res api.DeployerResources) error {
	return res.KubeClient.RbacV1().Roles(r.Namespace).Delete(r.Name, &metav1.DeleteOptions{})
}
