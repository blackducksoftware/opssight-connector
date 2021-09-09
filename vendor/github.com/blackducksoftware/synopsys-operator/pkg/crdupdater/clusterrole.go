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
	rbacv1 "k8s.io/api/rbac/v1"
)

// ClusterRole stores the configuration to add or delete the cluster role
type ClusterRole struct {
	config          *CommonConfig
	deployer        *util.DeployerHelper
	clusterRoles    []*components.ClusterRole
	oldClusterRoles map[string]rbacv1.ClusterRole
	newClusterRoles map[string]*rbacv1.ClusterRole
}

// NewClusterRole returns the cluster role
func NewClusterRole(config *CommonConfig, clusterRoles []*components.ClusterRole) (*ClusterRole, error) {
	deployer, err := util.NewDeployer(config.kubeConfig)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to get deployer object for %s", config.namespace)
	}
	newClusterRoles := append([]*components.ClusterRole{}, clusterRoles...)
	for i := 0; i < len(newClusterRoles); i++ {
		if !isLabelsExist(config.expectedLabels, newClusterRoles[i].Labels) {
			newClusterRoles = append(newClusterRoles[:i], newClusterRoles[i+1:]...)
			i--
		}
	}
	return &ClusterRole{
		config:          config,
		deployer:        deployer,
		clusterRoles:    newClusterRoles,
		oldClusterRoles: make(map[string]rbacv1.ClusterRole, 0),
		newClusterRoles: make(map[string]*rbacv1.ClusterRole, 0),
	}, nil
}

// buildNewAndOldObject builds the old and new cluster role
func (c *ClusterRole) buildNewAndOldObject() error {
	// build old cluster role
	oldCrs, err := c.list()
	if err != nil {
		return errors.Annotatef(err, "unable to get cluster roles for %s", c.config.namespace)
	}

	for _, oldCr := range oldCrs.(*rbacv1.ClusterRoleList).Items {
		c.oldClusterRoles[oldCr.GetName()] = oldCr
	}

	// build new cluster role
	for _, newCr := range c.clusterRoles {
		c.newClusterRoles[newCr.GetName()] = newCr.ClusterRole
	}

	return nil
}

// add adds the cluster role
func (c *ClusterRole) add(isPatched bool) (bool, error) {
	isAdded := false
	for _, clusterRole := range c.clusterRoles {
		if _, ok := c.oldClusterRoles[clusterRole.GetName()]; !ok {
			c.deployer.Deployer.AddComponent(horizonapi.ClusterRoleComponent, clusterRole)
			isAdded = true
		} else {
			_, err := c.patch(clusterRole, isPatched)
			if err != nil {
				return false, errors.Annotatef(err, "patch cluster role:")
			}
		}
	}
	if isAdded && !c.config.dryRun {
		err := c.deployer.Deployer.Run()
		if err != nil {
			return false, errors.Annotatef(err, "unable to deploy cluster role in %s", c.config.namespace)
		}
	}
	return false, nil
}

// get gets the cluster role
func (c *ClusterRole) get(name string) (interface{}, error) {
	return util.GetClusterRole(c.config.kubeClient, name)
}

// list lists all the cluster roles
func (c *ClusterRole) list() (interface{}, error) {
	return util.ListClusterRoles(c.config.kubeClient, c.config.labelSelector)
}

// delete deletes the cluster role
func (c *ClusterRole) delete(name string) error {
	log.Infof("deleting the cluster role: %s", name)
	return util.DeleteClusterRole(c.config.kubeClient, name)
}

// remove removes the cluster role
func (c *ClusterRole) remove() error {
	// compare the old and new cluster role and delete if needed
	for _, oldClusterRole := range c.oldClusterRoles {
		if _, ok := c.newClusterRoles[oldClusterRole.GetName()]; !ok {
			err := c.delete(oldClusterRole.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete cluster role %s in namespace %s", oldClusterRole.GetName(), c.config.namespace)
			}
		}
	}
	return nil
}

// patch patches the cluster role
func (c *ClusterRole) patch(cr interface{}, isPatched bool) (bool, error) {
	clusterRole := cr.(*components.ClusterRole)
	clusterRoleName := clusterRole.GetName()
	oldclusterRole := c.oldClusterRoles[clusterRoleName]
	newClusterRole := c.newClusterRoles[clusterRoleName]
	if !reflect.DeepEqual(sortPolicyRule(oldclusterRole.Rules), sortPolicyRule(newClusterRole.Rules)) && !c.config.dryRun {
		log.Infof("updating the cluster role %s for %s namespace", clusterRoleName, c.config.namespace)
		getCr, err := c.get(clusterRoleName)
		if err != nil {
			return false, errors.Annotatef(err, "unable to get the cluster role %s for namespace %s", clusterRoleName, c.config.namespace)
		}
		oldLatestClusterRole := getCr.(*rbacv1.ClusterRole)
		oldLatestClusterRole.Rules = newClusterRole.Rules
		_, err = util.UpdateClusterRole(c.config.kubeClient, oldLatestClusterRole)
		if err != nil {
			return false, errors.Annotatef(err, "unable to update the cluster role %s for namespace %s", clusterRoleName, c.config.namespace)
		}
	}
	return false, nil
}
