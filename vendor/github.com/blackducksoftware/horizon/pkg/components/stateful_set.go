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
KIND, either express or implies. See the License for the
specific language governing permissions and limitations
under the License.
*/

package components

import (
	"reflect"

	"github.com/blackducksoftware/horizon/pkg/api"

	"k8s.io/api/apps/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StatefulSet defines the stateful set component
type StatefulSet struct {
	*v1.StatefulSet
	MetadataFuncs
	LabelSelectorFuncs
	PodFuncs
}

// NewStatefulSet creates a StatefulSet object
func NewStatefulSet(config api.StatefulSetConfig) *StatefulSet {
	version := "apps/v1"
	if len(config.APIVersion) > 0 {
		version = config.APIVersion
	}

	s := v1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: version,
		},
		ObjectMeta: generateObjectMeta(config.Name, config.Namespace, config.ClusterName),
		Spec: v1.StatefulSetSpec{
			Replicas:             config.Replicas,
			ServiceName:          config.Service,
			RevisionHistoryLimit: config.RevisionHistoryLimit,
		},
	}

	switch config.UpdateStrategy {
	case api.StatefulSetUpdateStrategyOnDelete:
		s.Spec.UpdateStrategy.Type = v1.OnDeleteStatefulSetStrategyType
	case api.StatefulSetUpdateStrategyRollingUpdate:
		s.Spec.UpdateStrategy.Type = v1.RollingUpdateStatefulSetStrategyType
		if config.Partition != nil {
			s.Spec.UpdateStrategy.RollingUpdate = &v1.RollingUpdateStatefulSetStrategy{
				Partition: config.Partition,
			}
		}
	}

	switch config.PodManagementPolicy {
	case api.PodManagementPolicyOrdered:
		s.Spec.PodManagementPolicy = v1.OrderedReadyPodManagement
	case api.PodManagementPolicyParallel:
		s.Spec.PodManagementPolicy = v1.ParallelPodManagement
	}

	return &StatefulSet{&s, MetadataFuncs{&s}, LabelSelectorFuncs{&s}, PodFuncs{&s}}
}

// AddVolumeClaimTemplate adds a volume claim template to the stateful set
func (s *StatefulSet) AddVolumeClaimTemplate(claim PersistentVolumeClaim) {
	s.Spec.VolumeClaimTemplates = append(s.Spec.VolumeClaimTemplates, *claim.PersistentVolumeClaim)
}

// RemoveVolumeClaimTemplate removes a volume claim template from the stateful set
func (s *StatefulSet) RemoveVolumeClaimTemplate(claim PersistentVolumeClaim) {
	for l, c := range s.Spec.VolumeClaimTemplates {
		if reflect.DeepEqual(c, claim.PersistentVolumeClaim) {
			s.Spec.VolumeClaimTemplates = append(s.Spec.VolumeClaimTemplates[:l], s.Spec.VolumeClaimTemplates[l+1:]...)
			break
		}
	}
}

// Deploy will deploy the stateful set to the cluster
func (s *StatefulSet) Deploy(res api.DeployerResources) error {
	_, err := res.KubeClient.AppsV1().StatefulSets(s.Namespace).Create(s.StatefulSet)
	return err
}

// Undeploy will remove the stateful set from the cluster
func (s *StatefulSet) Undeploy(res api.DeployerResources) error {
	return res.KubeClient.AppsV1().StatefulSets(s.Namespace).Delete(s.Name, &metav1.DeleteOptions{})
}
