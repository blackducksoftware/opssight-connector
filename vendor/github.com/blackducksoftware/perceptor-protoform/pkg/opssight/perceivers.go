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
	"encoding/json"
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/juju/errors"
)

func (p *SpecConfig) perceiverReplicationController(imageName string, replicas int32) *components.ReplicationController {
	rc := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      imageName,
		Namespace: p.config.Namespace,
	})
	rc.AddLabelSelectors(map[string]string{"name": imageName})

	return rc
}

// PodPerceiverReplicationController creates a replication controller for the pod perceiver
func (p *SpecConfig) PodPerceiverReplicationController() (*components.ReplicationController, error) {
	name := p.config.PodPerceiverImageName
	rc := p.perceiverReplicationController(name, 1)

	pod, err := p.perceiverPod(name, p.config.ServiceAccounts["pod-perceiver"], fmt.Sprintf("./%s", name))
	if err != nil {
		return nil, fmt.Errorf("failed to create pod perceiver pod: %v", err)
	}
	rc.AddPod(pod)

	return rc, nil
}

// ImagePerceiverReplicationController creates a replication controller for the image perceiver
func (p *SpecConfig) ImagePerceiverReplicationController() (*components.ReplicationController, error) {
	name := p.config.ImagePerceiverImageName
	rc := p.perceiverReplicationController(name, 1)

	pod, err := p.perceiverPod(name, p.config.ServiceAccounts["image-perceiver"], fmt.Sprintf("./%s", name))
	if err != nil {
		return nil, fmt.Errorf("failed to create image perceiver pod: %v", err)
	}
	rc.AddPod(pod)

	return rc, nil
}

func (p *SpecConfig) perceiverPod(imageName string, account string, cmd string) (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name:           imageName,
		ServiceAccount: account,
	})

	pod.AddLabels(map[string]string{"name": imageName})
	pod.AddContainer(p.perceiverContainer(imageName, cmd))

	vols, err := p.perceiverVolumes()

	if err != nil {
		return nil, err
	}

	for _, v := range vols {
		pod.AddVolume(v)
	}

	return pod, nil
}

func (p *SpecConfig) perceiverContainer(imageName string, cmd string) *components.Container {
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:    imageName,
		Image:   fmt.Sprintf("%s/%s/%s:%s", p.config.Registry, p.config.ImagePath, imageName, p.config.PerceiverImageVersion),
		Command: []string{cmd},
		Args:    []string{"/etc/perceiver/perceiver.yaml"},
		MinCPU:  p.config.DefaultCPU,
		MinMem:  p.config.DefaultMem,
	})

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: fmt.Sprintf("%d", *p.config.PerceiverPort),
		Protocol:      horizonapi.ProtocolTCP,
	})

	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "perceiver",
		MountPath: "/etc/perceiver",
	})
	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "logs",
		MountPath: "/tmp",
	})

	return container
}

func (p *SpecConfig) perceiverVolumes() ([]*components.Volume, error) {
	vols := []*components.Volume{}

	vols = append(vols, components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "perceiver",
		MapOrSecretName: "perceiver",
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

func (p *SpecConfig) perceiverService(imageName string) *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      imageName,
		Namespace: p.config.Namespace,
	})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(*p.config.PerceiverPort),
		TargetPort: fmt.Sprintf("%d", *p.config.PerceiverPort),
		Protocol:   horizonapi.ProtocolTCP,
	})

	service.AddSelectors(map[string]string{"name": imageName})

	return service
}

// PodPerceiverService creates a service for the pod perceiver
func (p *SpecConfig) PodPerceiverService() *components.Service {
	return p.perceiverService(p.config.PodPerceiverImageName)
}

// ImagePerceiverService creates a service for the image perceiver
func (p *SpecConfig) ImagePerceiverService() *components.Service {
	return p.perceiverService(p.config.ImagePerceiverImageName)
}

// PerceiverConfigMap creates a config map for perceivers
func (p *SpecConfig) PerceiverConfigMap() (*components.ConfigMap, error) {
	configMap := components.NewConfigMap(horizonapi.ConfigMapConfig{
		Name:      "perceiver",
		Namespace: p.config.Namespace,
	})

	data := map[string]interface{}{
		"PerceptorHost":             p.config.PerceptorImageName,
		"PerceptorPort":             *p.config.PerceptorPort,
		"AnnotationIntervalSeconds": *p.config.AnnotationIntervalSeconds,
		"DumpIntervalMinutes":       *p.config.DumpIntervalMinutes,
		"Port":                      *p.config.PerceiverPort,
		"LogLevel":                  p.config.LogLevel,
		"RequireLabel":              *p.config.RequireLabel,
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Trace(err)
	}
	configMap.AddData(map[string]string{"perceiver.yaml": string(bytes)})

	return configMap, nil
}

func (p *SpecConfig) perceiverServiceAccount(name string) *components.ServiceAccount {
	serviceAccount := components.NewServiceAccount(horizonapi.ServiceAccountConfig{
		Name:      name,
		Namespace: p.config.Namespace,
	})

	return serviceAccount
}

// PodPerceiverServiceAccount creates a service account for the pod perceiver
func (p *SpecConfig) PodPerceiverServiceAccount() *components.ServiceAccount {
	return p.perceiverServiceAccount(p.config.ServiceAccounts["pod-perceiver"])
}

// ImagePerceiverServiceAccount creates a service account for the image perceiver
func (p *SpecConfig) ImagePerceiverServiceAccount() *components.ServiceAccount {
	return p.perceiverServiceAccount(p.config.ServiceAccounts["image-perceiver"])
}

// PodPerceiverClusterRole creates a cluster role for the pod perceiver
func (p *SpecConfig) PodPerceiverClusterRole() *components.ClusterRole {
	clusterRole := components.NewClusterRole(horizonapi.ClusterRoleConfig{
		Name:       "pod-perceiver",
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	clusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		APIGroups: []string{"*"},
		Resources: []string{"pods"},
		Verbs:     []string{"get", "watch", "list", "update"},
	})

	return clusterRole
}

// ImagePerceiverClusterRole creates a cluster role for the image perceiver
func (p *SpecConfig) ImagePerceiverClusterRole() *components.ClusterRole {
	clusterRole := components.NewClusterRole(horizonapi.ClusterRoleConfig{
		Name:       "image-perceiver",
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	clusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		APIGroups: []string{"*"},
		Resources: []string{"images"},
		Verbs:     []string{"get", "watch", "list", "update"},
	})

	return clusterRole
}

// PodPerceiverClusterRoleBinding creates a cluster role binding for the pod perceiver
func (p *SpecConfig) PodPerceiverClusterRoleBinding(clusterRole *components.ClusterRole) *components.ClusterRoleBinding {
	clusterRoleBinding := components.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		Name:       "pod-perceiver",
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	clusterRoleBinding.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      p.config.ServiceAccounts["pod-perceiver"],
		Namespace: p.config.Namespace,
	})
	clusterRoleBinding.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     clusterRole.GetName(),
	})

	return clusterRoleBinding
}

// ImagePerceiverClusterRoleBinding creates a cluster role binding for the image perceiver
func (p *SpecConfig) ImagePerceiverClusterRoleBinding(clusterRole *components.ClusterRole) *components.ClusterRoleBinding {
	clusterRoleBinding := components.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		Name:       "image-perceiver",
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	clusterRoleBinding.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      p.config.ServiceAccounts["image-perceiver"],
		Namespace: p.config.Namespace,
	})
	clusterRoleBinding.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     clusterRole.GetName(),
	})

	return clusterRoleBinding
}
