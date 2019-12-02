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
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
)

// ScannerReplicationController creates a replication controller for the perceptor scanner
func (p *SpecConfig) ScannerReplicationController() (*components.ReplicationController, error) {
	replicas := int32(p.opssight.Spec.ScannerPod.ReplicaCount)
	rc := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["scanner"]),
		Namespace: p.opssight.Spec.Namespace,
	})

	rc.AddSelectors(map[string]string{"component": p.names["scanner"], "app": "opssight", "name": p.opssight.Name})
	rc.AddLabels(map[string]string{"component": p.names["scanner"], "app": "opssight", "name": p.opssight.Name})
	pod, err := p.scannerPod()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create scanner pod")
	}
	rc.AddPod(pod)

	return rc, nil
}

func (p *SpecConfig) scannerPod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name:           util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["scanner"]),
		ServiceAccount: util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["perceptor-imagefacade"]),
	})
	pod.AddLabels(map[string]string{"component": p.names["scanner"], "app": "opssight", "name": p.opssight.Name})

	cont, err := p.scannerContainer()
	if err != nil {
		return nil, errors.Trace(err)
	}
	pod.AddContainer(cont)

	facadecont, err := p.imageFacadeContainer()
	if err != nil {
		return nil, errors.Trace(err)
	}
	pod.AddContainer(facadecont)

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

	if p.opssight.Spec.RegistryConfiguration != nil && len(p.opssight.Spec.RegistryConfiguration.PullSecrets) > 0 {
		pod.AddImagePullSecrets(p.opssight.Spec.RegistryConfiguration.PullSecrets)
	}

	return pod, nil
}

func (p *SpecConfig) scannerContainer() (*components.Container, error) {
	priv := false
	name := p.names["scanner"]
	command := name
	if name == "scanner" {
		command = fmt.Sprintf("opssight-%s", name)
	}
	container, err := components.NewContainer(horizonapi.ContainerConfig{
		Name:       name,
		Image:      p.images["scanner"],
		Command:    []string{fmt.Sprintf("./%s", command)},
		Args:       []string{fmt.Sprintf("/etc/%s/%s.json", name, p.names["configmap"])},
		MinCPU:     p.opssight.Spec.ScannerCPU,
		MinMem:     p.opssight.Spec.ScannerMem,
		Privileged: &priv,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: int32(3003),
		Protocol:      horizonapi.ProtocolTCP,
	})

	err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      name,
		MountPath: fmt.Sprintf("/etc/%s", name),
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "var-images",
		MountPath: "/var/images",
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddEnv(horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, FromName: util.GetResourceName(p.opssight.Name, util.OpsSightName, "blackduck")})

	return container, nil
}

func (p *SpecConfig) imageFacadeContainer() (*components.Container, error) {
	priv := false
	if !strings.EqualFold(p.opssight.Spec.ScannerPod.ImageFacade.ImagePullerType, "skopeo") {
		priv = true
	}

	name := p.names["perceptor-imagefacade"]
	command := name
	if name == "image-getter" {
		command = fmt.Sprintf("opssight-%s", name)
	}
	container, err := components.NewContainer(horizonapi.ContainerConfig{
		Name:       name,
		Image:      p.images["perceptor-imagefacade"],
		Command:    []string{fmt.Sprintf("./%s", command)},
		Args:       []string{fmt.Sprintf("/etc/%s/%s.json", name, p.names["configmap"])},
		MinCPU:     p.opssight.Spec.ScannerCPU,
		MinMem:     p.opssight.Spec.ScannerMem,
		Privileged: &priv,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: int32(3004),
		Protocol:      horizonapi.ProtocolTCP,
	})

	err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      name,
		MountPath: fmt.Sprintf("/etc/%s", name),
	})
	if err != nil {
		return nil, errors.Trace(err)
	}
	err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "var-images",
		MountPath: "/var/images",
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	if !strings.EqualFold(p.opssight.Spec.ScannerPod.ImageFacade.ImagePullerType, "skopeo") {
		err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
			Name:      "dir-docker-socket",
			MountPath: "/var/run/docker.sock",
		})
		if err != nil {
			return nil, errors.Trace(err)
		}
	}

	container.AddEnv(horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, FromName: util.GetResourceName(p.opssight.Name, util.OpsSightName, "blackduck")})

	return container, nil
}

func (p *SpecConfig) scannerVolumes() ([]*components.Volume, error) {
	vols := []*components.Volume{p.configMapVolume(p.names["scanner"])}

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
	vols := []*components.Volume{p.configMapVolume(p.names["perceptor-imagefacade"])}

	vol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "var-images",
		Medium:     horizonapi.StorageMediumDefault,
	})

	if err != nil {
		return nil, errors.Annotate(err, "failed to create empty dir volume")
	}
	vols = append(vols, vol)

	if !strings.EqualFold(p.opssight.Spec.ScannerPod.ImageFacade.ImagePullerType, "skopeo") {
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
		Name:      util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["scanner"]),
		Namespace: p.opssight.Spec.Namespace,
		Type:      horizonapi.ServiceTypeServiceIP,
	})
	service.AddLabels(map[string]string{"component": p.names["scanner"], "app": "opssight", "name": p.opssight.Name})
	service.AddSelectors(map[string]string{"component": p.names["scanner"], "app": "opssight", "name": p.opssight.Name})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(3003),
		TargetPort: fmt.Sprintf("%d", 3003),
		Protocol:   horizonapi.ProtocolTCP,
		Name:       fmt.Sprintf("port-%s", p.names["scanner"]),
	})

	return service
}

// ImageFacadeService creates a service for perceptor image-facade
func (p *SpecConfig) ImageFacadeService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["perceptor-imagefacade"]),
		Namespace: p.opssight.Spec.Namespace,
		Type:      horizonapi.ServiceTypeServiceIP,
	})
	// TODO verify that this hits the *perceptor-scanner pod* !!!
	service.AddLabels(map[string]string{"component": p.names["scanner"], "app": "opssight", "name": p.opssight.Name})
	service.AddSelectors(map[string]string{"component": p.names["scanner"], "app": "opssight", "name": p.opssight.Name})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(3004),
		TargetPort: fmt.Sprintf("%d", 3004),
		Protocol:   horizonapi.ProtocolTCP,
		Name:       fmt.Sprintf("port-%s", p.names["perceptor-imagefacade"]),
	})

	return service
}

// ScannerServiceAccount creates a service account for the perceptor scanner
func (p *SpecConfig) ScannerServiceAccount() *components.ServiceAccount {
	serviceAccount := components.NewServiceAccount(horizonapi.ServiceAccountConfig{
		Name:      util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["perceptor-imagefacade"]),
		Namespace: p.opssight.Spec.Namespace,
	})
	serviceAccount.AddLabels(map[string]string{"component": p.names["perceptor-imagefacade"], "app": "opssight", "name": p.opssight.Name})
	return serviceAccount
}

// ScannerClusterRoleBinding creates a cluster role binding for the perceptor scanner
func (p *SpecConfig) ScannerClusterRoleBinding() (*components.ClusterRoleBinding, error) {
	scannerCRB := components.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		Name:       util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["scanner"]),
		APIVersion: "rbac.authorization.k8s.io/v1",
	})

	scannerCRB.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["perceptor-imagefacade"]),
		Namespace: p.opssight.Spec.Namespace,
	})
	scannerCRB.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     "system:image-puller",
	})
	scannerCRB.AddLabels(map[string]string{"component": p.names["scanner"], "app": "opssight", "name": p.opssight.Name})

	return scannerCRB, nil
}
