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
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
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
	ns, err := n.get(n.config.namespace)
	if err != nil {
		namespace := components.NewNamespace(
			horizonapi.NamespaceConfig{
				Name:      n.config.namespace,
				Namespace: n.config.namespace,
			})
		labels := make(map[string]string, 0)
		labels["owner"] = util.OperatorName
		var app, name string
		if appVal, ok := n.config.expectedLabels["app"]; ok {
			app = appVal.value[0]
		}
		if nameVal, ok := n.config.expectedLabels["name"]; ok {
			name = nameVal.value[0]
		}
		if len(app) > 0 && len(name) > 0 && len(n.config.version) > 0 {
			labels[fmt.Sprintf("synopsys.com/%s.%s", app, name)] = n.config.version
		} else if len(n.config.version) > 0 {
			log.Errorf("unable to add the Synopsys label because app: %s, name: %s, version: %s is missing", app, name, n.config.namespace)
		}
		namespace.AddLabels(labels)
		n.deployer.Deployer.AddComponent(horizonapi.NamespaceComponent, namespace)
		err := n.deployer.Deployer.Run()
		if err != nil {
			return false, fmt.Errorf("unable to create namespace due to %+v", err)
		}
		return false, nil
	}
	return n.patch(ns, false)
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
	namespace := ns.(*corev1.Namespace)
	namespace.Labels = util.InitLabels(namespace.Labels)
	var app, name string
	if appVal, ok := n.config.expectedLabels["app"]; ok {
		app = appVal.value[0]
	}
	if nameVal, ok := n.config.expectedLabels["name"]; ok {
		name = nameVal.value[0]
	}
	if val, ok := namespace.Labels[fmt.Sprintf("synopsys.com/%s.%s", app, name)]; !ok || val != n.config.version {
		if len(app) > 0 && len(name) > 0 && len(n.config.version) > 0 {
			log.Debugf("patch namespace for synopsys label in namespace '%s'", namespace.Name)

			getN, err := n.get(namespace.GetName())
			if err != nil {
				return false, errors.Annotatef(err, "unable to get the namespace %s", namespace.GetName())
			}
			oldNamespace := getN.(*corev1.Namespace)
			oldNamespace.Labels = util.InitLabels(oldNamespace.Labels)
			oldNamespace.Labels[fmt.Sprintf("synopsys.com/%s.%s", app, name)] = n.config.version

			_, err = util.UpdateNamespace(n.config.kubeClient, oldNamespace)
			if err != nil {
				return false, fmt.Errorf("unable to update namespace %s due to %+v", namespace.GetName(), err)
			}
		} else if len(n.config.version) > 0 {
			log.Errorf("unable to update the Synopsys label because app: %s, name: %s, version: %s is missing", app, name, n.config.namespace)
		}
	}
	return false, nil
}
