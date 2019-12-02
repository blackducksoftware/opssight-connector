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

// ClusterRoleBinding stores the configuration to add or delete the cluster role binding
type ClusterRoleBinding struct {
	config                 *CommonConfig
	deployer               *util.DeployerHelper
	clusterRoleBindings    []*components.ClusterRoleBinding
	oldClusterRoleBindings map[string]rbacv1.ClusterRoleBinding
	newClusterRoleBindings map[string]*rbacv1.ClusterRoleBinding
}

// NewClusterRoleBinding returns the cluster role binding
func NewClusterRoleBinding(config *CommonConfig, clusterRoleBindings []*components.ClusterRoleBinding) (*ClusterRoleBinding, error) {
	deployer, err := util.NewDeployer(config.kubeConfig)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to get deployer object for %s", config.namespace)
	}
	newClusterRoleBindings := append([]*components.ClusterRoleBinding{}, clusterRoleBindings...)
	for i := 0; i < len(newClusterRoleBindings); i++ {
		if !isLabelsExist(config.expectedLabels, newClusterRoleBindings[i].Labels) {
			newClusterRoleBindings = append(newClusterRoleBindings[:i], newClusterRoleBindings[i+1:]...)
			i--
		}
	}
	return &ClusterRoleBinding{
		config:                 config,
		deployer:               deployer,
		clusterRoleBindings:    newClusterRoleBindings,
		oldClusterRoleBindings: make(map[string]rbacv1.ClusterRoleBinding, 0),
		newClusterRoleBindings: make(map[string]*rbacv1.ClusterRoleBinding, 0),
	}, nil
}

// buildNewAndOldObject builds the old and new cluster role binding
func (c *ClusterRoleBinding) buildNewAndOldObject() error {
	// build old cluster role binding
	oldCrbs, err := c.list()
	if err != nil {
		return errors.Annotatef(err, "unable to get cluster role bindings for %s", c.config.namespace)
	}

	for _, oldCrb := range oldCrbs.(*rbacv1.ClusterRoleBindingList).Items {
		c.oldClusterRoleBindings[oldCrb.GetName()] = oldCrb
	}

	// build new cluster role binding
	for _, newCrb := range c.clusterRoleBindings {
		c.newClusterRoleBindings[newCrb.GetName()] = newCrb.ClusterRoleBinding
	}

	return nil
}

// add adds the cluster role binding
func (c *ClusterRoleBinding) add(isPatched bool) (bool, error) {
	isAdded := false
	var err error
	for _, clusterRoleBinding := range c.clusterRoleBindings {
		if _, ok := c.oldClusterRoleBindings[clusterRoleBinding.GetName()]; !ok {
			c.deployer.Deployer.AddComponent(horizonapi.ClusterRoleBindingComponent, clusterRoleBinding)
			isAdded = true
		} else {
			_, err = c.patch(clusterRoleBinding, isPatched)
			if err != nil {
				return false, errors.Annotatef(err, "patch cluster role")
			}
		}
	}
	if isAdded && !c.config.dryRun {
		err := c.deployer.Deployer.Run()
		if err != nil {
			return false, errors.Annotatef(err, "unable to deploy cluster role binding in %s", c.config.namespace)
		}
	}
	return false, nil
}

// get gets the cluster role binding
func (c *ClusterRoleBinding) get(name string) (interface{}, error) {
	return util.GetClusterRoleBinding(c.config.kubeClient, name)
}

// list lists all the cluster role bindings
func (c *ClusterRoleBinding) list() (interface{}, error) {
	return util.ListClusterRoleBindings(c.config.kubeClient, c.config.labelSelector)
}

// delete deletes the cluster role binding
func (c *ClusterRoleBinding) delete(name string) error {
	log.Infof("deleting the cluster role binding: %s", name)
	return util.DeleteClusterRoleBinding(c.config.kubeClient, name)
}

