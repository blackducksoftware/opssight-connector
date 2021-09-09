/*
Copyright (C) 2019 Synopsys, Inc.

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

package crdupdater

import (
	"reflect"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

// ConfigMap stores the configuration to add or delete the config map object
type ConfigMap struct {
	config        *CommonConfig
	deployer      *util.DeployerHelper
	configMaps    []*components.ConfigMap
	oldConfigMaps map[string]corev1.ConfigMap
	newConfigMaps map[string]*corev1.ConfigMap
}

// NewConfigMap returns the config map
func NewConfigMap(config *CommonConfig, configMaps []*components.ConfigMap) (*ConfigMap, error) {
	deployer, err := util.NewDeployer(config.kubeConfig)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to get deployer object for %s", config.namespace)
	}
	newConfigMaps := append([]*components.ConfigMap{}, configMaps...)
	for i := 0; i < len(newConfigMaps); i++ {
		if !isLabelsExist(config.expectedLabels, newConfigMaps[i].Labels) {
			newConfigMaps = append(newConfigMaps[:i], newConfigMaps[i+1:]...)
			i--
		}
	}
	return &ConfigMap{
		config:        config,
		deployer:      deployer,
		configMaps:    newConfigMaps,
		oldConfigMaps: make(map[string]corev1.ConfigMap, 0),
		newConfigMaps: make(map[string]*corev1.ConfigMap, 0),
	}, nil
}

// buildNewAndOldObject builds the old and new config map
func (c *ConfigMap) buildNewAndOldObject() error {
	// build old config map
	oldConfigMaps, err := c.list()
	if err != nil {
		return errors.Annotatef(err, "unable to get config maps for %s", c.config.namespace)
	}

	for _, oldConfigMap := range oldConfigMaps.(*corev1.ConfigMapList).Items {
		c.oldConfigMaps[oldConfigMap.GetName()] = oldConfigMap
	}

	// build new config map
	for _, newCm := range c.configMaps {
		c.newConfigMaps[newCm.GetName()] = newCm.ConfigMap
	}
	return nil
}

// add adds the config map
func (c *ConfigMap) add(isPatched bool) (bool, error) {
	isAdded := false
	isUpdated := false
	var err error
	for _, configMap := range c.configMaps {
		if _, ok := c.oldConfigMaps[configMap.GetName()]; !ok {
			c.deployer.Deployer.AddComponent(horizonapi.ConfigMapComponent, configMap)
			isAdded = true
		} else {
			isUpdated, err = c.patch(configMap, isPatched)
			if err != nil {
				return false, errors.Annotatef(err, "patch config map:")
			}
		}
		isPatched = isAdded || isUpdated || isPatched
	}
	if isAdded && !c.config.dryRun {
		err = c.deployer.Deployer.Run()
		if err != nil {
			return false, errors.Annotatef(err, "unable to deploy config map in %s", c.config.namespace)
		}
	}
	return isPatched, nil
}

// get gets the config map
func (c *ConfigMap) get(name string) (interface{}, error) {
	return util.GetConfigMap(c.config.kubeClient, c.config.namespace, name)
}

// list lists all the config maps
func (c *ConfigMap) list() (interface{}, error) {
	return util.ListConfigMaps(c.config.kubeClient, c.config.namespace, c.config.labelSelector)
}

// delete deletes the config map
func (c *ConfigMap) delete(name string) error {
	log.Infof("deleting the config map %s in %s namespace", name, c.config.namespace)
	return util.DeleteConfigMap(c.config.kubeClient, c.config.namespace, name)
}

// remove removes the config map
func (c *ConfigMap) remove() error {
	// compare the old and new config map and delete if needed
	for _, oldConfigMap := range c.oldConfigMaps {
		if _, ok := c.newConfigMaps[oldConfigMap.GetName()]; !ok {
			err := c.delete(oldConfigMap.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete config map %s in namespace %s", oldConfigMap.GetName(), c.config.namespace)
			}
		}
	}
	return nil
}

// patch patches the config map
func (c *ConfigMap) patch(cm interface{}, isPatched bool) (bool, error) {
	configMap := cm.(*components.ConfigMap)
	configMapName := configMap.GetName()
	oldConfigMap := c.oldConfigMaps[configMapName]
	newConfigMap := c.newConfigMaps[configMapName]

	if (!reflect.DeepEqual(newConfigMap.Data, oldConfigMap.Data) || !reflect.DeepEqual(newConfigMap.BinaryData, oldConfigMap.BinaryData)) && !c.config.dryRun {
		log.Infof("updating the config map %s in %s namespace", configMapName, c.config.namespace)
		getCm, err := c.get(configMapName)
		if err != nil {
			return false, errors.Annotatef(err, "unable to get the config map %s in namespace %s", configMapName, c.config.namespace)
		}
		oldLatestConfigMap := getCm.(*corev1.ConfigMap)
		oldLatestConfigMap.Data = newConfigMap.Data
		oldLatestConfigMap.BinaryData = newConfigMap.BinaryData
		_, err = util.UpdateConfigMap(c.config.kubeClient, c.config.namespace, oldLatestConfigMap)
		if err != nil {
			return false, errors.Annotatef(err, "unable to update the config map %s in namespace %s", configMapName, c.config.namespace)
		}
		return true, nil
	}
	return false, nil
}
