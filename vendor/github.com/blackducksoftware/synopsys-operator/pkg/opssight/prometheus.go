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

// PerceptorMetricsDeployment creates a deployment for perceptor metrics
func (p *SpecConfig) PerceptorMetricsDeployment() (*components.Deployment, error) {
	replicas := int32(0)
	if p.config.EnableMetrics {
		replicas = 1
	}
	deployment := components.NewDeployment(horizonapi.DeploymentConfig{
		Replicas:  &replicas,
		Name:      "prometheus",
		Namespace: p.config.Namespace,
	})
	deployment.AddMatchLabelsSelectors(map[string]string{"app": "prometheus"})

	pod, err := p.perceptorMetricsPod()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create metrics pod")
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
		return nil, errors.Annotate(err, "error creating metrics volumes")
	}
	for _, v := range vols {
		pod.AddVolume(v)
	}

	return pod, nil
}

func (p *SpecConfig) perceptorMetricsContainer() *components.Container {
	container := components.NewContainer(horizonapi.ContainerConfig{
		Name:  p.config.Prometheus.Name,
		Image: p.config.Prometheus.Image,
		Args:  []string{"--log.level=debug", "--config.file=/etc/prometheus/prometheus.yml", "--storage.tsdb.path=/tmp/data/", "--storage.tsdb.retention=120d"},
	})

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: fmt.Sprintf("%d", p.config.Prometheus.Port),
		Protocol:      horizonapi.ProtocolTCP,
		Name:          "web",
	})

	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "data",
		MountPath: "/data",
	})
	container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "prometheus",
		MountPath: "/etc/prometheus",
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
		return nil, errors.Annotate(err, "failed to create empty dir volume")
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
func (p *SpecConfig) PerceptorMetricsConfigMap() (*components.ConfigMap, error) {
	configMap := components.NewConfigMap(horizonapi.ConfigMapConfig{
		Name:      "prometheus",
		Namespace: p.config.Namespace,
	})

	/*
			example:

		{
		  "global": {
		    "scrape_interval": "5s"
		  },
		  "scrape_configs": [
		    {
		      "job_name": "perceptor-scrape",
		      "scrape_interval": "5s",
		      "static_configs": [
		        {
		          "targets": [
		            "perceptor:3001",
		            "perceptor-scanner:3003",
		            "perceptor-imagefacade:3004",
		            "pod-perceiver:3002"
		          ]
		        }
		      ]
		    }
		  ]
		}
	*/
	targets := []string{
		fmt.Sprintf("%s:%d", p.config.Perceptor.Name, p.config.Perceptor.Port),
		fmt.Sprintf("%s:%d", p.config.ScannerPod.Scanner.Name, p.config.ScannerPod.Scanner.Port),
		fmt.Sprintf("%s:%d", p.config.ScannerPod.ImageFacade.Name, p.config.ScannerPod.ImageFacade.Port),
	}
	if p.config.Perceiver.EnableImagePerceiver {
		targets = append(targets, fmt.Sprintf("%s:%d", p.config.Perceiver.ImagePerceiver.Name, p.config.Perceiver.Port))
	}
	if p.config.Perceiver.EnablePodPerceiver {
		targets = append(targets, fmt.Sprintf("%s:%d", p.config.Perceiver.PodPerceiver.Name, p.config.Perceiver.Port))
	}
	if p.config.EnableSkyfire {
		targets = append(targets, fmt.Sprintf("%s:%d", p.config.Skyfire.Name, p.config.Skyfire.PrometheusPort))
	}
	data := map[string]interface{}{
		"global": map[string]interface{}{
			"scrape_interval": "5s",
		},
		"scrape_configs": []interface{}{
			map[string]interface{}{
				"job_name":        "perceptor-scrape",
				"scrape_interval": "5s",
				"static_configs": []interface{}{
					map[string]interface{}{
						"targets": targets,
					},
				},
			},
		},
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Trace(err)
	}
	configMap.AddData(map[string]string{"prometheus.yml": string(bytes)})

	return configMap, nil
}
