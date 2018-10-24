/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownershia. The ASF licenses this file
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

package alert

import (
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
)

// cfsslDeployment creates a new deployment for cfssl
func (a *SpecConfig) cfsslDeployment() (*components.Deployment, error) {
	replicas := int32(1)
	deployment := components.NewDeployment(horizonapi.DeploymentConfig{
		Replicas:  &replicas,
		Name:      "cfssl",
		Namespace: a.config.Namespace,
	})
	deployment.AddMatchLabelsSelectors(map[string]string{"app": "cfssl", "tier": "cfssl"})

	pod, err := a.cfsslPod()
	if err != nil {
		return nil, fmt.Errorf("failed to create pod: %v", err)
	}
	deployment.AddPod(pod)

	return deployment, nil
}

func (a *SpecConfig) cfsslPod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name: "cfssl",
	})
	pod.AddLabels(map[string]string{"app": "cfssl", "tier": "cfssl"})

	pod.AddContainer(a.cfsslContainer())

	vol, err := a.cfsslVolume()
	if err != nil {
		return nil, fmt.Errorf("error creating volumes: %v", err)
	}
	pod.AddVolume(vol)

	return pod, nil
}

func (a *SpecConfig) cfsslContainer() *components.Container {
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:   "hub-cfssl",
		Image:  fmt.Sprintf("%s/%s/%s:%s", a.config.Registry, a.config.ImagePath, a.config.CfsslImageName, a.config.CfsslImageVersion),
		MinMem: a.config.CfsslMemory,
	})

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: "8888",
		Protocol:      horizonapi.ProtocolTCP,
	})

	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "dir-cfssl",
		MountPath: "/etc/cfssl",
	})

	container.AddEnv(horizonapi.EnvConfig{
		Type:     horizonapi.EnvFromConfigMap,
		FromName: "alert",
	})

	container.AddLivenessProbe(horizonapi.ProbeConfig{
		ActionConfig: horizonapi.ActionConfig{
			Command: []string{"/usr/local/bin/docker-healthcheck.sh", "http://localhost:8888/api/v1/cfssl/scaninfo"},
		},
		Delay:           240,
		Timeout:         10,
		Interval:        30,
		MinCountFailure: 10,
	})

	return container
}

func (a *SpecConfig) cfsslVolume() (*components.Volume, error) {
	vol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "dir-cfssl",
		Medium:     horizonapi.StorageMediumDefault,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create empty dir volume: %v", err)
	}

	return vol, nil
}

// CfsslService creates a service for cfssl
func (a *SpecConfig) cfsslService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:          "cfssl",
		Namespace:     a.config.Namespace,
		IPServiceType: horizonapi.ClusterIPServiceTypeNodePort,
	})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       8888,
		TargetPort: "8888",
		Protocol:   horizonapi.ProtocolTCP,
		Name:       "8888-tcp",
	})

	service.AddSelectors(map[string]string{"app": "cfssl"})

	return service
}
