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

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Deployment stores the configuration to add or delete the deployment object
type Deployment struct {
	config         *CommonConfig
	deployer       *util.DeployerHelper
	deployments    []*components.Deployment
	oldDeployments map[string]appsv1.Deployment
	newDeployments map[string]*appsv1.Deployment
}

// NewDeployment returns the deployment
func NewDeployment(config *CommonConfig, deployments []*components.Deployment) (*Deployment, error) {
	deployer, err := util.NewDeployer(config.kubeConfig)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to get deployer object for %s", config.namespace)
	}
	newDeployments := append([]*components.Deployment{}, deployments...)
	for i := 0; i < len(newDeployments); i++ {
		if !isLabelsExist(config.expectedLabels, newDeployments[i].Labels) {
			newDeployments = append(newDeployments[:i], newDeployments[i+1:]...)
			i--
		}
	}
	return &Deployment{
		config:         config,
		deployer:       deployer,
		deployments:    newDeployments,
		oldDeployments: make(map[string]appsv1.Deployment, 0),
		newDeployments: make(map[string]*appsv1.Deployment, 0),
	}, nil
}

// buildNewAndOldObject builds the old and new deployment
func (d *Deployment) buildNewAndOldObject() error {
	// build old deployment
	oldRCs, err := d.list()
	if err != nil {
		return errors.Annotatef(err, "unable to get deployments for %s", d.config.namespace)
	}
	for _, oldRC := range oldRCs.(*appsv1.DeploymentList).Items {
		d.oldDeployments[oldRC.GetName()] = oldRC
	}

	// build new deployment
	for _, newDp := range d.deployments {
		d.newDeployments[newDp.GetName()] = newDp.Deployment
	}

	return nil
}

// add adds the deployment
func (d *Deployment) add(isPatched bool) (bool, error) {
	isAdded := false
	for _, deployment := range d.deployments {
		if _, ok := d.oldDeployments[deployment.GetName()]; !ok {
			d.deployer.Deployer.AddComponent(horizonapi.DeploymentComponent, deployment)
			isAdded = true
		} else {
			_, err := d.patch(deployment, isPatched)
			if err != nil {
				return false, errors.Annotatef(err, "patch deployment")
			}
		}
	}
	if isAdded && !d.config.dryRun {
		err := d.deployer.Deployer.Run()
		if err != nil {
			return false, errors.Annotatef(err, "unable to deploy deployment in %s", d.config.namespace)
		}
	}
	return false, nil
}

// get gets the deployment
func (d *Deployment) get(name string) (interface{}, error) {
	return util.GetDeployment(d.config.kubeClient, d.config.namespace, name)
}

// list lists all the deployments
func (d *Deployment) list() (interface{}, error) {
	return util.ListDeployments(d.config.kubeClient, d.config.namespace, d.config.labelSelector)
}

// delete deletes the deployment
func (d *Deployment) delete(name string) error {
	log.Infof("deleting the deployment %s in %s namespace", name, d.config.namespace)
	return util.DeleteDeployment(d.config.kubeClient, d.config.namespace, name)
}

// remove removes the deployment
func (d *Deployment) remove() error {
	// compare the old and new deployment and delete if needed
	for _, oldDeployment := range d.oldDeployments {
		if _, ok := d.newDeployments[oldDeployment.GetName()]; !ok {
			err := d.delete(oldDeployment.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete deployment %s in namespace %s", oldDeployment.GetName(), d.config.namespace)
			}
		}
	}
	return nil
}

// deploymentComparator used to compare deployment attributes
type deploymentComparator struct {
	Image          string
	Replicas       *int32
	MinCPU         *resource.Quantity
	MaxCPU         *resource.Quantity
	MinMem         *resource.Quantity
	MaxMem         *resource.Quantity
	EnvFrom        []corev1.EnvFromSource
	ServiceAccount string
}

