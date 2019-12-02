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
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
)

// PerceptorSkyfireReplicationController creates a replication controller for perceptor skyfire
func (p *SpecConfig) PerceptorSkyfireReplicationController() (*components.ReplicationController, error) {
	replicas := int32(1)
	rc := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["skyfire"]),
		Namespace: p.opssight.Spec.Namespace,
	})
	rc.AddSelectors(map[string]string{"component": p.names["skyfire"], "app": "opssight", "name": p.opssight.Name})
	rc.AddLabels(map[string]string{"component": p.names["skyfire"], "app": "opssight", "name": p.opssight.Name})
	pod, err := p.perceptorSkyfirePod()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create skyfire volumes")
	}
	rc.AddPod(pod)

	return rc, nil
}

func (p *SpecConfig) perceptorSkyfirePod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name:           util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["skyfire"]),
		ServiceAccount: util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["skyfire"]),
	})
	pod.AddLabels(map[string]string{"component": p.names["skyfire"], "app": "opssight", "name": p.opssight.Name})

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

	if p.opssight.Spec.RegistryConfiguration != nil && len(p.opssight.Spec.RegistryConfiguration.PullSecrets) > 0 {
		pod.AddImagePullSecrets(p.opssight.Spec.RegistryConfiguration.PullSecrets)
	}

	return pod, nil
}

func (p *SpecConfig) pyfireContainer() (*components.Container, error) {
	return components.NewContainer(horizonapi.ContainerConfig{
		Name:    "skyfire",
		Image:   p.images["skyfire"],
		Command: []string{"python3"},
		Args: []string{
			"src/main.py",
			fmt.Sprintf("/etc/skyfire/%s.json", p.names["configmap"]),
		},
		MinCPU: p.opssight.Spec.DefaultCPU,
		MinMem: p.opssight.Spec.DefaultMem,
	})
}

func (p *SpecConfig) perceptorSkyfireContainer() (*components.Container, error) {
	container, err := p.pyfireContainer()
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: int32(3005),
		Protocol:      horizonapi.ProtocolTCP,
	})

	err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
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

	container.AddEnv(horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, FromName: util.GetResourceName(p.opssight.Name, util.OpsSightName, "blackduck")})

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
		Name:      util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["skyfire"]),
		Namespace: p.opssight.Spec.Namespace,
		Type:      horizonapi.ServiceTypeServiceIP,
	})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(3005),
		TargetPort: fmt.Sprintf("%d", 3005),
		Protocol:   horizonapi.ProtocolTCP,
		Name:       "main-skyfire",
	})
	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(3006),
		TargetPort: fmt.Sprintf("%d", 3006),
		Protocol:   horizonapi.ProtocolTCP,
		Name:       "skyfire-prometheus",
	})

	service.AddLabels(map[string]string{"component": p.names["skyfire"], "app": "opssight", "name": p.opssight.Name})
	service.AddSelectors(map[string]string{"component": p.names["skyfire"], "app": "opssight", "name": p.opssight.Name})

	return service
}

// PerceptorSkyfireServiceAccount creates a service account for perceptor skyfire
func (p *SpecConfig) PerceptorSkyfireServiceAccount() *components.ServiceAccount {
	serviceAccount := components.NewServiceAccount(horizonapi.ServiceAccountConfig{
		Name:      util.GetResourceName(p.opssight.Name, util.OpsSightName, "skyfire"),
		Namespace: p.opssight.Spec.Namespace,
	})
	serviceAccount.AddLabels(map[string]string{"component": "skyfire", "app": "opssight", "name": p.opssight.Name})
	return serviceAccount
}

// PerceptorSkyfireClusterRole creates a cluster role for perceptor skyfire
func (p *SpecConfig) PerceptorSkyfireClusterRole() *components.ClusterRole {
	clusterRole := components.NewClusterRole(horizonapi.ClusterRoleConfig{
		Name:       util.GetResourceName(p.opssight.Name, util.OpsSightName, "skyfire"),
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	clusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		APIGroups: []string{"*"},
		Resources: []string{"pods", "nodes", "namespaces"},
		Verbs:     []string{"get", "watch", "list", "create", "delete"},
	})
	clusterRole.AddLabels(map[string]string{"component": "skyfire", "app": "opssight", "name": p.opssight.Name})

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
		Name:      util.GetResourceName(p.opssight.Name, util.OpsSightName, "skyfire"),
		Namespace: p.opssight.Spec.Namespace,
	})
	clusterRoleBinding.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     clusterRole.GetName(),
	})
	clusterRoleBinding.AddLabels(map[string]string{"component": "skyfire", "app": "opssight", "name": p.opssight.Name})

	return clusterRoleBinding
}
