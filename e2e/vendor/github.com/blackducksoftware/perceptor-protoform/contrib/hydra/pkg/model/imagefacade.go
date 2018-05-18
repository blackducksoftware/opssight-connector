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

type ImagefacadeConfigMap struct {
	DockerUser               string
	DockerPassword           string
	InternalDockerRegistries []string
	LogLevel                 string
	CreateImagesOnly         bool
	Port                     int32
}

func NewImagefacadeConfigMap(dockerUser string, dockerPassword string, internalDockerRegistries []string, logLevel string, createImagesOnly bool, port int32) *ImagefacadeConfigMap {
	return &ImagefacadeConfigMap{
		DockerUser:               dockerUser,
		DockerPassword:           dockerPassword,
		InternalDockerRegistries: internalDockerRegistries,
		LogLevel:                 logLevel,
		CreateImagesOnly:         createImagesOnly,
		Port:                     port,
	}
}

type Imagefacade struct {
	Image  string
	CPU    resource.Quantity
	Memory resource.Quantity

	ConfigMapName  string
	ConfigMapMount string
	ConfigMapPath  string
	Config         ImagefacadeConfigMap

	ServiceAccountName string
	ServiceName        string

	DockerSocketName string
	DockerSocketPath string

	PodName string

	ImagesMountName string
	ImagesMountPath string
}

func NewImagefacade(serviceAccountName string) *Imagefacade {
	defaultMem, err := resource.ParseQuantity("512Mi")
	if err != nil {
		panic(err)
	}
	defaultCPU, err := resource.ParseQuantity("100m")
	if err != nil {
		panic(err)
	}
	return &Imagefacade{
		Image:              "gcr.io/gke-verification/blackducksoftware/perceptor-imagefacade:master",
		CPU:                defaultCPU,
		Memory:             defaultMem,
		ConfigMapName:      "perceptor-imagefacade-config",
		ConfigMapMount:     "/etc/perceptor_imagefacade",
		ConfigMapPath:      "perceptor_imagefacade_conf.yaml",
		ServiceAccountName: serviceAccountName,
		ServiceName:        "perceptor-imagefacade",

		DockerSocketName: "dir-docker-socket",
		DockerSocketPath: "/var/run/docker.sock",

		// Must fill these out before using this object
		PodName: "",

		ImagesMountName: "var-images",
		ImagesMountPath: "/var/images",
	}
}

func (pif *Imagefacade) Container() *v1.Container {
	privileged := true
	return &v1.Container{
		Name:            "perceptor-imagefacade",
		Image:           pif.Image,
		ImagePullPolicy: "Always",
		Command:         []string{},
		Ports: []v1.ContainerPort{
			{
				ContainerPort: pif.Config.Port,
				Protocol:      "TCP",
			},
		},
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    pif.CPU,
				v1.ResourceMemory: pif.Memory,
			},
		},
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      pif.ImagesMountName,
				MountPath: pif.ImagesMountPath,
			},
			{
				Name:      pif.ConfigMapName,
				MountPath: pif.ConfigMapMount,
			},
			{
				Name:      pif.DockerSocketName,
				MountPath: pif.DockerSocketPath,
			},
		},
		SecurityContext: &v1.SecurityContext{Privileged: &privileged},
	}
}

func (pif *Imagefacade) Service() *v1.Service {
	return &v1.Service{
		ObjectMeta: v1meta.ObjectMeta{
			Name: pif.ServiceName,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: pif.ServiceName,
					Port: pif.Config.Port,
				},
			},
			Selector: map[string]string{"name": pif.PodName}}}
}

func (pif *Imagefacade) ReplicationController() *v1.ReplicationController {
	replicaCount := int32(1)
	return &v1.ReplicationController{
		ObjectMeta: v1meta.ObjectMeta{Name: pif.PodName},
		Spec: v1.ReplicationControllerSpec{
			Replicas: &replicaCount,
			Selector: map[string]string{"name": pif.PodName},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: v1meta.ObjectMeta{Labels: map[string]string{"name": pif.PodName}},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						{
							Name: pif.ConfigMapName,
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{Name: pif.ConfigMapName},
								},
							},
						},
						{
							Name:         pif.ImagesMountName,
							VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{}},
						},
						{
							Name: pif.DockerSocketName,
							VolumeSource: v1.VolumeSource{
								HostPath: &v1.HostPathVolumeSource{Path: pif.DockerSocketPath},
							},
						},
					},
					Containers:         []v1.Container{*pif.Container()},
					ServiceAccountName: pif.ServiceAccountName,
				}}}}
}

func (pif *Imagefacade) ConfigMap() *v1.ConfigMap {
	jsonBytes, err := json.Marshal(pif.Config)
	if err != nil {
		panic(err)
	}
	return MakeConfigMap(pif.ConfigMapName, pif.ConfigMapPath, string(jsonBytes))
}
