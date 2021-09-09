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

	"k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/imdario/mergo"
)

// ReplicationController defines the replication controller component
type ReplicationController struct {
	*v1.ReplicationController
	MetadataFuncs
	PodFuncs
}

// NewReplicationController creates a ReplicationController object
func NewReplicationController(config api.ReplicationControllerConfig) *ReplicationController {
	version := "v1"
	if len(config.APIVersion) > 0 {
		version = config.APIVersion
	}

	rc := v1.ReplicationController{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: version,
		},
		ObjectMeta: generateObjectMeta(config.Name, config.Namespace, config.ClusterName),
		Spec: v1.ReplicationControllerSpec{
			Replicas:        config.Replicas,
			MinReadySeconds: config.ReadySeconds,
		},
	}

	return &ReplicationController{&rc, MetadataFuncs{&rc}, PodFuncs{&rc}}
}

// AddSelectors adds the given selectors to the replication controller
func (rc *ReplicationController) AddSelectors(new map[string]string) {
	mergo.Merge(&rc.Spec.Selector, new, mergo.WithOverride)
}

// RemoveSelectors removes the given selectors from the replication controller
func (rc *ReplicationController) RemoveSelectors(remove []string) {
	for _, s := range remove {
		delete(rc.Spec.Selector, s)
	}
}

// Deploy will deploy the replication controller to the cluster
func (rc *ReplicationController) Deploy(res api.DeployerResources) error {
	_, err := res.KubeClient.CoreV1().ReplicationControllers(rc.Namespace).Create(rc.ReplicationController)
	return err
}

// Undeploy will remove the replication controller from the cluster
func (rc *ReplicationController) Undeploy(res api.DeployerResources) error {
	return res.KubeClient.CoreV1().ReplicationControllers(rc.Namespace).Delete(rc.Name, &metav1.DeleteOptions{})
}
