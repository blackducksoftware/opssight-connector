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

// PerceptorReplicationController creates a replication controller for perceptor
func (p *SpecConfig) PerceptorReplicationController() (*components.ReplicationController, error) {
	replicas := int32(1)
	rc := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      p.opssight.Spec.Perceptor.Name,
		Namespace: p.opssight.Spec.Namespace,
	})
	pod, err := p.perceptorPod()
	if err != nil {
		return nil, errors.Trace(err)
	}
	err = rc.AddPod(pod)
	if err != nil {
		return nil, errors.Trace(err)
	}
	rc.AddLabelSelectors(map[string]string{"name": p.opssight.Spec.Perceptor.Name, "app": "opssight"})
	return rc, nil
}

func (p *SpecConfig) perceptorPod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name: p.opssight.Spec.Perceptor.Name,
	})
	pod.AddLabels(map[string]string{"name": p.opssight.Spec.Perceptor.Name, "app": "opssight"})
	cont, err := p.perceptorContainer()
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = pod.AddContainer(cont)
	if err != nil {
		return nil, errors.Trace(err)
	}
	err = pod.AddVolume(p.configMapVolume(p.opssight.Spec.Perceptor.Name))
	if err != nil {
		return nil, errors.Trace(err)
	}

	return pod, nil
}

func (p *SpecConfig) perceptorContainer() (*components.Container, error) {
	name := p.opssight.Spec.Perceptor.Name
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:    name,
		Image:   p.opssight.Spec.Perceptor.Image,
		Command: []string{fmt.Sprintf("./%s", name)},
		Args:    []string{fmt.Sprintf("/etc/%s/%s.json", name, p.opssight.Spec.ConfigMapName)},
		MinCPU:  p.opssight.Spec.DefaultCPU,
		MinMem:  p.opssight.Spec.DefaultMem,
	})
	container.AddPort(horizonapi.PortConfig{
		ContainerPort: fmt.Sprintf("%d", p.opssight.Spec.Perceptor.Port),
		Protocol:      horizonapi.ProtocolTCP,
	})
	err := container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      name,
		MountPath: fmt.Sprintf("/etc/%s", name),
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = container.AddEnv(horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, FromName: p.opssight.Spec.SecretName})

	if err != nil {
		return nil, errors.Trace(err)
	}

	return container, nil
}

// PerceptorService creates a service for perceptor
func (p *SpecConfig) PerceptorService() (*components.Service, error) {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      p.opssight.Spec.Perceptor.Name,
		Namespace: p.opssight.Spec.Namespace,
	})

	err := service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(p.opssight.Spec.Perceptor.Port),
		TargetPort: fmt.Sprintf("%d", p.opssight.Spec.Perceptor.Port),
		Protocol:   horizonapi.ProtocolTCP,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	service.AddLabels(map[string]string{"name": p.opssight.Spec.Perceptor.Name, "app": "opssight"})
	service.AddSelectors(map[string]string{"name": p.opssight.Spec.Perceptor.Name})

	return service, nil
}

// PerceptorSecret create a secret for perceptor
func (p *SpecConfig) PerceptorSecret() *components.Secret {
	secretConfig := horizonapi.SecretConfig{
		Name:      p.opssight.Spec.SecretName,
		Namespace: p.opssight.Spec.Namespace,
		Type:      horizonapi.SecretTypeOpaque,
	}
	secret := components.NewSecret(secretConfig)
	secret.AddLabels(map[string]string{"name": p.opssight.Spec.SecretName, "app": "opssight"})
	return secret
}
