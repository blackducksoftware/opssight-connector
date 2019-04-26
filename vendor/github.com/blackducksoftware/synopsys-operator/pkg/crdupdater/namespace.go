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
)

// Namespace stores the configuration to add or delete the namespace
type Namespace struct {
	config   *CommonConfig
	deployer *util.DeployerHelper
}

// NewNamespace returns the namespace
func NewNamespace(config *CommonConfig) (*Namespace, error) {
	deployer, err := util.NewDeployer(config.kubeConfig)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to get deployer object for %s", config.namespace)
	}
	return &Namespace{
		config:   config,
		deployer: deployer,
	}, nil
}

// buildNewAndOldObject builds the old and new namespace
func (n *Namespace) buildNewAndOldObject() error {
	return nil
}

// add adds the namespace
func (n *Namespace) add(isPatched bool) (bool, error) {
	_, err := n.get(n.config.namespace)
	if err != nil {
		n.deployer.Deployer.AddNamespace(components.NewNamespace(horizonapi.NamespaceConfig{Name: n.config.namespace}))
		n.deployer.Deployer.Run()
	}
	return false, nil
}

// get get the namespace
func (n *Namespace) get(name string) (interface{}, error) {
	return util.GetNamespace(n.config.kubeClient, name)
}

// list lists all the namespaces
func (n *Namespace) list() (interface{}, error) {
	return nil, nil
}

// delete deletes the namespace
func (n *Namespace) delete(name string) error {
	return util.DeleteNamespace(n.config.kubeClient, n.config.namespace)
}

// remove removes the namespace
func (n *Namespace) remove() error {
	return nil
}

// patch patches the namespace
func (n *Namespace) patch(ns interface{}, isPatched bool) (bool, error) {
	return false, nil
}
