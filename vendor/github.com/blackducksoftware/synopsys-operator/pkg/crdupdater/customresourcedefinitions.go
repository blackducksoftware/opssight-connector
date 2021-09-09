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
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
)

// CustomResourceDefinition stores the configuration to add or delete the custom resource definition
type CustomResourceDefinition struct {
	config                       *CommonConfig
	apiExtensionClient           *apiextensionsclient.Clientset
	deployer                     *util.DeployerHelper
	customResourceDefinitions    []*components.CustomResourceDefinition
	oldCustomResourceDefinitions map[string]apiextensions.CustomResourceDefinition
	newCustomResourceDefinitions map[string]*apiextensions.CustomResourceDefinition
}

// NewCustomResourceDefinition returns the custom resource defintion
func NewCustomResourceDefinition(config *CommonConfig, customResourceDefinitions []*components.CustomResourceDefinition) (*CustomResourceDefinition, error) {
	apiExtensionClient, err := apiextensionsclient.NewForConfig(config.kubeConfig)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to create the api extension client for %s", config.namespace)
	}
	deployer, err := util.NewDeployer(config.kubeConfig)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to get deployer object for %s", config.namespace)
	}
	newCustomResourceDefinitions := append([]*components.CustomResourceDefinition{}, customResourceDefinitions...)
	for i := 0; i < len(newCustomResourceDefinitions); i++ {
		if !isLabelsExist(config.expectedLabels, newCustomResourceDefinitions[i].Labels) {
			newCustomResourceDefinitions = append(newCustomResourceDefinitions[:i], newCustomResourceDefinitions[i+1:]...)
			i--
		}
	}
	return &CustomResourceDefinition{
		config:                       config,
		apiExtensionClient:           apiExtensionClient,
		deployer:                     deployer,
		customResourceDefinitions:    newCustomResourceDefinitions,
		oldCustomResourceDefinitions: make(map[string]apiextensions.CustomResourceDefinition, 0),
		newCustomResourceDefinitions: make(map[string]*apiextensions.CustomResourceDefinition, 0),
	}, nil
}

// buildNewAndOldObject builds the old and new custom resource defintion
func (c *CustomResourceDefinition) buildNewAndOldObject() error {
	// build old customResourceDefinition
	oldCrds, err := c.list()
	if err != nil {
		return errors.Annotatef(err, "unable to get custom resource definitions for %s", c.config.namespace)
	}

	for _, oldCrd := range oldCrds.(*apiextensions.CustomResourceDefinitionList).Items {
		c.oldCustomResourceDefinitions[oldCrd.GetName()] = oldCrd
	}

	// build new customResourceDefinition
	for _, newCrd := range c.customResourceDefinitions {
		c.newCustomResourceDefinitions[newCrd.GetName()] = newCrd.CustomResourceDefinition
	}

	return nil
}

// add adds the custom resource defintion
func (c *CustomResourceDefinition) add(isPatched bool) (bool, error) {
	isAdded := false
	for _, customResourceDefinition := range c.customResourceDefinitions {
		if _, ok := c.oldCustomResourceDefinitions[customResourceDefinition.GetName()]; !ok {
			c.deployer.Deployer.AddComponent(horizonapi.CRDComponent, customResourceDefinition)
			isAdded = true
		} else {
			_, err := c.patch(customResourceDefinition, isPatched)
			if err != nil {
				return false, errors.Annotatef(err, "patch custom resource definition:")
			}
		}
	}
	if isAdded && !c.config.dryRun {
		err := c.deployer.Deployer.Run()
		if err != nil {
			return false, errors.Annotatef(err, "unable to deploy custom resource definition in %s", c.config.namespace)
		}
	}
	return false, nil
}

// get gets the custom resource defintion
func (c *CustomResourceDefinition) get(name string) (interface{}, error) {
	return util.GetCustomResourceDefinition(c.apiExtensionClient, name)
}

// list lists all the custom resource defintions
func (c *CustomResourceDefinition) list() (interface{}, error) {
	return util.ListCustomResourceDefinitions(c.apiExtensionClient, c.config.labelSelector)
}

// delete deletes the custom resource defintion
func (c *CustomResourceDefinition) delete(name string) error {
	log.Infof("deleting the custom resource definition: %s", name)
	return util.DeleteCustomResourceDefinition(c.apiExtensionClient, name)
}

// remove removes the custom resource defintion
func (c *CustomResourceDefinition) remove() error {
	// compare the old and new customResourceDefinition and delete if needed
	for _, oldCustomResourceDefinition := range c.oldCustomResourceDefinitions {
		if _, ok := c.newCustomResourceDefinitions[oldCustomResourceDefinition.GetName()]; !ok {
			err := c.delete(oldCustomResourceDefinition.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete customResourceDefinition %s in namespace %s", oldCustomResourceDefinition.GetName(), c.config.namespace)
			}
		}
	}
	return nil
}

// patch patches the custom resource defintion
func (c *CustomResourceDefinition) patch(cr interface{}, isPatched bool) (bool, error) {
	crd := cr.(*components.CustomResourceDefinition)
	crdName := crd.GetName()
	oldCrd := c.oldCustomResourceDefinitions[crdName]
	newCrd := c.newCustomResourceDefinitions[crdName]
	if oldCrd.Spec.Scope != newCrd.Spec.Scope {
		log.Warnf("updating the %s custom resource definition scope is not supported... please contact the support team to handle it...", crdName)
	}
	return false, nil
}