// patch patches the deployment
func (d *Deployment) patch(rc interface{}, isPatched bool) (bool, error) {
	deployment := rc.(*components.Deployment)
	// check isPatched, why?
	// if there is any configuration change, irrespective of comparing any changes, patch the deployment
	if isPatched && !d.config.dryRun {
		log.Infof("updating the deployment %s in %s namespace", deployment.GetName(), d.config.namespace)
		err := util.PatchDeployment(d.config.kubeClient, d.oldDeployments[deployment.GetName()], *d.newDeployments[deployment.GetName()])
		if err != nil {
			return false, errors.Annotatef(err, "unable to patch deployment %s in namespace %s", deployment.GetName(), d.config.namespace)
		}
		return false, nil
	}

	// check whether the deployment or its container got changed
	isChanged := false
	for _, oldContainer := range d.oldDeployments[deployment.GetName()].Spec.Template.Spec.Containers {
		for _, newContainer := range d.newDeployments[deployment.GetName()].Spec.Template.Spec.Containers {
			if strings.EqualFold(oldContainer.Name, newContainer.Name) && !d.config.dryRun &&
				(!reflect.DeepEqual(
					deploymentComparator{
						Image:          oldContainer.Image,
						Replicas:       d.oldDeployments[deployment.GetName()].Spec.Replicas,
						MinCPU:         oldContainer.Resources.Requests.Cpu(),
						MaxCPU:         oldContainer.Resources.Limits.Cpu(),
						MinMem:         oldContainer.Resources.Requests.Memory(),
						MaxMem:         oldContainer.Resources.Limits.Memory(),
						EnvFrom:        oldContainer.EnvFrom,
						ServiceAccount: d.oldDeployments[deployment.GetName()].Spec.Template.Spec.ServiceAccountName,
					},
					deploymentComparator{
						Image:          newContainer.Image,
						Replicas:       d.newDeployments[deployment.GetName()].Spec.Replicas,
						MinCPU:         newContainer.Resources.Requests.Cpu(),
						MaxCPU:         newContainer.Resources.Limits.Cpu(),
						MinMem:         newContainer.Resources.Requests.Memory(),
						MaxMem:         newContainer.Resources.Limits.Memory(),
						EnvFrom:        newContainer.EnvFrom,
						ServiceAccount: d.newDeployments[deployment.GetName()].Spec.Template.Spec.ServiceAccountName,
					}) ||
					!reflect.DeepEqual(sortEnvs(oldContainer.Env), sortEnvs(newContainer.Env)) ||
					!reflect.DeepEqual(sortVolumeMounts(oldContainer.VolumeMounts), sortVolumeMounts(newContainer.VolumeMounts)) ||
					!compareVolumes(sortVolumes(d.oldDeployments[deployment.GetName()].Spec.Template.Spec.Volumes), sortVolumes(d.newDeployments[deployment.GetName()].Spec.Template.Spec.Volumes)) ||
					!compareAffinities(d.oldDeployments[deployment.GetName()].Spec.Template.Spec.Affinity, d.newDeployments[deployment.GetName()].Spec.Template.Spec.Affinity) ||
					!compareProbes(oldContainer.LivenessProbe, newContainer.LivenessProbe) ||
					!compareProbes(oldContainer.ReadinessProbe, newContainer.ReadinessProbe)) {
				isChanged = true
				break
			}
		}
		if isChanged {
			break
		}
	}

	// if there is any change from the above step, patch the deployment
	if isChanged {
		log.Infof("updating the deployment %s in %s namespace", deployment.GetName(), d.config.namespace)
		err := util.PatchDeployment(d.config.kubeClient, d.oldDeployments[deployment.GetName()], *d.newDeployments[deployment.GetName()])
		if err != nil {
			return false, errors.Annotatef(err, "unable to patch rc %s to kube in namespace %s", deployment.GetName(), d.config.namespace)
		}
	}
	return false, nil
}
