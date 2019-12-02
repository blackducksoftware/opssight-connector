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
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	routev1 "github.com/openshift/api/route/v1"
)

// PerceptorMetricsDeployment creates a deployment for perceptor metrics
func (p *SpecConfig) PerceptorMetricsDeployment() (*components.Deployment, error) {
	replicas := int32(1)
	deployment := components.NewDeployment(horizonapi.DeploymentConfig{
		Replicas:  &replicas,
		Name:      util.GetResourceName(p.opssight.Name, util.OpsSightName, "prometheus"),
		Namespace: p.opssight.Spec.Namespace,
	})
	deployment.AddLabels(map[string]string{"component": "prometheus", "app": "opssight", "name": p.opssight.Name})
	deployment.AddMatchLabelsSelectors(map[string]string{"component": "prometheus", "app": "opssight", "name": p.opssight.Name})

	pod, err := p.perceptorMetricsPod()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create metrics pod")
	}
	deployment.AddPod(pod)

	return deployment, nil
}

func (p *SpecConfig) perceptorMetricsPod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name: util.GetResourceName(p.opssight.Name, util.OpsSightName, "prometheus"),
	})
	pod.AddLabels(map[string]string{"component": "prometheus", "app": "opssight", "name": p.opssight.Name})
	container, err := p.perceptorMetricsContainer()
	if err != nil {
		return nil, errors.Trace(err)
	}
	pod.AddContainer(container)

	vols, err := p.perceptorMetricsVolumes()
	if err != nil {
		return nil, errors.Annotate(err, "error creating metrics volumes")
	}
	for _, v := range vols {
		pod.AddVolume(v)
	}

	if p.opssight.Spec.RegistryConfiguration != nil && len(p.opssight.Spec.RegistryConfiguration.PullSecrets) > 0 {
		pod.AddImagePullSecrets(p.opssight.Spec.RegistryConfiguration.PullSecrets)
	}

	return pod, nil
}

func (p *SpecConfig) perceptorMetricsContainer() (*components.Container, error) {
	container, err := components.NewContainer(horizonapi.ContainerConfig{
		Name:  p.names["prometheus"],
		Image: p.images["prometheus"],
		Args:  []string{"--log.level=debug", "--config.file=/etc/prometheus/prometheus.yml", "--storage.tsdb.path=/tmp/data/", "--storage.tsdb.retention=120d", "--web.listen-address=:3006"},
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: int32(3006),
		Protocol:      horizonapi.ProtocolTCP,
		Name:          "web",
	})

	err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "data",
		MountPath: "/data",
	})
	if err != nil {
		return nil, errors.Trace(err)
	}
	err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "prometheus",
		MountPath: "/etc/prometheus",
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	return container, nil
}

func (p *SpecConfig) perceptorMetricsVolumes() ([]*components.Volume, error) {
	vols := []*components.Volume{}
	vols = append(vols, components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "prometheus",
		MapOrSecretName: util.GetResourceName(p.opssight.Name, util.OpsSightName, "prometheus"),
		DefaultMode:     util.IntToInt32(420),
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
func (p *SpecConfig) PerceptorMetricsService() (*components.Service, error) {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      util.GetResourceName(p.opssight.Name, util.OpsSightName, "prometheus"),
		Namespace: p.opssight.Spec.Namespace,
		Type:      horizonapi.ServiceTypeServiceIP,
	})

	err := service.AddPort(horizonapi.ServicePortConfig{
		Port:       3006,
		TargetPort: "3006",
		Protocol:   horizonapi.ProtocolTCP,
		Name:       fmt.Sprintf("port-%s", "prometheus"),
	})

	service.AddAnnotations(map[string]string{"prometheus.io/scrape": "true"})
	service.AddLabels(map[string]string{"component": "prometheus", "app": "opssight", "name": p.opssight.Name})
	service.AddSelectors(map[string]string{"component": "prometheus", "app": "opssight", "name": p.opssight.Name})

	return service, err
}

