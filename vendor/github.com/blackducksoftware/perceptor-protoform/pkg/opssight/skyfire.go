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

package opssight

import (
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
)

// PerceptorSkyfireReplicationController creates a replication controller for perceptor skyfire
func (p *SpecConfig) PerceptorSkyfireReplicationController() (*components.ReplicationController, error) {
	replicas := int32(1)
	rc := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      p.config.SkyfireImageName,
		Namespace: p.config.Namespace,
	})
	rc.AddLabelSelectors(map[string]string{"name": p.config.SkyfireImageName})
	pod, err := p.perceptorSkyfirePod()
	if err != nil {
		return nil, fmt.Errorf("failed to create skyfire volumes: %v", err)
	}
	rc.AddPod(pod)

	return rc, nil
}

func (p *SpecConfig) perceptorSkyfirePod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name:           p.config.SkyfireImageName,
		ServiceAccount: p.config.ServiceAccounts["skyfire"],
	})
	pod.AddLabels(map[string]string{"name": p.config.SkyfireImageName})

	cont, err := p.perceptorSkyfireContainer()
	if err != nil {
		return nil, err
	}
	err = pod.AddContainer(cont)
	if err != nil {
		return nil, fmt.Errorf("unable to add skyfire container: %v", err)
	}

	vols, err := p.perceptorSkyfireVolumes()
	if err != nil {
		return nil, fmt.Errorf("error creating skyfire volumes: %v", err)
	}
	for _, v := range vols {
		err = pod.AddVolume(v)
		if err != nil {
			return nil, fmt.Errorf("error add pod volume: %v", err)
		}
	}

	return pod, nil
}

func (p *SpecConfig) perceptorSkyfireContainer() (*components.Container, error) {
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:    p.config.SkyfireImageName,
		Image:   fmt.Sprintf("%s/%s/%s:%s", p.config.Registry, p.config.ImagePath, p.config.SkyfireImageName, p.config.SkyfireImageVersion),
		Command: []string{fmt.Sprintf("./%s", p.config.SkyfireImageName)},
		Args:    []string{"/etc/skyfire/skyfire.yaml"},
		MinCPU:  p.config.DefaultCPU,
		MinMem:  p.config.DefaultMem,
	})

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: fmt.Sprintf("%d", *p.config.SkyfirePort),
		Protocol:      horizonapi.ProtocolTCP,
	})

	err := container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "skyfire",
		MountPath: "/etc/skyfire",
	})
	if err != nil {
		return nil, err
	}
	err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "logs",
		MountPath: "/tmp",
	})
	if err != nil {
		return nil, err
	}

	err = container.AddEnv(horizonapi.EnvConfig{
		NameOrPrefix: p.config.HubUserPasswordEnvVar,
		Type:         horizonapi.EnvFromSecret,
		KeyOrVal:     "HubUserPassword",
		FromName:     p.config.SecretName,
	})
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (p *SpecConfig) perceptorSkyfireVolumes() ([]*components.Volume, error) {
	vols := []*components.Volume{}

	vols = append(vols, components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "skyfire",
		MapOrSecretName: "skyfire",
	}))

	vol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "logs",
		Medium:     horizonapi.StorageMediumDefault,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create empty dir volume: %v", err)
	}
	vols = append(vols, vol)

	return vols, nil
}

// PerceptorSkyfireService creates a service for perceptor skyfire
func (p *SpecConfig) PerceptorSkyfireService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      p.config.SkyfireImageName,
		Namespace: p.config.Namespace,
	})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(*p.config.SkyfirePort),
		TargetPort: fmt.Sprintf("%d", *p.config.SkyfirePort),
		Protocol:   horizonapi.ProtocolTCP,
	})

	service.AddSelectors(map[string]string{"name": p.config.SkyfireImageName})

	return service
}

// PerceptorSkyfireConfigMap creates a config map for perceptor skyfire
func (p *SpecConfig) PerceptorSkyfireConfigMap() *components.ConfigMap {
	configMap := components.NewConfigMap(horizonapi.ConfigMapConfig{
		Name:      "skyfire",
		Namespace: p.config.Namespace,
	})
	configMap.AddData(map[string]string{"skyfire.yaml": fmt.Sprint(`{"UseInClusterConfig": "`, "true", `","Port": "`, *p.config.SkyfirePort, `","HubHost": "`, p.config.HubHost, `","HubPort": "`, *p.config.HubPort, `","HubUser": "`, p.config.HubUser, `","HubUserPasswordEnvVar": "`, p.config.HubUserPasswordEnvVar, `","HubClientTimeoutSeconds": "`, *p.config.HubClientTimeoutScannerSeconds, `","PerceptorHost": "`, p.config.PerceptorImageName, `","PerceptorPort": "`, *p.config.PerceptorPort, `","KubeDumpIntervalSeconds": "`, "15", `","PerceptorDumpIntervalSeconds": "`, "15", `","HubDumpPauseSeconds": "`, "30", `","ImageFacadePort": "`, *p.config.ImageFacadePort, `","LogLevel": "`, p.config.LogLevel, `"}`)})

	return configMap
}

// PerceptorSkyfireServiceAccount creates a service account for perceptor skyfire
func (p *SpecConfig) PerceptorSkyfireServiceAccount() *components.ServiceAccount {
	serviceAccount := components.NewServiceAccount(horizonapi.ServiceAccountConfig{
		Name:      "skyfire",
		Namespace: p.config.Namespace,
	})

	return serviceAccount
}

// PerceptorSkyfireClusterRole creates a cluster role for perceptor skyfire
func (p *SpecConfig) PerceptorSkyfireClusterRole() *components.ClusterRole {
	clusterRole := components.NewClusterRole(horizonapi.ClusterRoleConfig{
		Name:       "skyfire",
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	clusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		APIGroups: []string{"*"},
		Resources: []string{"pods", "nodes"},
		Verbs:     []string{"get", "watch", "list"},
	})

	return clusterRole
}

// PerceptorSkyfireClusterRoleBinding creates a cluster role binding for perceptor skyfire
func (p *SpecConfig) PerceptorSkyfireClusterRoleBinding(clusterRole *components.ClusterRole) *components.ClusterRoleBinding {
	clusterRoleBinding := components.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		Name:       "skyfire",
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	clusterRoleBinding.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      "skyfire",
		Namespace: p.config.Namespace,
	})
	clusterRoleBinding.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     clusterRole.GetName(),
	})

	return clusterRoleBinding
}
