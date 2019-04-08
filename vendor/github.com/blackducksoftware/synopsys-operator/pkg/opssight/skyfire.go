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
	"github.com/juju/errors"
)

// PerceptorSkyfireReplicationController creates a replication controller for perceptor skyfire
func (p *SpecConfig) PerceptorSkyfireReplicationController() (*components.ReplicationController, error) {
	replicas := int32(1)
	rc := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      p.opssight.Spec.Skyfire.Name,
		Namespace: p.opssight.Spec.Namespace,
	})
	rc.AddLabelSelectors(map[string]string{"name": p.opssight.Spec.Skyfire.Name, "app": "opssight"})
	pod, err := p.perceptorSkyfirePod()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create skyfire volumes")
	}
	rc.AddPod(pod)

	return rc, nil
}

func (p *SpecConfig) perceptorSkyfirePod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name:           p.opssight.Spec.Skyfire.Name,
		ServiceAccount: p.opssight.Spec.Skyfire.ServiceAccount,
	})
	pod.AddLabels(map[string]string{"name": p.opssight.Spec.Skyfire.Name, "app": "opssight"})

	cont, err := p.perceptorSkyfireContainer()
	if err != nil {
		return nil, err
	}
	err = pod.AddContainer(cont)
	if err != nil {
		return nil, errors.Annotate(err, "unable to add skyfire container")
	}

	vols, err := p.perceptorSkyfireVolumes()
	if err != nil {
		return nil, errors.Annotate(err, "error creating skyfire volumes")
	}
	for _, v := range vols {
		err = pod.AddVolume(v)
		if err != nil {
			return nil, errors.Annotate(err, "error add pod volume")
		}
	}

	return pod, nil
}

func (p *SpecConfig) pyfireContainer() *components.Container {
	return components.NewContainer(horizonapi.ContainerConfig{
		Name:    p.opssight.Spec.Skyfire.Name,
		Image:   p.opssight.Spec.Skyfire.Image,
		Command: []string{"python3"},
		Args: []string{
			"src/main.py",
			fmt.Sprintf("/etc/skyfire/%s.json", p.opssight.Spec.ConfigMapName),
		},
		MinCPU: p.opssight.Spec.DefaultCPU,
		MinMem: p.opssight.Spec.DefaultMem,
	})
}

func (p *SpecConfig) golangSkyfireContainer() *components.Container {
	return components.NewContainer(horizonapi.ContainerConfig{
		Name:    p.opssight.Spec.Skyfire.Name,
		Image:   p.opssight.Spec.Skyfire.Image,
		Command: []string{fmt.Sprintf("./%s", p.opssight.Spec.Skyfire.Name)},
		Args:    []string{fmt.Sprintf("/etc/skyfire/%s.json", p.opssight.Spec.ConfigMapName)},
		MinCPU:  p.opssight.Spec.DefaultCPU,
		MinMem:  p.opssight.Spec.DefaultMem,
	})
}

func (p *SpecConfig) perceptorSkyfireContainer() (*components.Container, error) {
	container := p.pyfireContainer()

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: fmt.Sprintf("%d", p.opssight.Spec.Skyfire.Port),
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

	err = container.AddEnv(horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, FromName: p.opssight.Spec.SecretName})
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (p *SpecConfig) perceptorSkyfireVolumes() ([]*components.Volume, error) {
	vols := []*components.Volume{p.configMapVolume("skyfire")}

	vol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "logs",
		Medium:     horizonapi.StorageMediumDefault,
	})
	if err != nil {
		return nil, errors.Annotate(err, "failed to create empty dir volume")
	}
	vols = append(vols, vol)

	return vols, nil
}

// PerceptorSkyfireService creates a service for perceptor skyfire
func (p *SpecConfig) PerceptorSkyfireService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      p.opssight.Spec.Skyfire.Name,
		Namespace: p.opssight.Spec.Namespace,
	})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(p.opssight.Spec.Skyfire.Port),
		TargetPort: fmt.Sprintf("%d", p.opssight.Spec.Skyfire.Port),
		Protocol:   horizonapi.ProtocolTCP,
		Name:       "main-skyfire",
	})
	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(p.opssight.Spec.Skyfire.PrometheusPort),
		TargetPort: fmt.Sprintf("%d", p.opssight.Spec.Skyfire.PrometheusPort),
		Protocol:   horizonapi.ProtocolTCP,
		Name:       "skyfire-prometheus",
	})

	service.AddLabels(map[string]string{"name": p.opssight.Spec.Skyfire.Name, "app": "opssight"})
	service.AddSelectors(map[string]string{"name": p.opssight.Spec.Skyfire.Name})

	return service
}

// PerceptorSkyfireServiceAccount creates a service account for perceptor skyfire
func (p *SpecConfig) PerceptorSkyfireServiceAccount() *components.ServiceAccount {
	serviceAccount := components.NewServiceAccount(horizonapi.ServiceAccountConfig{
		Name:      "skyfire",
		Namespace: p.opssight.Spec.Namespace,
	})
	serviceAccount.AddLabels(map[string]string{"name": "skyfire", "app": "opssight"})
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
		Resources: []string{"pods", "nodes", "namespaces"},
		Verbs:     []string{"get", "watch", "list", "create", "delete"},
	})
	clusterRole.AddLabels(map[string]string{"name": "skyfire", "app": "opssight"})

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
		Namespace: p.opssight.Spec.Namespace,
	})
	clusterRoleBinding.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     clusterRole.GetName(),
	})
	clusterRoleBinding.AddLabels(map[string]string{"name": "skyfire", "app": "opssight"})

	return clusterRoleBinding
}
