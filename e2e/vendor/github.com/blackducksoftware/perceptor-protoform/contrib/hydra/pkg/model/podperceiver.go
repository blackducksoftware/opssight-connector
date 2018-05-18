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

package model

import (
	"encoding/json"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodPerceiverConfigMap struct {
	PerceptorHost             string
	PerceptorPort             int32
	AnnotationIntervalSeconds int
	DumpIntervalMinutes       int
	Port                      int32
}

func NewPodPerceiverConfigMap(perceptorHost string, perceptorPort int32, annotationIntervalSeconds int, dumpIntervalMinutes int, port int32) *PodPerceiverConfigMap {
	return &PodPerceiverConfigMap{
		PerceptorHost:             perceptorHost,
		PerceptorPort:             perceptorPort,
		AnnotationIntervalSeconds: annotationIntervalSeconds,
		DumpIntervalMinutes:       dumpIntervalMinutes,
		Port:                      port,
	}
}

type PodPerceiver struct {
	PodName string
	Image   string
	CPU     resource.Quantity
	Memory  resource.Quantity

	ConfigMapName  string
	ConfigMapMount string
	ConfigMapPath  string
	Config         PodPerceiverConfigMap

	ReplicaCount       int32
	ServiceName        string
	ServiceAccountName string
}

func NewPodPerceiver(serviceAccountName string, replicaCount int) *PodPerceiver {
	memory, err := resource.ParseQuantity("512Mi")
	if err != nil {
		panic(err)
	}
	cpu, err := resource.ParseQuantity("100m")
	if err != nil {
		panic(err)
	}

	return &PodPerceiver{
		PodName:            "pod-perceiver",
		Image:              "gcr.io/gke-verification/blackducksoftware/pod-perceiver:master",
		CPU:                cpu,
		Memory:             memory,
		ConfigMapName:      "perceiver",
		ConfigMapMount:     "/etc/perceiver",
		ConfigMapPath:      "perceiver.yaml",
		ReplicaCount:       int32(replicaCount),
		ServiceName:        "pod-perceiver",
		ServiceAccountName: serviceAccountName,
	}
}

func (pp *PodPerceiver) container() *v1.Container {
	return &v1.Container{
		Name:            "pod-perceiver",
		Image:           pp.Image,
		ImagePullPolicy: "Always",
		Command:         []string{},
		Ports: []v1.ContainerPort{
			{
				ContainerPort: pp.Config.Port,
				Protocol:      "TCP",
			},
		},
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    pp.CPU,
				v1.ResourceMemory: pp.Memory,
			},
		},
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      pp.ConfigMapName,
				MountPath: pp.ConfigMapMount,
			},
		},
	}
}

func (pp *PodPerceiver) ReplicationController() *v1.ReplicationController {
	return &v1.ReplicationController{
		ObjectMeta: v1meta.ObjectMeta{Name: pp.PodName},
		Spec: v1.ReplicationControllerSpec{
			Replicas: &pp.ReplicaCount,
			Selector: map[string]string{"name": pp.PodName},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: v1meta.ObjectMeta{Labels: map[string]string{"name": pp.PodName}},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						{
							Name: pp.ConfigMapName,
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{Name: pp.ConfigMapName},
								},
							},
						},
					},
					Containers:         []v1.Container{*pp.container()},
					ServiceAccountName: pp.ServiceAccountName,
					// TODO: RestartPolicy?  terminationGracePeriodSeconds? dnsPolicy?
				}}}}
}

func (pp *PodPerceiver) Service() *v1.Service {
	return &v1.Service{
		ObjectMeta: v1meta.ObjectMeta{
			Name: pp.ServiceName,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: pp.ServiceName,
					Port: pp.Config.Port,
				},
			},
			Selector: map[string]string{"name": pp.PodName}}}
}

func (pp *PodPerceiver) ConfigMap() *v1.ConfigMap {
	jsonBytes, err := json.Marshal(pp.Config)
	if err != nil {
		panic(err)
	}
	return MakeConfigMap(pp.ConfigMapName, pp.ConfigMapPath, string(jsonBytes))
}
