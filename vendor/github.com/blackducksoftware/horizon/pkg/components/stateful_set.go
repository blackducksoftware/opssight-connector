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
KIND, either express or implies. See the License for the
specific language governing permissions and limitations
under the License.
*/

package components

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/util"

	"github.com/koki/short/converter/converters"
	"github.com/koki/short/types"

	"k8s.io/apimachinery/pkg/runtime"
)

// StatefulSet defines the stateful set component
type StatefulSet struct {
	obj *types.StatefulSet
}

// NewStatefulSet creates a StatefulSet object
func NewStatefulSet(config api.StatefulSetConfig) *StatefulSet {
	s := &types.StatefulSet{
		Version:              config.APIVersion,
		Cluster:              config.ClusterName,
		Name:                 config.Name,
		Namespace:            config.Namespace,
		Replicas:             config.Replicas,
		Partition:            config.Partition,
		RevisionHistoryLimit: config.RevisionHistoryLimit,
		Service:              config.Service,
	}

	if config.UpdateStrategy == api.StatefulSetUpdateStrategyOnDelete {
		s.OnDelete = true
	} else {
		s.OnDelete = false
	}

	switch config.PodManagementPolicy {
	case api.PodManagementPolicyOrdered:
		s.PodManagementPolicy = types.OrderedReadyPodManagement
	case api.PodManagementPolicyParallel:
		s.PodManagementPolicy = types.ParallelPodManagement
	}

	return &StatefulSet{obj: s}
}

// GetObj returns the stateful set object in a format the deployer can use
func (s *StatefulSet) GetObj() *types.StatefulSet {
	return s.obj
}

// GetName returns the name of the stateful set
func (s *StatefulSet) GetName() string {
	return s.obj.Name
}

// AddAnnotations adds annotations to the stateful set
func (s *StatefulSet) AddAnnotations(new map[string]string) {
	s.obj.Annotations = util.MapMerge(s.obj.Annotations, new)
}

// RemoveAnnotations removes annotations from the stateful set
func (s *StatefulSet) RemoveAnnotations(remove []string) {
	for _, k := range remove {
		s.obj.Annotations = util.RemoveElement(s.obj.Annotations, k)
	}
}

// AddLabels adds labels to the stateful set
func (s *StatefulSet) AddLabels(new map[string]string) {
	s.obj.Labels = util.MapMerge(s.obj.Labels, new)
}

// RemoveLabels removes labels from the stateful set
func (s *StatefulSet) RemoveLabels(remove []string) {
	for _, k := range remove {
		s.obj.Labels = util.RemoveElement(s.obj.Labels, k)
	}
}

// AddPod adds a pod to the stateful set
func (s *StatefulSet) AddPod(obj *Pod) error {
	o := obj.GetObj()
	s.obj.TemplateMetadata = &o.PodTemplateMeta
	s.obj.PodTemplate = o.PodTemplate

	return nil
}

// RemovePod removes a pod from the stateful set
func (s *StatefulSet) RemovePod(name string) error {
	if strings.Compare(s.obj.TemplateMetadata.Name, name) != 0 {
		return fmt.Errorf("pod with name %s doesn't exist on stateful set", name)
	}
	s.obj.TemplateMetadata = nil
	s.obj.PodTemplate = types.PodTemplate{}
	return nil
}

// AddMatchLabelsSelectors adds the given match label selectors to the stateful set
func (s *StatefulSet) AddMatchLabelsSelectors(new map[string]string) {
	if s.obj.Selector == nil {
		s.obj.Selector = &types.RSSelector{}
	}
	s.obj.Selector.Labels = util.MapMerge(s.obj.Selector.Labels, new)
}

// RemoveMatchLabelsSelectors removes the given match label selectors from the stateful set
func (s *StatefulSet) RemoveMatchLabelsSelectors(remove []string) {
	for _, k := range remove {
		s.obj.Selector.Labels = util.RemoveElement(s.obj.Selector.Labels, k)
	}
}

// AddMatchExpressionsSelector will add match expressions selectors to the stateful set.
// It takes a string in the following form:
// key <op> <value>
// Where op can be:
// = 	Equal to value ot should be one of the comma separated values
// !=	Key should not be one of the comma separated values
// If no op is provided, then the key should (or should not) exist
// <key>	key should exist
// !<key>	key should not exist
func (s *StatefulSet) AddMatchExpressionsSelector(add string) {
	s.obj.Selector.Shorthand = add
}

// RemoveMatchExpressionsSelector removes the match expressions selector from the stateful set
func (s *StatefulSet) RemoveMatchExpressionsSelector() {
	s.obj.Selector.Shorthand = ""
}

// AddVolumeClaimTemplate adds a volume claim template to the stateful set
func (s *StatefulSet) AddVolumeClaimTemplate(claim PersistentVolumeClaim) {
	pvc := claim.GetObj()
	s.obj.PVCs = append(s.obj.PVCs, *pvc)
}

// RemoveVolumeClaimTemplate removes a volume claim template from the stateful set
func (s *StatefulSet) RemoveVolumeClaimTemplate(claim PersistentVolumeClaim) {
	pvc := claim.GetObj()
	for l, p := range s.obj.PVCs {
		if reflect.DeepEqual(p, *pvc) {
			s.obj.PVCs = append(s.obj.PVCs[:l], s.obj.PVCs[l+1:]...)
			break
		}
	}
}

// ToKube returns the kubernetes version of the stateful set
func (s *StatefulSet) ToKube() (runtime.Object, error) {
	wrapper := &types.StatefulSetWrapper{StatefulSet: *s.obj}
	return converters.Convert_Koki_StatefulSet_to_Kube_apps_v1beta2_StatefulSet(wrapper)
}
