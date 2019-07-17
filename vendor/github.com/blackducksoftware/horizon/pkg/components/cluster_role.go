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
	"strings"

	"github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/util"

	"github.com/koki/short/converter/converters"
	"github.com/koki/short/types"

	"k8s.io/apimachinery/pkg/runtime"
)

// ClusterRole defines the cluster role component
type ClusterRole struct {
	obj *types.ClusterRole
}

// NewClusterRole creates a ClusterRole object
func NewClusterRole(config api.ClusterRoleConfig) *ClusterRole {
	version := "rbac.authorization.k8s.io/v1"
	if len(config.APIVersion) > 0 {
		version = config.APIVersion
	}
	cr := &types.ClusterRole{
		Version:   version,
		Name:      config.Name,
		Cluster:   config.ClusterName,
		Namespace: config.Namespace,
	}

	return &ClusterRole{obj: cr}
}

// GetObj will return the cluster role object in a format the deployer can use
func (cr *ClusterRole) GetObj() *types.ClusterRole {
	return cr.obj
}

// GetName will return the name of cluster role
func (cr *ClusterRole) GetName() string {
	return cr.obj.Name
}

// AddAnnotations adds annotations to the cluster role
func (cr *ClusterRole) AddAnnotations(new map[string]string) {
	cr.obj.Annotations = util.MapMerge(cr.obj.Annotations, new)
}

// RemoveAnnotations removes annotations from the cluster role
func (cr *ClusterRole) RemoveAnnotations(remove []string) {
	for _, k := range remove {
		cr.obj.Annotations = util.RemoveElement(cr.obj.Annotations, k)
	}
}

// AddLabels adds labels to the cluster role
func (cr *ClusterRole) AddLabels(new map[string]string) {
	cr.obj.Labels = util.MapMerge(cr.obj.Labels, new)
}

// RemoveLabels removes labels from the cluster role
func (cr *ClusterRole) RemoveLabels(remove []string) {
	for _, k := range remove {
		cr.obj.Labels = util.RemoveElement(cr.obj.Labels, k)
	}
}

// AddPolicyRule will add a policy rule to the cluster role
func (cr *ClusterRole) AddPolicyRule(config api.PolicyRuleConfig) {
	r := types.PolicyRule{
		Verbs:           config.Verbs,
		APIGroups:       config.APIGroups,
		Resources:       config.Resources,
		ResourceNames:   config.ResourceNames,
		NonResourceURLs: config.NonResourceURLs,
	}

	cr.obj.Rules = append(cr.obj.Rules, r)
}

// RemovePolicyRule will remove a policy rule from the cluster role
func (cr *ClusterRole) RemovePolicyRule(config api.PolicyRuleConfig) {
	pr := types.PolicyRule{
		Verbs:           config.Verbs,
		APIGroups:       config.APIGroups,
		Resources:       config.Resources,
		ResourceNames:   config.ResourceNames,
		NonResourceURLs: config.NonResourceURLs,
	}
	for l, r := range cr.obj.Rules {
		if reflect.DeepEqual(r, pr) {
			cr.obj.Rules = append(cr.obj.Rules[:l], cr.obj.Rules[l+1:]...)
			break
		}
	}
}

// AddAggregationRule will add an aggregation rule to the cluster role
func (cr *ClusterRole) AddAggregationRule(rule string) {
	cr.obj.AggregationRule = append(cr.obj.AggregationRule, rule)
}

// RemoveAggregationRule will remove an aggregation rule from the cluster role
func (cr *ClusterRole) RemoveAggregationRule(rule string) {
	for l, r := range cr.obj.AggregationRule {
		if strings.Compare(r, rule) == 0 {
			cr.obj.AggregationRule = append(cr.obj.AggregationRule[:l], cr.obj.AggregationRule[l+1:]...)
			break
		}
	}
}

// ToKube returns the kubernetes version of the cluster role
func (cr *ClusterRole) ToKube() (runtime.Object, error) {
	wrapper := &types.ClusterRoleWrapper{ClusterRole: *cr.obj}
	return converters.Convert_Koki_ClusterRole_to_Kube(wrapper)
}
