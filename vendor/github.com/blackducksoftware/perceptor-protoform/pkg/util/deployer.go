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

package util

import (
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/horizon/pkg/deployer"
	"github.com/blackducksoftware/perceptor-protoform/pkg/api"
	"k8s.io/client-go/rest"
)

// Deployer will contain the deployer specification
type Deployer struct {
	Deployer *deployer.Deployer
}

// NewDeployer will create the horizon deployer
func NewDeployer(config *rest.Config) (*Deployer, error) {
	deployer, err := deployer.NewDeployer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployer: %v", err)
	}
	return &Deployer{Deployer: deployer}, nil
}

func (i *Deployer) addNS(ns string) {
	comp := components.NewNamespace(horizonapi.NamespaceConfig{
		Name: ns,
	})

	i.Deployer.AddNamespace(comp)
}

func (i *Deployer) addRCs(list []*components.ReplicationController) {
	if len(list) > 0 {
		for _, rc := range list {
			i.Deployer.AddReplicationController(rc)
		}
	}
}

func (i *Deployer) addSvcs(list []*components.Service) {
	if len(list) > 0 {
		for _, svc := range list {
			i.Deployer.AddService(svc)
		}
	}
}

func (i *Deployer) addCMs(list []*components.ConfigMap) {
	if len(list) > 0 {
		for _, cm := range list {
			i.Deployer.AddConfigMap(cm)
		}
	}
}

func (i *Deployer) addSAs(list []*components.ServiceAccount) {
	if len(list) > 0 {
		for _, sa := range list {
			i.Deployer.AddServiceAccount(sa)
		}
	}
}

func (i *Deployer) addCRs(list []*components.ClusterRole) {
	if len(list) > 0 {
		for _, cr := range list {
			i.Deployer.AddClusterRole(cr)
		}
	}
}

func (i *Deployer) addCRBs(list []*components.ClusterRoleBinding) {
	if len(list) > 0 {
		for _, crb := range list {
			i.Deployer.AddClusterRoleBinding(crb)
		}
	}
}

func (i *Deployer) addDeploys(list []*components.Deployment) {
	if len(list) > 0 {
		for _, d := range list {
			i.Deployer.AddDeployment(d)
		}
	}
}

func (i *Deployer) addSecrets(list []*components.Secret) {
	if len(list) > 0 {
		for _, s := range list {
			i.Deployer.AddSecret(s)
		}
	}
}

func (i *Deployer) addDefaultController(namespace string) {
	i.Deployer.AddController("Pod List Controller", NewPodListController(namespace))
}

// AddController will add the controller to the deployer
func (i *Deployer) AddController(name string, c horizonapi.DeployerControllerInterface) {
	i.Deployer.AddController(name, c)
}

// PreDeploy will provide the deploy objects
func (i *Deployer) PreDeploy(components *api.ComponentList, namespace string) {
	if components != nil {
		i.addNS(namespace)
		i.addRCs(components.ReplicationControllers)
		i.addSvcs(components.Services)
		i.addCMs(components.ConfigMaps)
		i.addSAs(components.ServiceAccounts)
		i.addCRs(components.ClusterRoles)
		i.addCRBs(components.ClusterRoleBindings)
		i.addDeploys(components.Deployments)
		i.addSecrets(components.Secrets)
		i.addDefaultController(namespace)
	}
}

// Run will run the deployer
func (i *Deployer) Run() error {
	return i.Deployer.Run()
}

// StartControllers will start the controllers
func (i *Deployer) StartControllers() {
	stopCh := make(chan struct{})
	go i.Deployer.StartControllers(stopCh)
}
