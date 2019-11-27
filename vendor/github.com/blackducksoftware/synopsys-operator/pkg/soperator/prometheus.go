/*
Copyright (C) 2019 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownershia. The ASF licenses this file
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
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	"github.com/juju/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// PrometheusSpecConfig represents the Promtheus component
// Its methods include GetComponents() and any functions
// that create Kubernetes Resources for Prometheus
type PrometheusSpecConfig struct {
	Namespace  string
	Image      string
	Expose     string
	RestConfig *rest.Config
	KubeClient *kubernetes.Clientset
}

// NewPrometheus will create a PromtheusSpecConfig type
func NewPrometheus(namespace, image string, expose string, restConfig *rest.Config, kubeClient *kubernetes.Clientset) *PrometheusSpecConfig {
	return &PrometheusSpecConfig{
		Namespace:  namespace,
		Image:      image,
		Expose:     expose,
		RestConfig: restConfig,
		KubeClient: kubeClient,
	}
}

// GetComponents will return a ComponentList representing all
// Kubernetes Resources for Prometheus
func (specConfig *PrometheusSpecConfig) GetComponents() (*api.ComponentList, error) {
	deployment, err := specConfig.GetPrometheusDeployment()
	if err != nil {
		return nil, errors.Trace(err)
	}

	configMap, err := specConfig.GetPrometheusConfigMap()
	if err != nil {
		return nil, errors.Trace(err)
	}
	components := &api.ComponentList{
		Deployments: []*components.Deployment{deployment},
		Services:    specConfig.GetPrometheusService(),
		ConfigMaps:  []*components.ConfigMap{configMap},
	}

	// Add routes for OpenShift
	route := specConfig.GetOpenShiftRoute()
	if route != nil {
		components.Routes = []*api.Route{route}
	}
	return components, nil
}
