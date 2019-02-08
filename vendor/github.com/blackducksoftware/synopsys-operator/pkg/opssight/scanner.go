/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.config. The ASF licenses this file
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
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/juju/errors"
)

// ScannerReplicationController creates a replication controller for the perceptor scanner
func (p *SpecConfig) ScannerReplicationController() (*components.ReplicationController, error) {
	replicas := int32(p.config.ScannerPod.ReplicaCount)
	rc := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      p.config.ScannerPod.Name,
		Namespace: p.config.Namespace,
	})

	rc.AddLabelSelectors(map[string]string{"name": p.config.ScannerPod.Name})

	pod, err := p.scannerPod()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create scanner pod")
	}
	rc.AddPod(pod)

	return rc, nil
}

func (p *SpecConfig) scannerPod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name:           p.config.ScannerPod.Name,
		ServiceAccount: p.config.ScannerPod.ImageFacade.ServiceAccount,
	})
	pod.AddLabels(map[string]string{"name": p.config.ScannerPod.Name})

	pod.AddContainer(p.scannerContainer())
	pod.AddContainer(p.imageFacadeContainer())

	vols, err := p.scannerVolumes()
	if err != nil {
		return nil, errors.Annotate(err, "error creating scanner volumes")
	}

	newVols, err := p.imageFacadeVolumes()
	if err != nil {
		return nil, errors.Annotate(err, "error creating image facade volumes")
	}
	for _, v := range append(vols, newVols...) {
		pod.AddVolume(v)
	}

	return pod, nil
}

func (p *SpecConfig) scannerContainer() *components.Container {
	priv := false
	name := p.config.ScannerPod.Scanner.Name
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:       name,
		Image:      p.config.ScannerPod.Scanner.Image,
		Command:    []string{fmt.Sprintf("./%s", name)},
		Args:       []string{fmt.Sprintf("/etc/%s/%s.json", name, p.config.ConfigMapName)},
		MinCPU:     p.config.DefaultCPU,
		MinMem:     p.config.DefaultMem,
		Privileged: &priv,
	})

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: fmt.Sprintf("%d", p.config.ScannerPod.Scanner.Port),
		Protocol:      horizonapi.ProtocolTCP,
	})

	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      name,
		MountPath: fmt.Sprintf("/etc/%s", name),
	})
	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "var-images",
		MountPath: "/var/images",
	})

	container.AddEnv(horizonapi.EnvConfig{
		NameOrPrefix: p.config.Blackduck.PasswordEnvVar,
		Type:         horizonapi.EnvFromSecret,
		KeyOrVal:     "HubUserPassword",
		FromName:     p.config.SecretName,
	})

	return container
}

func (p *SpecConfig) imageFacadeContainer() *components.Container {
	priv := true
	name := p.config.ScannerPod.ImageFacade.Name
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:       name,
		Image:      p.config.ScannerPod.ImageFacade.Image,
		Command:    []string{fmt.Sprintf("./%s", name)},
		Args:       []string{fmt.Sprintf("/etc/%s/%s.json", name, p.config.ConfigMapName)},
		MinCPU:     p.config.DefaultCPU,
		MinMem:     p.config.DefaultMem,
		Privileged: &priv,
	})

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: fmt.Sprintf("%d", p.config.ScannerPod.ImageFacade.Port),
		Protocol:      horizonapi.ProtocolTCP,
	})

	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      name,
		MountPath: fmt.Sprintf("/etc/%s", name),
	})
	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "var-images",
		MountPath: "/var/images",
	})

	if !strings.EqualFold(p.config.ScannerPod.ImageFacade.ImagePullerType, "skopeo") {
		container.AddVolumeMount(horizonapi.VolumeMountConfig{
			Name:      "dir-docker-socket",
			MountPath: "/var/run/docker.sock",
		})
	}

	return container
}

func (p *SpecConfig) scannerVolumes() ([]*components.Volume, error) {
	vols := []*components.Volume{p.configMapVolume(p.config.ScannerPod.Scanner.Name)}

	vol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "var-images",
		Medium:     horizonapi.StorageMediumDefault,
	})

	if err != nil {
		return nil, errors.Annotate(err, "failed to create empty dir volume")
	}
	vols = append(vols, vol)

	return vols, nil
}

func (p *SpecConfig) imageFacadeVolumes() ([]*components.Volume, error) {
	vols := []*components.Volume{p.configMapVolume(p.config.ScannerPod.ImageFacade.Name)}

	vol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "var-images",
		Medium:     horizonapi.StorageMediumDefault,
	})

	if err != nil {
		return nil, errors.Annotate(err, "failed to create empty dir volume")
	}
	vols = append(vols, vol)

	if !strings.EqualFold(p.config.ScannerPod.ImageFacade.ImagePullerType, "skopeo") {
		vols = append(vols, components.NewHostPathVolume(horizonapi.HostPathVolumeConfig{
			VolumeName: "dir-docker-socket",
			Path:       "/var/run/docker.sock",
		}))
	}

	return vols, nil
}

// ScannerService creates a service for perceptor scanner
func (p *SpecConfig) ScannerService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      p.config.ScannerPod.Scanner.Name,
		Namespace: p.config.Namespace,
	})
	service.AddSelectors(map[string]string{"name": p.config.ScannerPod.Name})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(p.config.ScannerPod.Scanner.Port),
		TargetPort: fmt.Sprintf("%d", p.config.ScannerPod.Scanner.Port),
		Protocol:   horizonapi.ProtocolTCP,
	})

	return service
}

// ImageFacadeService creates a service for perceptor image-facade
func (p *SpecConfig) ImageFacadeService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      p.config.ScannerPod.ImageFacade.Name,
		Namespace: p.config.Namespace,
	})
	// TODO verify that this hits the *perceptor-scanner pod* !!!
	service.AddSelectors(map[string]string{"name": p.config.ScannerPod.Name})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(p.config.ScannerPod.ImageFacade.Port),
		TargetPort: fmt.Sprintf("%d", p.config.ScannerPod.ImageFacade.Port),
		Protocol:   horizonapi.ProtocolTCP,
	})

	return service
}

// ScannerServiceAccount creates a service account for the perceptor scanner
func (p *SpecConfig) ScannerServiceAccount() *components.ServiceAccount {
	serviceAccount := components.NewServiceAccount(horizonapi.ServiceAccountConfig{
		Name:      p.config.ScannerPod.ImageFacade.ServiceAccount,
		Namespace: p.config.Namespace,
	})

	return serviceAccount
}

// ScannerClusterRoleBinding creates a cluster role binding for the perceptor scanner
func (p *SpecConfig) ScannerClusterRoleBinding() *components.ClusterRoleBinding {
	scannerCRB := components.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		Name:       p.config.ScannerPod.Name, // TODO is this right?  or should it be .ImageFacade.Name ?
		APIVersion: "rbac.authorization.k8s.io/v1",
	})

	scannerCRB.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      p.config.ScannerPod.ImageFacade.ServiceAccount,
		Namespace: p.config.Namespace,
	})
	scannerCRB.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     "cluster-admin",
	})

	return scannerCRB
}
