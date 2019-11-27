/*
Copyright (C) 2019 Synopsys, Inc.

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

package soperator

import (
	"encoding/json"
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	horizoncomponents "github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	routev1 "github.com/openshift/api/route/v1"
)

// GetPrometheusService creates a Horizon Service component for Prometheus
func (specConfig *PrometheusSpecConfig) GetPrometheusService() []*horizoncomponents.Service {
	services := []*horizoncomponents.Service{}
	// Add Service for Prometheus
	prometheusService := horizoncomponents.NewService(horizonapi.ServiceConfig{
		APIVersion: "v1",
		Name:       "prometheus",
		Namespace:  specConfig.Namespace,
	})
	prometheusService.AddAnnotations(map[string]string{"prometheus.io/scrape": "true"})
	prometheusService.AddSelectors(map[string]string{"app": "synopsys-operator", "component": "prometheus"})
	prometheusService.AddPort(horizonapi.ServicePortConfig{
		Name:       "prometheus",
		Port:       9090,
		TargetPort: "9090",
		Protocol:   horizonapi.ProtocolTCP,
	})

	prometheusService.AddLabels(map[string]string{"app": "synopsys-operator", "component": "prometheus"})
	services = append(services, prometheusService)

	if strings.EqualFold(specConfig.Expose, util.NODEPORT) || strings.EqualFold(specConfig.Expose, util.LOADBALANCER) {

		var exposedServiceType horizonapi.ServiceType
		if strings.EqualFold(specConfig.Expose, util.NODEPORT) {
			exposedServiceType = horizonapi.ServiceTypeNodePort
		} else {
			exposedServiceType = horizonapi.ServiceTypeLoadBalancer
		}
		// Add Service for Prometheus
		prometheusExposedService := horizoncomponents.NewService(horizonapi.ServiceConfig{
			APIVersion: "v1",
			Name:       "prometheus-exposed",
			Namespace:  specConfig.Namespace,
			Type:       exposedServiceType,
		})
		prometheusExposedService.AddAnnotations(map[string]string{"prometheus.io/scrape": "true"})
		prometheusExposedService.AddSelectors(map[string]string{"app": "synopsys-operator", "component": "prometheus"})
		prometheusExposedService.AddPort(horizonapi.ServicePortConfig{
			Name:       "prometheus",
			Port:       9090,
			TargetPort: "9090",
			Protocol:   horizonapi.ProtocolTCP,
		})

		prometheusExposedService.AddLabels(map[string]string{"app": "synopsys-operator", "component": "prometheus"})
		services = append(services, prometheusExposedService)
	}
	return services
}

// GetPrometheusDeployment creates a Horizon Deployment component for Prometheus
func (specConfig *PrometheusSpecConfig) GetPrometheusDeployment() (*horizoncomponents.Deployment, error) {
	// Deployment
	var prometheusDeploymentReplicas int32 = 1
	prometheusDeployment := horizoncomponents.NewDeployment(horizonapi.DeploymentConfig{
		APIVersion: "extensions/v1beta1",
		Name:       "prometheus",
		Namespace:  specConfig.Namespace,
		Replicas:   &prometheusDeploymentReplicas,
	})
	prometheusDeployment.AddMatchLabelsSelectors(map[string]string{"app": "synopsys-operator", "component": "prometheus"})

	prometheusPod := horizoncomponents.NewPod(horizonapi.PodConfig{
		APIVersion: "v1",
		Name:       "prometheus",
		Namespace:  specConfig.Namespace,
	})

	prometheusContainer, err := horizoncomponents.NewContainer(horizonapi.ContainerConfig{
		Name:  "prometheus",
		Args:  []string{"--log.level=debug", "--config.file=/etc/prometheus/prometheus.yml", "--storage.tsdb.path=/tmp/data/"},
		Image: specConfig.Image,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	prometheusContainer.AddPort(horizonapi.PortConfig{
		Name:          "web",
		ContainerPort: int32(9090),
	})

	err = prometheusContainer.AddVolumeMount(horizonapi.VolumeMountConfig{
		MountPath: "/data",
		Name:      "data",
	})
	if err != nil {
		return nil, errors.Trace(err)
	}
	err = prometheusContainer.AddVolumeMount(horizonapi.VolumeMountConfig{
		MountPath: "/etc/prometheus",
		Name:      "config-volume",
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	prometheusEmptyDirVolume, err := horizoncomponents.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "data",
	})
	if err != nil {
		return nil, errors.Trace(err)
	}
	prometheusConfigMapVolume := horizoncomponents.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "config-volume",
		MapOrSecretName: "prometheus",
		DefaultMode:     util.IntToInt32(420),
	})

	prometheusPod.AddContainer(prometheusContainer)
	prometheusPod.AddVolume(prometheusEmptyDirVolume)
	prometheusPod.AddVolume(prometheusConfigMapVolume)
	prometheusPod.AddLabels(map[string]string{"app": "synopsys-operator", "component": "prometheus"})
	prometheusDeployment.AddPod(prometheusPod)

	prometheusDeployment.AddLabels(map[string]string{"app": "synopsys-operator", "component": "prometheus"})
	return prometheusDeployment, nil
}

// GetPrometheusConfigMap creates a Horizon ConfigMap component for Prometheus
func (specConfig *PrometheusSpecConfig) GetPrometheusConfigMap() (*horizoncomponents.ConfigMap, error) {
	// Add prometheus config map
	prometheusConfigMap := horizoncomponents.NewConfigMap(horizonapi.ConfigMapConfig{
		APIVersion: "v1",
		Name:       "prometheus",
		Namespace:  specConfig.Namespace,
	})

	data := map[string]interface{}{
		"global": map[string]interface{}{
			"scrape_interval": "5s",
		},
		"scrape_configs": []interface{}{
			map[string]interface{}{
				"job_name":        "synopsys-operator-scrape",
				"scrape_interval": "5s",
				"static_configs": []interface{}{
					map[string]interface{}{
						"targets": []string{"synopsys-operator:8080"},
					},
				},
			},
		},
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Trace(err)
	}

	cmData := map[string]string{}
	cmData["prometheus.yml"] = string(bytes)
	cmData["Image"] = specConfig.Image
	cmData["Expose"] = specConfig.Expose
	prometheusConfigMap.AddData(cmData)

	prometheusConfigMap.AddLabels(map[string]string{"app": "synopsys-operator", "component": "prometheus"})
	return prometheusConfigMap, nil
}

// GetOpenShiftRoute creates the OpenShift route component for the prometheus
func (specConfig *PrometheusSpecConfig) GetOpenShiftRoute() *api.Route {
	if strings.ToUpper(specConfig.Expose) == util.OPENSHIFT {
		return &api.Route{
			Name:               "synopsys-operator-prometheus",
			Namespace:          specConfig.Namespace,
			Kind:               "Service",
			ServiceName:        "prometheus",
			PortName:           "prometheus",
			Labels:             map[string]string{"app": "synopsys-operator", "component": "prometheus"},
			TLSTerminationType: routev1.TLSTerminationEdge,
		}
	}
	return nil
}
