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
	"bytes"
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
)

// PerceptorMetricsDeployment creates a deployment for perceptor metrics
func (p *SpecConfig) PerceptorMetricsDeployment() (*components.Deployment, error) {
	replicas := int32(1)
	deployment := components.NewDeployment(horizonapi.DeploymentConfig{
		Replicas:  &replicas,
		Name:      "prometheus",
		Namespace: p.config.Namespace,
	})
	deployment.AddMatchLabelsSelectors(map[string]string{"app": "prometheus"})

	pod, err := p.perceptorMetricsPod()
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics pod: %v", err)
	}
	deployment.AddPod(pod)

	return deployment, nil
}

func (p *SpecConfig) perceptorMetricsPod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name: "prometheus",
	})
	pod.AddLabels(map[string]string{"app": "prometheus"})

	pod.AddContainer(p.perceptorMetricsContainer())

	vols, err := p.perceptorMetricsVolumes()
	if err != nil {
		return nil, fmt.Errorf("error creating metrics volumes: %v", err)
	}
	for _, v := range vols {
		pod.AddVolume(v)
	}

	return pod, nil
}

func (p *SpecConfig) perceptorMetricsContainer() *components.Container {
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:  "prometheus",
		Image: "prom/prometheus:v2.1.0",
		Args:  []string{"--log.level=debug", "--config.file=/etc/prometheus/prometheus.yml", "--storage.tsdb.path=/tmp/data/", "--storage.tsdb.retention=120d"},
	})

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: "9090",
		Protocol:      horizonapi.ProtocolTCP,
		Name:          "web",
	})

	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:        "data",
		MountPath:   "/data",
		Propagation: horizonapi.MountPropagationHostToContainer,
	})
	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:        "prometheus",
		MountPath:   "/etc/prometheus",
		Propagation: horizonapi.MountPropagationHostToContainer,
	})

	return container
}

func (p *SpecConfig) perceptorMetricsVolumes() ([]*components.Volume, error) {
	vols := []*components.Volume{}
	vols = append(vols, components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "prometheus",
		MapOrSecretName: "prometheus",
	}))

	vol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "data",
		Medium:     horizonapi.StorageMediumDefault,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create empty dir volume: %v", err)
	}
	vols = append(vols, vol)

	return vols, nil
}

// PerceptorMetricsService creates a service for perceptor metrics
func (p *SpecConfig) PerceptorMetricsService() *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:          "prometheus",
		Namespace:     p.config.Namespace,
		IPServiceType: horizonapi.ClusterIPServiceTypeNodePort,
	})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       9090,
		TargetPort: "9090",
		Protocol:   horizonapi.ProtocolTCP,
	})

	service.AddAnnotations(map[string]string{"prometheus.io/scrape": "true"})
	service.AddLabels(map[string]string{"name": "prometheus"})
	service.AddSelectors(map[string]string{"app": "prometheus"})

	return service
}

// PerceptorMetricsConfigMap creates a config map for perceptor metrics
func (p *SpecConfig) PerceptorMetricsConfigMap() *components.ConfigMap {
	configMap := components.NewConfigMap(horizonapi.ConfigMapConfig{
		Name:      "prometheus",
		Namespace: p.config.Namespace,
	})

	var promConfig bytes.Buffer
	promConfig.WriteString(fmt.Sprint(`{"global":{"scrape_interval":"5s"},"scrape_configs":[{"job_name":"perceptor-scrape","scrape_interval":"5s","static_configs":[{"targets":["`, p.config.PerceptorImageName, `:`, *p.config.PerceptorPort, `","`, p.config.ScannerImageName, `:`, *p.config.ScannerPort, `","`, p.config.ImageFacadeImageName, `:`, *p.config.ImageFacadePort))
	if p.config.ImagePerceiver != nil && *p.config.ImagePerceiver {
		promConfig.WriteString(fmt.Sprint(`","`, p.config.ImagePerceiverImageName, `:`, *p.config.PerceiverPort))
	}
	if p.config.PodPerceiver != nil && *p.config.PodPerceiver {
		promConfig.WriteString(fmt.Sprint(`","`, p.config.PodPerceiverImageName, `:`, *p.config.PerceiverPort))
	}
	if p.config.PerceptorSkyfire != nil && *p.config.PerceptorSkyfire {
		promConfig.WriteString(fmt.Sprint(`","`, p.config.SkyfireImageName, `:`, *p.config.SkyfirePort))

	}
	promConfig.WriteString(`"]}]}]}`)
	configMap.AddData(map[string]string{"prometheus.yml": promConfig.String()})

	return configMap
}
