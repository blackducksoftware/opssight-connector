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

// ConfigMap defines the config map component
type ConfigMap struct {
	obj *types.ConfigMap
}

// NewConfigMap creates a ConfigMap object
func NewConfigMap(config api.ConfigMapConfig) *ConfigMap {
	c := &types.ConfigMap{
		Version:   config.APIVersion,
		Cluster:   config.ClusterName,
		Name:      config.Name,
		Namespace: config.Namespace,
	}

	return &ConfigMap{obj: c}
}

// GetObj returns the config map object in a format the deployer
// can use
func (c *ConfigMap) GetObj() *types.ConfigMap {
	return c.obj
}

// GetName returns the name of the config map
func (c *ConfigMap) GetName() string {
	return c.obj.Name
}

// AddAnnotations adds annotations to the config map
func (c *ConfigMap) AddAnnotations(new map[string]string) {
	c.obj.Annotations = util.MapMerge(c.obj.Annotations, new)
}

// RemoveAnnotations removes annotations from the config map
func (c *ConfigMap) RemoveAnnotations(remove []string) {
	for _, k := range remove {
		c.obj.Annotations = util.RemoveElement(c.obj.Annotations, k)
	}
}

// AddLabels adds labels to the config map
func (c *ConfigMap) AddLabels(new map[string]string) {
	c.obj.Labels = util.MapMerge(c.obj.Labels, new)
}

// RemoveLabels removes labels from the config map
func (c *ConfigMap) RemoveLabels(remove []string) {
	for _, k := range remove {
		c.obj.Labels = util.RemoveElement(c.obj.Labels, k)
	}
}

// AddData adds key value pairs to the config map
func (c *ConfigMap) AddData(new map[string]string) {
	c.obj.Data = util.MapMerge(c.obj.Data, new)
}

// RemoveData removes the provided keys from the config map
func (c *ConfigMap) RemoveData(remove []string) {
	for _, k := range remove {
		c.obj.Data = util.RemoveElement(c.obj.Data, k)
	}
}

// ToKube returns the kubernetes version of the config map
func (c *ConfigMap) ToKube() (runtime.Object, error) {
	wrapper := &types.ConfigMapWrapper{ConfigMap: *c.obj}
	return converters.Convert_Koki_ConfigMap_to_Kube_v1_ConfigMap(wrapper)
}
