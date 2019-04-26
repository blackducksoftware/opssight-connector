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
	"strings"

	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// ReplicationController stores the configuration to add or delete the replication controller object
type ReplicationController struct {
	config                    *CommonConfig
	deployer                  *util.DeployerHelper
	replicationControllers    []*components.ReplicationController
	oldReplicationControllers map[string]corev1.ReplicationController
	newReplicationControllers map[string]*corev1.ReplicationController
}

// NewReplicationController returns the replication controller
func NewReplicationController(config *CommonConfig, replicationControllers []*components.ReplicationController) (*ReplicationController, error) {
	deployer, err := util.NewDeployer(config.kubeConfig)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to get deployer object for %s", config.namespace)
	}
	newReplicationControllers := append([]*components.ReplicationController{}, replicationControllers...)
	for i := 0; i < len(newReplicationControllers); i++ {
		if !isLabelsExist(config.expectedLabels, newReplicationControllers[i].GetObj().Labels) {
			// log.Debugf("removing::expected Labels: %+v, actual Labels: %+v", config.expectedLabels, newReplicationControllers[i].GetObj().Labels)
			newReplicationControllers = append(newReplicationControllers[:i], newReplicationControllers[i+1:]...)
			i--
		}
	}
	return &ReplicationController{
		config:                    config,
		deployer:                  deployer,
		replicationControllers:    newReplicationControllers,
		oldReplicationControllers: make(map[string]corev1.ReplicationController, 0),
		newReplicationControllers: make(map[string]*corev1.ReplicationController, 0),
	}, nil
}

// buildNewAndOldObject builds the old and new replication controller
func (r *ReplicationController) buildNewAndOldObject() error {
	// build old replication controller
	oldRCs, err := r.list()
	if err != nil {
		return errors.Annotatef(err, "unable to get replication controllers for %s", r.config.namespace)
	}
	for _, oldRC := range oldRCs.(*corev1.ReplicationControllerList).Items {
		r.oldReplicationControllers[oldRC.GetName()] = oldRC
	}

	// build new replication controller
	for _, newRc := range r.replicationControllers {
		newReplicationControllerKube, err := newRc.ToKube()
		if err != nil {
			return errors.Annotatef(err, "unable to convert replication controller %s to kube %s", newRc.GetName(), r.config.namespace)
		}
		r.newReplicationControllers[newRc.GetName()] = newReplicationControllerKube.(*corev1.ReplicationController)
	}

	return nil
}

// add adds the replication controller
func (r *ReplicationController) add(isPatched bool) (bool, error) {
	isAdded := false
	for _, replicationController := range r.replicationControllers {
		if _, ok := r.oldReplicationControllers[replicationController.GetName()]; !ok {
			r.deployer.Deployer.AddReplicationController(replicationController)
			isAdded = true
		} else {
			_, err := r.patch(replicationController, isPatched)
			if err != nil {
				return false, errors.Annotatef(err, "patch replication controller:")
			}
		}
	}
	if isAdded && !r.config.dryRun {
		err := r.deployer.Deployer.Run()
		if err != nil {
			return false, errors.Annotatef(err, "unable to deploy replication controller in %s", r.config.namespace)
		}
	}
	return false, nil
}

// get gets the replication controller
func (r *ReplicationController) get(name string) (interface{}, error) {
	return util.GetReplicationController(r.config.kubeClient, r.config.namespace, name)
}

// list lists all the replication controllers
func (r *ReplicationController) list() (interface{}, error) {
	return util.ListReplicationControllers(r.config.kubeClient, r.config.namespace, r.config.labelSelector)
}

// delete deletes the replication controller
func (r *ReplicationController) delete(name string) error {
	log.Infof("deleting the replication controller %s in %s namespace", name, r.config.namespace)
	return util.DeleteReplicationController(r.config.kubeClient, r.config.namespace, name)
}

// remove removes the replication controller
func (r *ReplicationController) remove() error {
	// compare the old and new replication controller and delete if needed
	for _, oldReplicationController := range r.oldReplicationControllers {
		if _, ok := r.newReplicationControllers[oldReplicationController.GetName()]; !ok {
			err := r.delete(oldReplicationController.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete replication controller %s in namespace %s", oldReplicationController.GetName(), r.config.namespace)
			}
		}
	}
	return nil
}

// replicationControllerComparator used to compare Replication controller attributes
type replicationControllerComparator struct {
	Image    string
	Replicas *int32
	MinCPU   *resource.Quantity
	MaxCPU   *resource.Quantity
	MinMem   *resource.Quantity
	MaxMem   *resource.Quantity
}

// patch patches the replication controller
func (r *ReplicationController) patch(rc interface{}, isPatched bool) (bool, error) {
	replicationController := rc.(*components.ReplicationController)
	// check isPatched, why?
	// if there is any configuration change, irrespective of comparing any changes, patch the replication controller
	if isPatched && !r.config.dryRun {
		log.Infof("updating the replication controller %s in %s namespace", replicationController.GetName(), r.config.namespace)
		err := util.PatchReplicationController(r.config.kubeClient, r.oldReplicationControllers[replicationController.GetName()], *r.newReplicationControllers[replicationController.GetName()])
		if err != nil {
			return false, errors.Annotatef(err, "unable to patch replication controller %s in namespace %s", replicationController.GetName(), r.config.namespace)
		}
		return false, nil
	}

	// check whether the replication controller or its container got changed
	isChanged := false
	for _, oldContainer := range r.oldReplicationControllers[replicationController.GetName()].Spec.Template.Spec.Containers {
		for _, newContainer := range r.newReplicationControllers[replicationController.GetName()].Spec.Template.Spec.Containers {
			if strings.EqualFold(oldContainer.Name, newContainer.Name) && !r.config.dryRun &&
				!reflect.DeepEqual(
					replicationControllerComparator{
						Image:    oldContainer.Image,
						Replicas: r.oldReplicationControllers[replicationController.GetName()].Spec.Replicas,
						MinCPU:   oldContainer.Resources.Requests.Cpu(),
						MaxCPU:   oldContainer.Resources.Limits.Cpu(),
						MinMem:   oldContainer.Resources.Requests.Memory(),
						MaxMem:   oldContainer.Resources.Limits.Memory(),
					},
					replicationControllerComparator{
						Image:    newContainer.Image,
						Replicas: r.newReplicationControllers[replicationController.GetName()].Spec.Replicas,
						MinCPU:   newContainer.Resources.Requests.Cpu(),
						MaxCPU:   newContainer.Resources.Limits.Cpu(),
						MinMem:   newContainer.Resources.Requests.Memory(),
						MaxMem:   newContainer.Resources.Limits.Memory(),
					}) {
				isChanged = true
			}
		}
	}

	// if there is any change from the above step, patch the replication controller
	if isChanged {
		log.Infof("updating the replication controller %s in %s namespace", replicationController.GetName(), r.config.namespace)
		err := util.PatchReplicationController(r.config.kubeClient, r.oldReplicationControllers[replicationController.GetName()], *r.newReplicationControllers[replicationController.GetName()])
		if err != nil {
			return false, errors.Annotatef(err, "unable to patch rc %s to kube in namespace %s", replicationController.GetName(), r.config.namespace)
		}
	}
	return false, nil
}