// PerceptorMetricsNodePortService creates a nodeport service for perceptor metrics
func (p *SpecConfig) PerceptorMetricsNodePortService() (*components.Service, error) {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      util.GetResourceName(p.opssight.Name, util.OpsSightName, "prometheus-exposed"),
		Namespace: p.opssight.Spec.Namespace,
		Type:      horizonapi.ServiceTypeNodePort,
	})

	err := service.AddPort(horizonapi.ServicePortConfig{
		Port:       3006,
		TargetPort: "3006",
		Protocol:   horizonapi.ProtocolTCP,
		Name:       fmt.Sprintf("port-%s", "prometheus-exposed"),
	})

	service.AddAnnotations(map[string]string{"prometheus.io/scrape": "true"})
	service.AddLabels(map[string]string{"component": "prometheus", "app": "opssight", "name": p.opssight.Name})
	service.AddSelectors(map[string]string{"component": "prometheus", "app": "opssight", "name": p.opssight.Name})

	return service, err
}

// PerceptorMetricsLoadBalancerService creates a loadbalancer service for perceptor metrics
func (p *SpecConfig) PerceptorMetricsLoadBalancerService() (*components.Service, error) {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      util.GetResourceName(p.opssight.Name, util.OpsSightName, "prometheus-exposed"),
		Namespace: p.opssight.Spec.Namespace,
		Type:      horizonapi.ServiceTypeLoadBalancer,
	})

	err := service.AddPort(horizonapi.ServicePortConfig{
		Port:       3006,
		TargetPort: "3006",
		Protocol:   horizonapi.ProtocolTCP,
		Name:       fmt.Sprintf("port-%s", "prometheus-exposed"),
	})

	service.AddAnnotations(map[string]string{"prometheus.io/scrape": "true"})
	service.AddLabels(map[string]string{"component": "prometheus", "app": "opssight", "name": p.opssight.Name})
	service.AddSelectors(map[string]string{"component": "prometheus", "app": "opssight", "name": p.opssight.Name})

	return service, err
}

// PerceptorMetricsConfigMap creates a config map for perceptor metrics
func (p *SpecConfig) PerceptorMetricsConfigMap() (*components.ConfigMap, error) {
	configMap := components.NewConfigMap(horizonapi.ConfigMapConfig{
		Name:      util.GetResourceName(p.opssight.Name, util.OpsSightName, "prometheus"),
		Namespace: p.opssight.Spec.Namespace,
	})

	targets := []string{
		fmt.Sprintf("%s:%d", util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["perceptor"]), 3001),
		fmt.Sprintf("%s:%d", util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["scanner"]), 3003),
		fmt.Sprintf("%s:%d", util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["perceptor-imagefacade"]), 3004),
	}
	if p.opssight.Spec.Perceiver.EnableImagePerceiver {
		targets = append(targets, fmt.Sprintf("%s:%d", util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["image-perceiver"]), 3002))
	}
	if p.opssight.Spec.Perceiver.EnableArtifactoryPerceiver {
		targets = append(targets, fmt.Sprintf("%s:%d", util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["artifactory-perceiver"]), 3007))
	}
	if p.opssight.Spec.Perceiver.EnableQuayPerceiver {
		targets = append(targets, fmt.Sprintf("%s:%d", util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["quay-perceiver"]), 3008))
	}
	if p.opssight.Spec.Perceiver.EnablePodPerceiver {
		targets = append(targets, fmt.Sprintf("%s:%d", util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["pod-perceiver"]), 3002))
	}
	if p.opssight.Spec.EnableSkyfire {
		targets = append(targets, fmt.Sprintf("%s:%d", util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["skyfire"]), 3005))
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
	configMap.AddLabels(map[string]string{"app": "opssight", "name": p.opssight.Name, "component": "prometheus"})
	configMap.AddData(map[string]string{"prometheus.yml": string(bytes)})

	return configMap, nil
}

// GetPrometheusOpenShiftRoute creates the OpenShift route component for the prometheus metrics
func (p *SpecConfig) GetPrometheusOpenShiftRoute() *api.Route {
	namespace := p.opssight.Spec.Namespace
	if strings.ToUpper(p.opssight.Spec.Prometheus.Expose) == util.OPENSHIFT {
		return &api.Route{
			Name:               util.GetResourceName(p.opssight.Name, util.OpsSightName, fmt.Sprintf("%s-metrics", p.names["prometheus"])),
			Namespace:          namespace,
			Kind:               "Service",
			ServiceName:        util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["prometheus"]),
			PortName:           fmt.Sprintf("port-%s", p.names["prometheus"]),
			Labels:             map[string]string{"app": "opssight", "name": p.opssight.Name, "component": "prometheus-metrics"},
			TLSTerminationType: routev1.TLSTerminationEdge,
		}
	}
	return nil
}
