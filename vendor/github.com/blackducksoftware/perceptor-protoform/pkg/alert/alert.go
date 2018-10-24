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

// alertDeployment creates a new deployment for alert
func (a *SpecConfig) alertDeployment() (*components.Deployment, error) {
	replicas := int32(1)
	deployment := components.NewDeployment(horizonapi.DeploymentConfig{
		Replicas:  &replicas,
		Name:      "alert",
		Namespace: a.config.Namespace,
	})
	deployment.AddMatchLabelsSelectors(map[string]string{"app": "alert", "tier": "alert"})

	pod, err := a.alertPod()
	if err != nil {
		return nil, fmt.Errorf("failed to create pod: %v", err)
	}
	deployment.AddPod(pod)

	return deployment, nil
}

func (a *SpecConfig) alertPod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name: "alert",
	})
	pod.AddLabels(map[string]string{"app": "alert", "tier": "alert"})

	pod.AddContainer(a.alertContainer())

	vol, err := a.alertVolume()
	if err != nil {
		return nil, fmt.Errorf("error creating volumes: %v", err)
	}
	pod.AddVolume(vol)

	return pod, nil
}

func (a *SpecConfig) alertContainer() *components.Container {
	// This will prevent it from working on openshift without a privileged service account.  Remove once the
	// chowns are removed in the image
	user := int64(0)
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:   "alert",
		Image:  fmt.Sprintf("%s/%s/%s:%s", a.config.Registry, a.config.ImagePath, a.config.AlertImageName, a.config.AlertImageVersion),
		MinMem: a.config.AlertMemory,
		UID:    &user,
	})

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: "8443",
		Protocol:      horizonapi.ProtocolTCP,
	})

	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "dir-alert",
		MountPath: "/opt/blackduck/alert/alert-config",
	})

	container.AddEnv(horizonapi.EnvConfig{
		Type:     horizonapi.EnvFromConfigMap,
		FromName: "alert",
	})

	container.AddLivenessProbe(horizonapi.ProbeConfig{
		ActionConfig: horizonapi.ActionConfig{
			Command: []string{"/usr/local/bin/docker-healthcheck.sh", "https://localhost:8443/alert/api/about"},
		},
		Delay:           240,
		Timeout:         10,
		Interval:        30,
		MinCountFailure: 5,
	})

	return container
}

func (a *SpecConfig) alertVolume() (*components.Volume, error) {
	vol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "dir-alert",
		Medium:     horizonapi.StorageMediumDefault,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create empty dir volume: %v", err)
	}

	return vol, nil
}

// alertService creates a service for alert
func (a *SpecConfig) alertService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:          "alert",
		Namespace:     a.config.Namespace,
		IPServiceType: horizonapi.ClusterIPServiceTypeNodePort,
	})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       8443,
		TargetPort: "8443",
		Protocol:   horizonapi.ProtocolTCP,
		Name:       "8443-tcp",
	})

	service.AddSelectors(map[string]string{"app": "alert"})

	return service
}

// alertExposedService creates a loadBalancer service for alert
func (a *SpecConfig) alertExposedService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:          "alert-lb",
		Namespace:     a.config.Namespace,
		IPServiceType: horizonapi.ClusterIPServiceTypeLoadBalancer,
	})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       8443,
		TargetPort: "8443",
		Protocol:   horizonapi.ProtocolTCP,
		Name:       "8443-tcp",
	})

	service.AddSelectors(map[string]string{"app": "alert"})

	return service
}

// alertConfigMap creates a config map for alert
func (a *SpecConfig) alertConfigMap() *components.ConfigMap {
	configMap := components.NewConfigMap(horizonapi.ConfigMapConfig{
		Name:      "alert",
		Namespace: a.config.Namespace,
	})

	configMap.AddData(map[string]string{
		"ALERT_SERVER_PORT":         fmt.Sprintf("%d", *a.config.Port),
		"PUBLIC_HUB_WEBSERVER_HOST": a.config.HubHost,
		"PUBLIC_HUB_WEBSERVER_PORT": fmt.Sprintf("%d", *a.config.HubPort),
	})

	return configMap
}
