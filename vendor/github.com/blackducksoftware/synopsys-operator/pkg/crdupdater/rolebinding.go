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
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	rbacv1 "k8s.io/api/rbac/v1"
)

// RoleBinding stores the configuration to add or delete the role binding
type RoleBinding struct {
	config          *CommonConfig
	deployer        *util.DeployerHelper
	roleBindings    []*components.RoleBinding
	oldRoleBindings map[string]rbacv1.RoleBinding
	newRoleBindings map[string]*rbacv1.RoleBinding
}

// NewRoleBinding returns the role binding
func NewRoleBinding(config *CommonConfig, roleBindings []*components.RoleBinding) (*RoleBinding, error) {
	deployer, err := util.NewDeployer(config.kubeConfig)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to get deployer object for %s", config.namespace)
	}
	newRoleBindings := append([]*components.RoleBinding{}, roleBindings...)
	for i := 0; i < len(newRoleBindings); i++ {
		if !isLabelsExist(config.expectedLabels, newRoleBindings[i].Labels) {
			newRoleBindings = append(newRoleBindings[:i], newRoleBindings[i+1:]...)
			i--
		}
	}
	return &RoleBinding{
		config:          config,
		deployer:        deployer,
		roleBindings:    newRoleBindings,
		oldRoleBindings: make(map[string]rbacv1.RoleBinding, 0),
		newRoleBindings: make(map[string]*rbacv1.RoleBinding, 0),
	}, nil
}

// buildNewAndOldObject builds the old and new role binding
func (c *RoleBinding) buildNewAndOldObject() error {
	// build old role binding
	oldRbs, err := c.list()
	if err != nil {
		return errors.Annotatef(err, "unable to get role bindings for %s", c.config.namespace)
	}

	for _, oldRb := range oldRbs.(*rbacv1.RoleBindingList).Items {
		c.oldRoleBindings[oldRb.GetName()] = oldRb
	}

	// build new role binding
	for _, newRb := range c.roleBindings {
		c.newRoleBindings[newRb.GetName()] = newRb.RoleBinding
	}

	return nil
}

// add adds the role binding
func (c *RoleBinding) add(isPatched bool) (bool, error) {
	isAdded := false
	var err error
	for _, roleBinding := range c.roleBindings {
		if _, ok := c.oldRoleBindings[roleBinding.GetName()]; !ok {
			c.deployer.Deployer.AddComponent(horizonapi.RoleBindingComponent, roleBinding)
			isAdded = true
		} else {
			_, err = c.patch(roleBinding, isPatched)
			if err != nil {
				return false, errors.Annotatef(err, "patch role")
			}
		}
	}
	if isAdded && !c.config.dryRun {
		err := c.deployer.Deployer.Run()
		if err != nil {
			return false, errors.Annotatef(err, "unable to deploy role binding in %s", c.config.namespace)
		}
	}
	return false, nil
}

// get gets the role binding
func (c *RoleBinding) get(name string) (interface{}, error) {
	return util.GetRoleBinding(c.config.kubeClient, c.config.namespace, name)
}

// list lists all the role bindings
func (c *RoleBinding) list() (interface{}, error) {
	return util.ListRoleBindings(c.config.kubeClient, c.config.namespace, c.config.labelSelector)
}

// delete deletes the role binding
func (c *RoleBinding) delete(name string) error {
	log.Infof("deleting the role binding: %s", name)
	return util.DeleteRoleBinding(c.config.kubeClient, c.config.namespace, name)
}

// remove removes the role binding
func (c *RoleBinding) remove() error {
	// compare the old and new role binding and delete if needed
	for _, oldRoleBinding := range c.oldRoleBindings {
		if _, ok := c.newRoleBindings[oldRoleBinding.GetName()]; !ok {
			err := c.delete(oldRoleBinding.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete role binding %s in namespace %s", oldRoleBinding.GetName(), c.config.namespace)
			}
		}
	}
	return nil
}

// patch patches the role binding
func (c *RoleBinding) patch(crb interface{}, isPatched bool) (bool, error) {
	roleBinding := crb.(*components.RoleBinding)
	roleBindingName := roleBinding.GetName()
	oldroleBinding := c.oldRoleBindings[roleBindingName]
	newRoleBinding := c.newRoleBindings[roleBindingName]
	isChanged := false
	// check for role ref changes
	if (!strings.EqualFold(oldroleBinding.RoleRef.Name, newRoleBinding.RoleRef.Name) || !strings.EqualFold(oldroleBinding.RoleRef.Kind, newRoleBinding.RoleRef.Kind)) && !c.config.dryRun {
		isChanged = true
	}
	// check for subject changes
	for _, subject := range newRoleBinding.Subjects {
		if !util.IsSubjectExist(oldroleBinding.Subjects, subject.Namespace, subject.Name) {
			oldroleBinding.Subjects = append(oldroleBinding.Subjects, rbacv1.Subject{Name: subject.Name, Namespace: subject.Namespace, Kind: "ServiceAccount"})
			isChanged = true
		}
	}
	if isChanged {
		log.Infof("updating the role binding %s for %s namespace", roleBindingName, c.config.namespace)
		getRb, err := c.get(roleBindingName)
		if err != nil {
			return false, errors.Annotatef(err, "unable to get the role binding %s for namespace %s", roleBindingName, c.config.namespace)
		}
		oldLatestRoleBinding := getRb.(*rbacv1.RoleBinding)
		oldLatestRoleBinding.RoleRef = newRoleBinding.RoleRef
		oldLatestRoleBinding.Subjects = oldroleBinding.Subjects
		_, err = util.UpdateRoleBinding(c.config.kubeClient, c.config.namespace, oldLatestRoleBinding)
		if err != nil {
			return false, errors.Annotate(err, fmt.Sprintf("failed to update %s role binding for namespace %s", roleBinding.GetName(), c.config.namespace))
		}
	}
	return false, nil
}
