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
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// CommonConfig stores the common configuration for add, patch or remove the components for update events
type CommonConfig struct {
	kubeConfig                *rest.Config
	kubeClient                *kubernetes.Clientset
	dryRun                    bool
	isPatched                 bool
	namespace                 string
	version                   string
	components                *api.ComponentList
	labelSelector             string
	expectedLabels            map[string]label
	controllers               map[string]horizonapi.DeployerControllerInterface
	isClusterLevelPermEnabled bool // for synopsysctl, it will be true because user might have cluster admin like privilege to add/delete crd, cluster role and role bindings.
	// for others, it will be based on the pod's service account
}

// NewCRUDComponents returns the common configuration which will be used to add, patch or remove the components
func NewCRUDComponents(kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, dryRun bool, isPatched bool, namespace string,
	version string, components *api.ComponentList, labelSelector string, isClusterLevelPermEnabled bool) *CommonConfig {
	return &CommonConfig{
		kubeConfig:                kubeConfig,
		kubeClient:                kubeClient,
		dryRun:                    dryRun,
		isPatched:                 isPatched,
		namespace:                 namespace,
		version:                   version,
		components:                components,
		labelSelector:             labelSelector,
		expectedLabels:            getLabelsMap(labelSelector),
		controllers:               make(map[string]horizonapi.DeployerControllerInterface, 0),
		isClusterLevelPermEnabled: isClusterLevelPermEnabled,
	}
}

// AddController will add the controller to the updater
func (c *CommonConfig) AddController(name string, controller horizonapi.DeployerControllerInterface) {
	c.controllers[name] = controller
}

// CRUDComponents will add, update or delete components
func (c *CommonConfig) CRUDComponents() (bool, []error) {
	// log.Debugf("expected labels: %+v", c.expectedLabels)
	var errors []error
	updater := NewUpdater(c.dryRun, c.isPatched)

	// namespace
	namespaces, err := NewNamespace(c)
	if err != nil {
		errors = append(errors, fmt.Errorf("unable to create new namespace updater due to %+v", err))
	} else {
		updater.AddUpdater(namespaces)
	}

	// service account
	serviceAccounts, err := NewServiceAccount(c, c.components.ServiceAccounts)
	if err != nil {
		errors = append(errors, fmt.Errorf("unable to create new service account updater due to %+v", err))
	} else {
		updater.AddUpdater(serviceAccounts)
	}

	// cluster role
	if c.isClusterLevelPermEnabled || len(c.components.ClusterRoles) > 0 {
		clusterRoles, err := NewClusterRole(c, c.components.ClusterRoles)
		if err != nil {
			errors = append(errors, fmt.Errorf("unable to create new cluster role updater due to %+v", err))
		} else {
			updater.AddUpdater(clusterRoles)
		}
	}

	// cluster role binding
	if c.isClusterLevelPermEnabled || len(c.components.ClusterRoleBindings) > 0 {
		clusterRoleBindings, err := NewClusterRoleBinding(c, c.components.ClusterRoleBindings)
		if err != nil {
			errors = append(errors, fmt.Errorf("unable to create new cluster role binding updater due to %+v", err))
		} else {
			updater.AddUpdater(clusterRoleBindings)
		}
	}

	// role
	roles, err := NewRole(c, c.components.Roles)
	if err != nil {
		errors = append(errors, fmt.Errorf("unable to create new role updater due to %+v", err))
	} else {
		updater.AddUpdater(roles)
	}

	// role binding
	roleBindings, err := NewRoleBinding(c, c.components.RoleBindings)
	if err != nil {
		errors = append(errors, fmt.Errorf("unable to create new role binding updater due to %+v", err))
	} else {
		updater.AddUpdater(roleBindings)
	}

	// config map
	configMaps, err := NewConfigMap(c, c.components.ConfigMaps)
	if err != nil {
		errors = append(errors, fmt.Errorf("unable to create new config map updater due to %+v", err))
	} else {
		updater.AddUpdater(configMaps)
	}

	// secret
	secrets, err := NewSecret(c, c.components.Secrets)
	if err != nil {
		errors = append(errors, fmt.Errorf("unable to create new secret updater due to %+v", err))
	} else {
		updater.AddUpdater(secrets)
	}

	// persistent volume claim
	pvcs, err := NewPersistentVolumeClaim(c, c.components.PersistentVolumeClaims)
	if err != nil {
		errors = append(errors, fmt.Errorf("unable to create new persistent volume claim updater due to %+v", err))
	} else {
		updater.AddUpdater(pvcs)
	}

	// service
	services, err := NewService(c, c.components.Services)
	if err != nil {
		errors = append(errors, fmt.Errorf("unable to create new service updater due to %+v", err))
	} else {
		updater.AddUpdater(services)
	}

	// replication controller
	rcs, err := NewReplicationController(c, c.components.ReplicationControllers)
	if err != nil {
		errors = append(errors, fmt.Errorf("unable to create new replication controller updater due to %+v", err))
	} else {
		updater.AddUpdater(rcs)
	}

	// deployment
	deployments, err := NewDeployment(c, c.components.Deployments)
	if err != nil {
		errors = append(errors, fmt.Errorf("unable to create new deployment updater due to %+v", err))
	} else {
		updater.AddUpdater(deployments)
	}

	// OpenShift routes
	routes, err := NewRoute(c, c.components.Routes)
	if err != nil {
		errors = append(errors, fmt.Errorf("unable to create route updater due to %+v", err))
	} else if routes != nil {
		updater.AddUpdater(routes)
	}

	// execute updates for all added components
	isPatched, err := updater.Update()
	if err != nil {
		errors = append(errors, err)
	}

	if !c.dryRun {
		deployer, err := util.NewDeployer(c.kubeConfig)
		if err != nil {
			errors = append(errors, fmt.Errorf("unable to get deployer object for %+v", err))
		}

		// add all controllers to the deployer object
		for name, controller := range c.controllers {
			deployer.AddController(name, controller)
		}

		// start the controller
		deployer.StartControllers()
	}

	return isPatched, errors
}
