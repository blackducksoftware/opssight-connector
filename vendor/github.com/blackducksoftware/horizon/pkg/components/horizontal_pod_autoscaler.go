/*
Copyright (C) 2019 Synopsys, Inc.

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

	"k8s.io/api/autoscaling/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HorizontalPodAutoscaler defines the HorizontalPodAutoscaler component
type HorizontalPodAutoscaler struct {
	*v1.HorizontalPodAutoscaler
	MetadataFuncs
}

// NewHorizontalPodAutoscaler creates a HorizontalPodAutoscaler object
func NewHorizontalPodAutoscaler(config api.HPAConfig) *HorizontalPodAutoscaler {
	version := "autoscaling/v1"
	if len(config.APIVersion) > 0 {
		version = config.APIVersion
	}

	hpa := v1.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HorizontalPodAutoscaler",
			APIVersion: version,
		},
		ObjectMeta: generateObjectMeta(config.Name, config.Namespace, config.ClusterName),
		Spec: v1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: v1.CrossVersionObjectReference{
				Kind:       config.ScaleTargetKind,
				Name:       config.ScaleTargetName,
				APIVersion: config.ScaleTargetAPIVersion,
			},
			MinReplicas:                    config.MinReplicas,
			MaxReplicas:                    config.MaxReplicas,
			TargetCPUUtilizationPercentage: config.TargetCPUUtilizationPercentage,
		},
	}

	return &HorizontalPodAutoscaler{&hpa, MetadataFuncs{&hpa}}
}

// Deploy will deploy the horizontal pod autoscaler to the cluster
func (hpa *HorizontalPodAutoscaler) Deploy(res api.DeployerResources) error {
	_, err := res.KubeClient.AutoscalingV1().HorizontalPodAutoscalers(hpa.Namespace).Create(hpa.HorizontalPodAutoscaler)
	return err
}

// Undeploy will remove the horizontal pod autoscaler from the cluster
func (hpa *HorizontalPodAutoscaler) Undeploy(res api.DeployerResources) error {
	return res.KubeClient.AutoscalingV1().HorizontalPodAutoscalers(hpa.Namespace).Delete(hpa.Name, &metav1.DeleteOptions{})
}
