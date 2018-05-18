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

type MockImagefacadeConfigMap struct {
	Port int32
}

func NewMockImagefacadeConfigMap(port int32) *MockImagefacadeConfigMap {
	return &MockImagefacadeConfigMap{
		Port: port,
	}
}

type MockImagefacade struct {
	Image  string
	CPU    resource.Quantity
	Memory resource.Quantity

	ConfigMapName  string
	ConfigMapMount string
	ConfigMapPath  string
	Config         MockImagefacadeConfigMap

	PodName     string
	ServiceName string

	ImagesMountName string
	ImagesMountPath string
}

func NewMockImagefacade() *MockImagefacade {
	defaultMem, err := resource.ParseQuantity("2Gi")
	if err != nil {
		panic(err)
	}
	defaultCPU, err := resource.ParseQuantity("500m")
	if err != nil {
		panic(err)
	}
	return &MockImagefacade{
		Image:          "gcr.io/gke-verification/blackducksoftware/mockimagefacade:master",
		CPU:            defaultCPU,
		Memory:         defaultMem,
		ConfigMapName:  "mockimagefacade-config",
		ConfigMapMount: "/etc/mockimagefacade",
		ConfigMapPath:  "mockimagefacade_conf.yaml",

		ServiceName: "perceptor-imagefacade",

		// Must fill these out before using this object
		PodName: "",

		ImagesMountName: "var-images",
		ImagesMountPath: "/var/images",
	}
}

func (mif *MockImagefacade) Container() *v1.Container {
	return &v1.Container{
		Name:            "perceptor-imagefacade",
		Image:           mif.Image,
		ImagePullPolicy: "Always",
		Command:         []string{},
		Ports: []v1.ContainerPort{
			{
				ContainerPort: mif.Config.Port,
				Protocol:      "TCP",
			},
		},
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    mif.CPU,
				v1.ResourceMemory: mif.Memory,
			},
		},
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      mif.ImagesMountName,
				MountPath: mif.ImagesMountPath,
			},
			{
				Name:      mif.ConfigMapName,
				MountPath: mif.ConfigMapMount,
			},
		},
	}
}

func (mif *MockImagefacade) Service() *v1.Service {
	return &v1.Service{
		ObjectMeta: v1meta.ObjectMeta{
			Name: mif.ServiceName,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: mif.ServiceName,
					Port: mif.Config.Port,
				},
			},
			Selector: map[string]string{"name": mif.PodName}}}
}

func (mif *MockImagefacade) ConfigMap() *v1.ConfigMap {
	jsonBytes, err := json.Marshal(mif.Config)
	if err != nil {
		panic(err)
	}
	return MakeConfigMap(mif.ConfigMapName, mif.ConfigMapPath, string(jsonBytes))
}
