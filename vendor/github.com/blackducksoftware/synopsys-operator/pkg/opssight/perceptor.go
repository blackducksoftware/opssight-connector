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
		Name:      p.config.Perceptor.Name,
		Namespace: p.config.Namespace,
	})
	pod, err := p.perceptorPod()
	if err != nil {
		return nil, errors.Trace(err)
	}
	err = rc.AddPod(pod)
	if err != nil {
		return nil, errors.Trace(err)
	}
	rc.AddLabelSelectors(map[string]string{"name": p.config.Perceptor.Name})
	return rc, nil
}

func (p *SpecConfig) perceptorPod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name: p.config.Perceptor.Name,
	})
	pod.AddLabels(map[string]string{"name": p.config.Perceptor.Name})
	cont, err := p.perceptorContainer()
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = pod.AddContainer(cont)
	if err != nil {
		return nil, errors.Trace(err)
	}
	err = pod.AddVolume(p.configMapVolume(p.config.Perceptor.Name))
	if err != nil {
		return nil, errors.Trace(err)
	}

	return pod, nil
}

func (p *SpecConfig) perceptorContainer() (*components.Container, error) {
	name := p.config.Perceptor.Name
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:    name,
		Image:   p.config.Perceptor.Image,
		Command: []string{fmt.Sprintf("./%s", name)},
		Args:    []string{fmt.Sprintf("/etc/%s/%s.json", name, p.config.ConfigMapName)},
		MinCPU:  p.config.DefaultCPU,
		MinMem:  p.config.DefaultMem,
	})
	container.AddPort(horizonapi.PortConfig{
		ContainerPort: fmt.Sprintf("%d", p.config.Perceptor.Port),
		Protocol:      horizonapi.ProtocolTCP,
	})
	err := container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      name,
		MountPath: fmt.Sprintf("/etc/%s", name),
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = container.AddEnv(horizonapi.EnvConfig{
		NameOrPrefix: p.config.Blackduck.PasswordEnvVar,
		Type:         horizonapi.EnvFromSecret,
		KeyOrVal:     "HubUserPassword",
		FromName:     p.config.SecretName,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	return container, nil
}

// PerceptorService creates a service for perceptor
func (p *SpecConfig) PerceptorService() (*components.Service, error) {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      p.config.Perceptor.Name,
		Namespace: p.config.Namespace,
	})

	err := service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(p.config.Perceptor.Port),
		TargetPort: fmt.Sprintf("%d", p.config.Perceptor.Port),
		Protocol:   horizonapi.ProtocolTCP,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	service.AddSelectors(map[string]string{"name": p.config.Perceptor.Name})

	return service, nil
}

// PerceptorSecret create a secret for perceptor
func (p *SpecConfig) PerceptorSecret() *components.Secret {
	secretConfig := horizonapi.SecretConfig{
		Name:      p.config.SecretName,
		Namespace: p.config.Namespace,
		Type:      horizonapi.SecretTypeOpaque,
	}
	secret := components.NewSecret(secretConfig)
	return secret
}