// remove removes the cluster role binding
func (c *ClusterRoleBinding) remove() error {
	// compare the old and new cluster role binding and delete if needed
	for _, oldClusterRoleBinding := range c.oldClusterRoleBindings {
		clusterRoleBindingName := oldClusterRoleBinding.GetName()
		if _, ok := c.newClusterRoleBindings[clusterRoleBindingName]; !ok {
			// check whether any subject present for other namespace before deleting them
			newSubjects := []rbacv1.Subject{}
			for _, subject := range oldClusterRoleBinding.Subjects {
				isExist := util.IsSubjectExistForOtherNamespace(subject, c.config.namespace)
				if isExist {
					newSubjects = append(newSubjects, subject)
				}
			}
			if len(newSubjects) > 0 {
				getCrb, err := c.get(clusterRoleBindingName)
				if err != nil {
					return errors.Annotatef(err, "unable to get the cluster role binding %s for namespace %s", clusterRoleBindingName, c.config.namespace)
				}
				oldLatestClusterRoleBinding := getCrb.(*rbacv1.ClusterRoleBinding)
				oldLatestClusterRoleBinding.Subjects = newSubjects
				// update the cluster role binding to remove the old cluster role binding subject
				_, err = util.UpdateClusterRoleBinding(c.config.kubeClient, oldLatestClusterRoleBinding)
				if err != nil {
					return errors.Annotate(err, fmt.Sprintf("failed to update %s cluster role binding for namespace %s", clusterRoleBindingName, c.config.namespace))
				}
			} else {
				err := c.delete(clusterRoleBindingName)
				if err != nil {
					return errors.Annotatef(err, "unable to delete cluster role binding %s in namespace %s", clusterRoleBindingName, c.config.namespace)
				}
			}
		}
	}
	return nil
}

// patch patches the cluster role binding
func (c *ClusterRoleBinding) patch(crb interface{}, isPatched bool) (bool, error) {
	clusterRoleBinding := crb.(*components.ClusterRoleBinding)
	clusterRoleBindingName := clusterRoleBinding.GetName()
	oldclusterRoleBinding := c.oldClusterRoleBindings[clusterRoleBindingName]
	newClusterRoleBinding := c.newClusterRoleBindings[clusterRoleBindingName]
	isChanged := false
	// check for role ref changes
	if (!strings.EqualFold(oldclusterRoleBinding.RoleRef.Name, newClusterRoleBinding.RoleRef.Name) || !strings.EqualFold(oldclusterRoleBinding.RoleRef.Kind, newClusterRoleBinding.RoleRef.Kind)) && !c.config.dryRun {
		isChanged = true
	}
	// check for subject changes
	for _, subject := range newClusterRoleBinding.Subjects {
		if !util.IsSubjectExist(oldclusterRoleBinding.Subjects, subject.Namespace, subject.Name) {
			oldclusterRoleBinding.Subjects = append(oldclusterRoleBinding.Subjects, rbacv1.Subject{Name: subject.Name, Namespace: subject.Namespace, Kind: "ServiceAccount"})
			isChanged = true
		}
	}
	if isChanged {
		log.Infof("updating the cluster role binding %s for %s namespace", clusterRoleBindingName, c.config.namespace)
		getCrb, err := c.get(clusterRoleBindingName)
		if err != nil {
			return false, errors.Annotatef(err, "unable to get the cluster role binding %s for namespace %s", clusterRoleBindingName, c.config.namespace)
		}
		oldLatestClusterRoleBinding := getCrb.(*rbacv1.ClusterRoleBinding)
		oldLatestClusterRoleBinding.RoleRef = newClusterRoleBinding.RoleRef
		oldLatestClusterRoleBinding.Subjects = oldclusterRoleBinding.Subjects
		_, err = util.UpdateClusterRoleBinding(c.config.kubeClient, oldLatestClusterRoleBinding)
		if err != nil {
			return false, errors.Annotate(err, fmt.Sprintf("failed to update %s cluster role binding for namespace %s", clusterRoleBindingName, c.config.namespace))
		}
	}
	return false, nil
}
