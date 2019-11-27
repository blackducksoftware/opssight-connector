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

// Role stores the configuration to add or delete the role
type Role struct {
	config   *CommonConfig
	deployer *util.DeployerHelper
	roles    []*components.Role
	oldRoles map[string]rbacv1.Role
	newRoles map[string]*rbacv1.Role
}

// NewRole returns the role
func NewRole(config *CommonConfig, roles []*components.Role) (*Role, error) {
	deployer, err := util.NewDeployer(config.kubeConfig)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to get deployer object for %s", config.namespace)
	}
	newRoles := append([]*components.Role{}, roles...)
	for i := 0; i < len(newRoles); i++ {
		if !isLabelsExist(config.expectedLabels, newRoles[i].Labels) {
			newRoles = append(newRoles[:i], newRoles[i+1:]...)
			i--
		}
	}
	return &Role{
		config:   config,
		deployer: deployer,
		roles:    newRoles,
		oldRoles: make(map[string]rbacv1.Role, 0),
		newRoles: make(map[string]*rbacv1.Role, 0),
	}, nil
}

// buildNewAndOldObject builds the old and new role
func (c *Role) buildNewAndOldObject() error {
	// build old role
	oldRs, err := c.list()
	if err != nil {
		return errors.Annotatef(err, "unable to get roles for %s", c.config.namespace)
	}

	for _, oldR := range oldRs.(*rbacv1.RoleList).Items {
		c.oldRoles[oldR.GetName()] = oldR
	}

	// build new role
	for _, newR := range c.roles {
		c.newRoles[newR.GetName()] = newR.Role
	}

	return nil
}

// add adds the role
func (c *Role) add(isPatched bool) (bool, error) {
	isAdded := false
	for _, role := range c.roles {
		if _, ok := c.oldRoles[role.GetName()]; !ok {
			c.deployer.Deployer.AddComponent(horizonapi.RoleComponent, role)
			isAdded = true
		} else {
			_, err := c.patch(role, isPatched)
			if err != nil {
				return false, errors.Annotatef(err, "patch role:")
			}
		}
	}
	if isAdded && !c.config.dryRun {
		err := c.deployer.Deployer.Run()
		if err != nil {
			return false, errors.Annotatef(err, "unable to deploy role in %s", c.config.namespace)
		}
	}
	return false, nil
}

// get gets the role
func (c *Role) get(name string) (interface{}, error) {
	return util.GetRole(c.config.kubeClient, c.config.namespace, name)
}

// list lists all the roles
func (c *Role) list() (interface{}, error) {
	return util.ListRoles(c.config.kubeClient, c.config.namespace, c.config.labelSelector)
}

// delete deletes the role
func (c *Role) delete(name string) error {
	log.Infof("deleting the role: %s", name)
	return util.DeleteRole(c.config.kubeClient, c.config.namespace, name)
}

// remove removes the role
func (c *Role) remove() error {
	// compare the old and new role and delete if needed
	for _, oldRole := range c.oldRoles {
		if _, ok := c.newRoles[oldRole.GetName()]; !ok {
			err := c.delete(oldRole.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete role %s in namespace %s", oldRole.GetName(), c.config.namespace)
			}
		}
	}
	return nil
}

// patch patches the role
func (c *Role) patch(cr interface{}, isPatched bool) (bool, error) {
	role := cr.(*components.Role)
	roleName := role.GetName()
	oldrole := c.oldRoles[roleName]
	newRole := c.newRoles[roleName]
	if !reflect.DeepEqual(sortPolicyRule(oldrole.Rules), sortPolicyRule(newRole.Rules)) && !c.config.dryRun {
		log.Infof("updating the role %s for %s namespace", roleName, c.config.namespace)
		getR, err := c.get(roleName)
		if err != nil {
			return false, errors.Annotatef(err, "unable to get the role %s for namespace %s", roleName, c.config.namespace)
		}
		oldLatestRole := getR.(*rbacv1.Role)
		oldLatestRole.Rules = newRole.Rules
		_, err = util.UpdateRole(c.config.kubeClient, c.config.namespace, oldLatestRole)
		if err != nil {
			return false, errors.Annotatef(err, "unable to update the role %s for namespace %s", roleName, c.config.namespace)
		}
	}
	return false, nil
}
