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
	"encoding/json"
	"fmt"
	"math"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/juju/errors"
)

// ScannerReplicationController creates a replication controller for the perceptor scanner
func (p *SpecConfig) ScannerReplicationController() (*components.ReplicationController, error) {
	replicas := int32(math.Ceil(float64(*p.config.ConcurrentScanLimit) / 2.0))
	rc := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      p.config.ContainerNames["perceptor-scanner"],
		Namespace: p.config.Namespace,
	})

	rc.AddLabelSelectors(map[string]string{"name": p.config.ContainerNames["perceptor-scanner"]})

	pod, err := p.scannerPod()
	if err != nil {
		return nil, fmt.Errorf("failed to create scanner pod: %v", err)
	}
	rc.AddPod(pod)

	return rc, nil
}

func (p *SpecConfig) scannerPod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name:           p.config.ContainerNames["perceptor-scanner"],
		ServiceAccount: p.config.ServiceAccounts["perceptor-image-facade"],
	})
	pod.AddLabels(map[string]string{"name": p.config.ContainerNames["perceptor-scanner"]})

	pod.AddContainer(p.scannerContainer())
	pod.AddContainer(p.imageFacadeContainer())

	vols, err := p.scannerVolumes()
	if err != nil {
		return nil, fmt.Errorf("error creating scanner volumes: %v", err)
	}

	newVols, err := p.imageFacadeVolumes()
	if err != nil {
		return nil, fmt.Errorf("error creating image facade volumes: %v", err)
	}
	for _, v := range append(vols, newVols...) {
		pod.AddVolume(v)
	}

	return pod, nil
}

func (p *SpecConfig) scannerContainer() *components.Container {
	priv := false
	name := p.config.ContainerNames["perceptor-scanner"]
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:       name,
		Image:      fmt.Sprintf("%s/%s/%s:%s", p.config.Registry, p.config.ImagePath, p.config.ScannerImageName, p.config.ScannerImageVersion),
		Command:    []string{fmt.Sprintf("./%s", name)},
		Args:       []string{fmt.Sprintf("/etc/%s/%s.yaml", name, name)},
		MinCPU:     p.config.DefaultCPU,
		MinMem:     p.config.DefaultMem,
		Privileged: &priv,
	})

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: fmt.Sprintf("%d", *p.config.ScannerPort),
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
		NameOrPrefix: p.config.HubUserPasswordEnvVar,
		Type:         horizonapi.EnvFromSecret,
		KeyOrVal:     "HubUserPassword",
		FromName:     p.config.SecretName,
	})

	return container
}

func (p *SpecConfig) imageFacadeContainer() *components.Container {
	priv := true
	name := p.config.ContainerNames["perceptor-image-facade"]
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:       name,
		Image:      fmt.Sprintf("%s/%s/%s:%s", p.config.Registry, p.config.ImagePath, p.config.ImageFacadeImageName, p.config.ImageFacadeImageVersion),
		Command:    []string{fmt.Sprintf("./%s", name)},
		Args:       []string{fmt.Sprintf("/etc/%s/%s.json", name, name)},
		MinCPU:     p.config.DefaultCPU,
		MinMem:     p.config.DefaultMem,
		Privileged: &priv,
	})

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: fmt.Sprintf("%d", *p.config.ImageFacadePort),
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
	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "dir-docker-socket",
		MountPath: "/var/run/docker.sock",
	})

	return container
}

func (p *SpecConfig) scannerVolumes() ([]*components.Volume, error) {
	vols := []*components.Volume{}

	vols = append(vols, components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      p.config.ContainerNames["perceptor-scanner"],
		MapOrSecretName: p.config.ContainerNames["perceptor-scanner"],
	}))

	vol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "var-images",
		Medium:     horizonapi.StorageMediumDefault,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create empty dir volume: %v", err)
	}
	vols = append(vols, vol)

	return vols, nil
}

func (p *SpecConfig) imageFacadeVolumes() ([]*components.Volume, error) {
	vols := []*components.Volume{}

	vols = append(vols, components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      p.config.ContainerNames["perceptor-image-facade"],
		MapOrSecretName: p.config.ContainerNames["perceptor-image-facade"],
	}))

	vol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "var-images",
		Medium:     horizonapi.StorageMediumDefault,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create empty dir volume: %v", err)
	}
	vols = append(vols, vol)

	vols = append(vols, components.NewHostPathVolume(horizonapi.HostPathVolumeConfig{
		VolumeName: "dir-docker-socket",
		Path:       "/var/run/docker.sock",
	}))

	return vols, nil
}

