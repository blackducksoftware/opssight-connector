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

// ClusterRole defines the cluster role component
type ClusterRole struct {
	*v1.ClusterRole
	MetadataFuncs
}

// NewClusterRole creates a ClusterRole object
func NewClusterRole(config api.ClusterRoleConfig) *ClusterRole {
	version := "rbac.authorization.k8s.io/v1"
	if len(config.APIVersion) > 0 {
		version = config.APIVersion
	}

	cr := v1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: version,
		},
		ObjectMeta: generateObjectMeta(config.Name, config.Namespace, config.ClusterName),
	}

	return &ClusterRole{&cr, MetadataFuncs{&cr}}
}

// AddPolicyRule will add a policy rule to the cluster role
func (cr *ClusterRole) AddPolicyRule(config api.PolicyRuleConfig) {
	r := v1.PolicyRule{
		Verbs:           config.Verbs,
		APIGroups:       config.APIGroups,
		Resources:       config.Resources,
		ResourceNames:   config.ResourceNames,
		NonResourceURLs: config.NonResourceURLs,
	}

	cr.Rules = append(cr.Rules, r)
}

// RemovePolicyRule will remove a policy rule from the cluster role
func (cr *ClusterRole) RemovePolicyRule(config api.PolicyRuleConfig) {
	pr := v1.PolicyRule{
		Verbs:           config.Verbs,
		APIGroups:       config.APIGroups,
		Resources:       config.Resources,
		ResourceNames:   config.ResourceNames,
		NonResourceURLs: config.NonResourceURLs,
	}

	for l, r := range cr.Rules {
		if reflect.DeepEqual(r, pr) {
			cr.Rules = append(cr.Rules[:l], cr.Rules[l+1:]...)
			break
		}
	}
}

// AddAggregationRule will add an aggregation rule to the cluster role
func (cr *ClusterRole) AddAggregationRule(rule api.SelectorConfig) {
	selector := createLabelSelector(rule)

	if cr.AggregationRule == nil {
		cr.AggregationRule = &v1.AggregationRule{
			ClusterRoleSelectors: []metav1.LabelSelector{selector},
		}
	} else {
		cr.AggregationRule.ClusterRoleSelectors = append(cr.AggregationRule.ClusterRoleSelectors, selector)
	}
}

// RemoveAggregationRule will remove an aggregation rule from the cluster role
func (cr *ClusterRole) RemoveAggregationRule(rule api.SelectorConfig) {
	if cr.AggregationRule != nil {
		selector := createLabelSelector(rule)
		for l, r := range cr.AggregationRule.ClusterRoleSelectors {
			if reflect.DeepEqual(r, selector) {
				cr.AggregationRule.ClusterRoleSelectors = append(cr.AggregationRule.ClusterRoleSelectors[:l], cr.AggregationRule.ClusterRoleSelectors[l+1:]...)
				break
			}
		}
	}
}

// Deploy will deploy the cluster role to the cluster
func (cr *ClusterRole) Deploy(res api.DeployerResources) error {
	_, err := res.KubeClient.RbacV1().ClusterRoles().Create(cr.ClusterRole)
	return err
}

// Undeploy will remove the cluster role from the cluster
func (cr *ClusterRole) Undeploy(res api.DeployerResources) error {
	return res.KubeClient.RbacV1().ClusterRoles().Delete(cr.Name, &metav1.DeleteOptions{})
}
