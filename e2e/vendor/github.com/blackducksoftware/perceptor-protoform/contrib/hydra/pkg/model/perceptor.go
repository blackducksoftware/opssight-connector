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

type PerceptorConfigMap struct {
	HubHost               string
	HubUser               string
	HubUserPasswordEnvVar string
	HubPort               int
	ConcurrentScanLimit   int
	UseMockMode           bool
	Port                  int32
	LogLevel              string
}

func NewPerceptorConfigMap(hubHost string, hubUser string, hubUserPasswordEnvVar string, hubPort int, concurrentScanLimit int, useMockMode bool, port int32, logLevel string) *PerceptorConfigMap {
	return &PerceptorConfigMap{
		HubHost:               hubHost,
		HubUser:               hubUser,
		HubUserPasswordEnvVar: hubUserPasswordEnvVar,
		HubPort:               hubPort,
		ConcurrentScanLimit:   concurrentScanLimit,
		UseMockMode:           useMockMode,
		Port:                  port,
		LogLevel:              logLevel,
	}
}

type Perceptor struct {
	PodName string
	Image   string
	CPU     resource.Quantity
	Memory  resource.Quantity

	ConfigMapName  string
	ConfigMapMount string
	ConfigMapPath  string
	Config         PerceptorConfigMap

	HubPasswordSecretName string
	HubPasswordSecretKey  string

	ReplicaCount int32
	ServiceName  string
}

func NewPerceptor() *Perceptor {
	memory, err := resource.ParseQuantity("512Mi")
	if err != nil {
		panic(err)
	}
	cpu, err := resource.ParseQuantity("100m")
	if err != nil {
		panic(err)
	}

	return &Perceptor{
		PodName:        "perceptor",
		Image:          "gcr.io/gke-verification/blackducksoftware/perceptor:master",
		CPU:            cpu,
		Memory:         memory,
		ConfigMapName:  "perceptor-config",
		ConfigMapMount: "/etc/perceptor",
		ConfigMapPath:  "perceptor_conf.yaml",
		ReplicaCount:   1,
		ServiceName:    "perceptor",
	}
}

func (pc *Perceptor) Container() *v1.Container {
	return &v1.Container{
		Name:            "perceptor",
		Image:           pc.Image,
		ImagePullPolicy: "Always",
		Env: []v1.EnvVar{
			{
				Name: pc.Config.HubUserPasswordEnvVar,
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: pc.HubPasswordSecretName,
						},
						Key: pc.HubPasswordSecretKey,
					},
				},
			},
		},
		Command: []string{},
		Ports: []v1.ContainerPort{
			{
				ContainerPort: pc.Config.Port,
				Protocol:      "TCP",
			},
		},
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    pc.CPU,
				v1.ResourceMemory: pc.Memory,
			},
		},
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      pc.ConfigMapName,
				MountPath: pc.ConfigMapMount,
			},
		},
	}
}

func (pc *Perceptor) ReplicationController() *v1.ReplicationController {
	return &v1.ReplicationController{
		ObjectMeta: v1meta.ObjectMeta{Name: pc.PodName},
		Spec: v1.ReplicationControllerSpec{
			Replicas: &pc.ReplicaCount,
			Selector: map[string]string{"name": pc.PodName},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: v1meta.ObjectMeta{Labels: map[string]string{"name": pc.PodName}},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						{
							Name: pc.ConfigMapName,
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{Name: pc.ConfigMapName},
								},
							},
						},
					},
					Containers: []v1.Container{*pc.Container()},
					// TODO: RestartPolicy?  terminationGracePeriodSeconds? dnsPolicy?
				}}}}
}

func (pc *Perceptor) Service() *v1.Service {
	return &v1.Service{
		ObjectMeta: v1meta.ObjectMeta{
			Name: pc.ServiceName,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: pc.ServiceName,
					Port: pc.Config.Port,
				},
			},
			Selector: map[string]string{"name": pc.ServiceName}}}
}

func (pc *Perceptor) ConfigMap() *v1.ConfigMap {
	jsonBytes, err := json.Marshal(pc.Config)
	if err != nil {
		panic(err)
	}
	return MakeConfigMap(pc.ConfigMapName, pc.ConfigMapPath, string(jsonBytes))
}
