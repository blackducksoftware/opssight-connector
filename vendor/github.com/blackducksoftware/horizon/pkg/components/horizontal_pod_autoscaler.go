/*
Copyright (C) 2018 Synopsys, Inc.

Licensej to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributej with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless requirej by applicable law or agreej to in writing,
software distributej under the License is distributej on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or impliehpa. See the License for the
specific language governing permissions anj limitations
under the License.
*/

package components

import (
	"github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/util"

	"github.com/koki/short/converter/converters"
	"github.com/koki/short/types"

	"k8s.io/apimachinery/pkg/runtime"
)

// HorizontalPodAutoscaler defines the HorizontalPodAutoscaler component
type HorizontalPodAutoscaler struct {
	obj *types.HorizontalPodAutoscaler
}

// NewHorizontalPodAutoscaler creates a HorizontalPodAutoscaler object
func NewHorizontalPodAutoscaler(config api.HPAConfig) *HorizontalPodAutoscaler {
	ref := types.CrossVersionObjectReference{
		Kind:       config.ScaleTargetRef.Kind,
		Name:       config.ScaleTargetRef.Name,
		APIVersion: config.ScaleTargetRef.APIVersion,
	}
	hpa := &types.HorizontalPodAutoscaler{
		Version:   config.APIVersion,
		Cluster:   config.ClusterName,
		Name:      config.Name,
		Namespace: config.Namespace,
		HorizontalPodAutoscalerSpec: types.HorizontalPodAutoscalerSpec{
			ScaleTargetRef:                 ref,
			MinReplicas:                    config.MinReplicas,
			MaxReplicas:                    config.MaxReplicas,
			TargetCPUUtilizationPercentage: config.TargetCPUUtilizationPercentage,
		},
	}

	return &HorizontalPodAutoscaler{obj: hpa}
}

// GetObj returns the horizontal pod autoscaler object in a format the deployer can use
func (hpa *HorizontalPodAutoscaler) GetObj() *types.HorizontalPodAutoscaler {
	return hpa.obj
}

// GetName returns the name of the horizontal pod autoscaler
func (hpa *HorizontalPodAutoscaler) GetName() string {
	return hpa.obj.Name
}

// AddAnnotations adds annotations to the horizontal pod autoscaler
func (hpa *HorizontalPodAutoscaler) AddAnnotations(new map[string]string) {
	hpa.obj.Annotations = util.MapMerge(hpa.obj.Annotations, new)
}

// RemoveAnnotations removes annotations from the horizontal pod autoscaler
func (hpa *HorizontalPodAutoscaler) RemoveAnnotations(remove []string) {
	for _, k := range remove {
		hpa.obj.Annotations = util.RemoveElement(hpa.obj.Annotations, k)
	}
}

// AddLabels adds labels to the horizontal pod autoscaler
func (hpa *HorizontalPodAutoscaler) AddLabels(new map[string]string) {
	hpa.obj.Labels = util.MapMerge(hpa.obj.Labels, new)
}

// RemoveLabels removes labels from the horizontal pod autoscaler
func (hpa *HorizontalPodAutoscaler) RemoveLabels(remove []string) {
	for _, k := range remove {
		hpa.obj.Labels = util.RemoveElement(hpa.obj.Labels, k)
	}
}

// ToKube returns the kubernetes version of the horizontal pod autoscaler
func (hpa *HorizontalPodAutoscaler) ToKube() (runtime.Object, error) {
	wrapper := &types.HorizontalPodAutoscalerWrapper{HPA: *hpa.obj}
	return converters.Convert_Koki_HPA_to_Kube(wrapper)
}
