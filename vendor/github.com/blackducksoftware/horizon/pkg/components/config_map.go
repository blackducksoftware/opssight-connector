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
	"github.com/imdario/mergo"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigMap defines the config map component
type ConfigMap struct {
	*v1.ConfigMap
	MetadataFuncs
}

// NewConfigMap creates a ConfigMap object
func NewConfigMap(config api.ConfigMapConfig) *ConfigMap {
	version := "v1"
	if len(config.APIVersion) > 0 {
		version = config.APIVersion
	}

	c := v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: version,
		},
		ObjectMeta: generateObjectMeta(config.Name, config.Namespace, config.ClusterName),
	}

	return &ConfigMap{&c, MetadataFuncs{&c}}
}

// AddData adds key value pairs to the config map
func (c *ConfigMap) AddData(new map[string]string) {
	_ = mergo.Merge(&c.Data, new, mergo.WithOverride)
}

// RemoveData removes the provided keys from the config map
func (c *ConfigMap) RemoveData(remove []string) {
	for _, k := range remove {
		delete(c.Data, k)
	}
}

// AddBinaryData adds key value pairs to the config map
func (c *ConfigMap) AddBinaryData(new map[string][]byte) {
	_ = mergo.Merge(&c.BinaryData, new, mergo.WithOverride)
}

// RemoveBinaryData removes the provided keys from the config map
func (c *ConfigMap) RemoveBinaryData(remove map[string][]byte) {
	for k := range remove {
		if _, exists := c.BinaryData[k]; exists {
			delete(c.BinaryData, k)
		}
	}
}

// Deploy will deploy the config map to the cluster
func (c *ConfigMap) Deploy(res api.DeployerResources) error {
	_, err := res.KubeClient.CoreV1().ConfigMaps(c.Namespace).Create(c.ConfigMap)
	return err
}

// Undeploy will remove the config map from the cluster
func (c *ConfigMap) Undeploy(res api.DeployerResources) error {
	return res.KubeClient.CoreV1().ConfigMaps(c.Namespace).Delete(c.Name, &metav1.DeleteOptions{})
}