// ScannerService creates a service for perceptor scanner
func (p *SpecConfig) ScannerService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      p.config.ContainerNames["perceptor-scanner"],
		Namespace: p.config.Namespace,
	})
	service.AddSelectors(map[string]string{"name": p.config.ContainerNames["perceptor-scanner"]})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(*p.config.ScannerPort),
		TargetPort: fmt.Sprintf("%d", *p.config.ScannerPort),
		Protocol:   horizonapi.ProtocolTCP,
	})

	return service
}

// ImageFacadeService creates a service for perceptor image-facade
func (p *SpecConfig) ImageFacadeService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      p.config.ContainerNames["perceptor-image-facade"],
		Namespace: p.config.Namespace,
	})
	service.AddSelectors(map[string]string{"name": p.config.ContainerNames["perceptor-image-facade"]})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(*p.config.ImageFacadePort),
		TargetPort: fmt.Sprintf("%d", *p.config.ImageFacadePort),
		Protocol:   horizonapi.ProtocolTCP,
	})

	return service
}

// ScannerConfigMap creates a config map for the perceptor scanner
func (p *SpecConfig) ScannerConfigMap() (*components.ConfigMap, error) {
	configMap := components.NewConfigMap(horizonapi.ConfigMapConfig{
		Name:      p.config.ContainerNames["perceptor-scanner"],
		Namespace: p.config.Namespace,
	})
	data := map[string]interface{}{
		"Hub": map[string]interface{}{
			"Port":                 *p.config.HubPort,
			"User":                 p.config.HubUser,
			"PasswordEnvVar":       p.config.HubUserPasswordEnvVar,
			"ClientTimeoutSeconds": *p.config.HubClientTimeoutScannerSeconds,
		},
		"ImageFacade": map[string]interface{}{
			"Port": *p.config.ImageFacadePort,
			"Host": "localhost",
		},
		"Perceptor": map[string]interface{}{
			"Port": *p.config.PerceptorPort,
			"Host": p.config.PerceptorImageName,
		},
		"Port":     *p.config.ScannerPort,
		"LogLevel": p.config.LogLevel,
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Trace(err)
	}
	configMap.AddData(map[string]string{fmt.Sprintf("%s.yaml", p.config.ContainerNames["perceptor-scanner"]): string(bytes)})

	return configMap, nil
}

//ImageFacadeConfigMap creates a config map for the perceptor image-facade
func (p *SpecConfig) ImageFacadeConfigMap() (*components.ConfigMap, error) {
	configMap := components.NewConfigMap(horizonapi.ConfigMapConfig{
		Name:      p.config.ContainerNames["perceptor-image-facade"],
		Namespace: p.config.Namespace,
	})
	data := map[string]interface{}{
		"PrivateDockerRegistries": p.config.InternalRegistries,
		"Port":                    *p.config.ImageFacadePort,
		"LogLevel":                p.config.LogLevel,
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Trace(err)
	}
	configMap.AddData(map[string]string{fmt.Sprintf("%s.json", p.config.ContainerNames["perceptor-image-facade"]): string(bytes)})

	return configMap, nil
}

// ScannerServiceAccount creates a service account for the perceptor scanner
func (p *SpecConfig) ScannerServiceAccount() *components.ServiceAccount {
	serviceAccount := components.NewServiceAccount(horizonapi.ServiceAccountConfig{
		Name:      p.config.ServiceAccounts["perceptor-image-facade"],
		Namespace: p.config.Namespace,
	})

	return serviceAccount
}

// ScannerClusterRoleBinding creates a cluster role binding for the perceptor scanner
func (p *SpecConfig) ScannerClusterRoleBinding() *components.ClusterRoleBinding {
	scannerCRB := components.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		Name:       p.config.ContainerNames["perceptor-scanner"],
		APIVersion: "rbac.authorization.k8s.io/v1",
	})

	scannerCRB.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      p.config.ServiceAccounts["perceptor-image-facade"],
		Namespace: p.config.Namespace,
	})
	scannerCRB.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     "cluster-admin",
	})

	return scannerCRB
}
