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

// PerceptorReplicationController creates a replication controller for perceptor
func (p *SpecConfig) PerceptorReplicationController() (*components.ReplicationController, error) {
	replicas := int32(1)
	rc := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      p.config.ContainerNames["perceptor"],
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
	rc.AddLabelSelectors(map[string]string{"name": p.config.ContainerNames["perceptor"]})
	return rc, nil
}

func (p *SpecConfig) perceptorPod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name: p.config.ContainerNames["perceptor"],
	})
	pod.AddLabels(map[string]string{"name": p.config.ContainerNames["perceptor"]})
	cont, err := p.perceptorContainer()
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = pod.AddContainer(cont)
	if err != nil {
		return nil, errors.Trace(err)
	}
	err = pod.AddVolume(p.perceptorVolume())
	if err != nil {
		return nil, errors.Trace(err)
	}

	return pod, nil
}

func (p *SpecConfig) perceptorContainer() (*components.Container, error) {
	name := p.config.ContainerNames["perceptor"]
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:    name,
		Image:   fmt.Sprintf("%s/%s/%s:%s", p.config.Registry, p.config.ImagePath, p.config.PerceptorImageName, p.config.PerceptorImageVersion),
		Command: []string{fmt.Sprintf("./%s", name)},
		Args:    []string{fmt.Sprintf("/etc/%s/%s.yaml", name, name)},
		MinCPU:  p.config.DefaultCPU,
		MinMem:  p.config.DefaultMem,
	})
	container.AddPort(horizonapi.PortConfig{
		ContainerPort: fmt.Sprintf("%d", *p.config.PerceptorPort),
		Protocol:      horizonapi.ProtocolTCP,
	})
	err := container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:        name,
		MountPath:   fmt.Sprintf("/etc/%s", name),
		Propagation: horizonapi.MountPropagationHostToContainer,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = container.AddEnv(horizonapi.EnvConfig{
		NameOrPrefix: p.config.HubUserPasswordEnvVar,
		Type:         horizonapi.EnvFromSecret,
		KeyOrVal:     "HubUserPassword",
		FromName:     p.config.SecretName,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	return container, nil
}

func (p *SpecConfig) perceptorVolume() *components.Volume {
	return components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      p.config.ContainerNames["perceptor"],
		MapOrSecretName: p.config.ContainerNames["perceptor"],
	})
}

// PerceptorService creates a service for perceptor
func (p *SpecConfig) PerceptorService() (*components.Service, error) {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      p.config.ContainerNames["perceptor"],
		Namespace: p.config.Namespace,
	})

	err := service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(*p.config.PerceptorPort),
		TargetPort: fmt.Sprintf("%d", *p.config.PerceptorPort),
		Protocol:   horizonapi.ProtocolTCP,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	service.AddSelectors(map[string]string{"name": p.config.ContainerNames["perceptor"]})

	return service, nil
}

// PerceptorConfigMap creates a config map for perceptor
func (p *SpecConfig) PerceptorConfigMap() (*components.ConfigMap, error) {
	cm := components.NewConfigMap(horizonapi.ConfigMapConfig{
		Name:      p.config.ContainerNames["perceptor"],
		Namespace: p.config.Namespace,
	})
	data := map[string]interface{}{
		"Hub": map[string]interface{}{
			"Hosts":                     []string{},
			"Port":                      *p.config.HubPort,
			"User":                      p.config.HubUser,
			"PasswordEnvVar":            p.config.HubUserPasswordEnvVar,
			"ClientTimeoutMilliseconds": *p.config.HubClientTimeoutPerceptorMilliseconds,
			"ConcurrentScanLimit":       *p.config.ConcurrentScanLimit,
			"TotalScanLimit":            *p.config.TotalScanLimit,
		},
		"Port":        *p.config.PerceptorPort,
		"LogLevel":    p.config.LogLevel,
		"UseMockMode": *p.config.UseMockMode,
		"Timings": map[string]interface{}{
			"CheckForStalledScansPauseHours": *p.config.CheckForStalledScansPauseHours,
			"StalledScanClientTimeoutHours":  *p.config.StalledScanClientTimeoutHours,
			"ModelMetricsPauseSeconds":       *p.config.ModelMetricsPauseSeconds,
			"UnknownImagePauseMilliseconds":  *p.config.UnknownImagePauseMilliseconds,
		},
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Trace(err)
	}
	cm.AddData(map[string]string{fmt.Sprintf("%s.yaml", p.config.ContainerNames["perceptor"]): string(bytes)})

	return cm, nil
}

// PerceptorSecret create a secret for perceptor
func (p *SpecConfig) PerceptorSecret() *components.Secret {
	secretConfig := horizonapi.SecretConfig{
		Name:      p.config.SecretName,
		Namespace: p.config.Namespace,
		Type:      horizonapi.SecretTypeOpaque,
	}
	secret := components.NewSecret(secretConfig)
	secret.AddData(map[string][]byte{"HubUserPassword": []byte(p.config.HubUserPassword)})

	return secret
}
