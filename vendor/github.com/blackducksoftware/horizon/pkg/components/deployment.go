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

package components

import (
	"github.com/blackducksoftware/horizon/pkg/api"

	"k8s.io/api/apps/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Deployment defines the deployment component
type Deployment struct {
	*v1.Deployment
	MetadataFuncs
	LabelSelectorFuncs
	PodFuncs
}

// NewDeployment creates a Deployment object
func NewDeployment(config api.DeploymentConfig) *Deployment {
	version := "apps/v1"
	if len(config.APIVersion) > 0 {
		version = config.APIVersion
	}

	d := v1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: version,
		},
		ObjectMeta: generateObjectMeta(config.Name, config.Namespace, config.ClusterName),
		Spec: v1.DeploymentSpec{
			Replicas:                config.Replicas,
			MinReadySeconds:         config.MinReadySeconds,
			RevisionHistoryLimit:    config.RevisionHistoryLimit,
			Paused:                  config.Paused,
			ProgressDeadlineSeconds: config.ProgressDeadlineSeconds,
		},
	}

	switch config.Strategy {
	case api.DeploymentStrategyTypeRollingUpdate:
		d.Spec.Strategy = v1.DeploymentStrategy{
			Type: v1.RollingUpdateDeploymentStrategyType,
			RollingUpdate: &v1.RollingUpdateDeployment{
				MaxUnavailable: createIntOrStr(config.MaxUnavailable),
				MaxSurge:       createIntOrStr(config.MaxExtra),
			},
		}

	case api.DeploymentStrategyTypeRecreate:
		d.Spec.Strategy = v1.DeploymentStrategy{
			Type: v1.RecreateDeploymentStrategyType,
		}
	}

	return &Deployment{&d, MetadataFuncs{&d}, LabelSelectorFuncs{&d}, PodFuncs{&d}}
}

// Deploy will deploy the deployment to the cluster
func (d *Deployment) Deploy(res api.DeployerResources) error {
	_, err := res.KubeClient.AppsV1().Deployments(d.Namespace).Create(d.Deployment)
	return err
}

// Undeploy will remove the deployment from the cluster
func (d *Deployment) Undeploy(res api.DeployerResources) error {
	return res.KubeClient.AppsV1().Deployments(d.Namespace).Delete(d.Name, &metav1.DeleteOptions{})
}
