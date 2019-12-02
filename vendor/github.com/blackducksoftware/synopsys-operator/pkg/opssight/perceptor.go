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
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	routev1 "github.com/openshift/api/route/v1"
)

// PerceptorReplicationController creates a replication controller for perceptor
func (p *SpecConfig) PerceptorReplicationController() (*components.ReplicationController, error) {
	replicas := int32(1)
	rc := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["perceptor"]),
		Namespace: p.opssight.Spec.Namespace,
	})
	pod, err := p.perceptorPod()
	if err != nil {
		return nil, errors.Trace(err)
	}
	rc.AddPod(pod)
	rc.AddSelectors(map[string]string{"component": p.names["perceptor"], "app": "opssight", "name": p.opssight.Name})
	rc.AddLabels(map[string]string{"component": p.names["perceptor"], "app": "opssight", "name": p.opssight.Name})
	return rc, nil
}

func (p *SpecConfig) perceptorPod() (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name: util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["perceptor"]),
	})
	pod.AddLabels(map[string]string{"component": p.names["perceptor"], "app": "opssight", "name": p.opssight.Name})
	cont, err := p.perceptorContainer()
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = pod.AddContainer(cont)
	if err != nil {
		return nil, errors.Trace(err)
	}
	err = pod.AddVolume(p.configMapVolume(p.names["perceptor"]))
	if err != nil {
		return nil, errors.Trace(err)
	}

	if p.opssight.Spec.RegistryConfiguration != nil && len(p.opssight.Spec.RegistryConfiguration.PullSecrets) > 0 {
		pod.AddImagePullSecrets(p.opssight.Spec.RegistryConfiguration.PullSecrets)
	}

	return pod, nil
}

func (p *SpecConfig) perceptorContainer() (*components.Container, error) {
	name := p.names["perceptor"]
	command := name
	if name == "core" {
		command = fmt.Sprintf("opssight-%s", name)
	}
	container, err := components.NewContainer(horizonapi.ContainerConfig{
		Name:    name,
		Image:   p.images["perceptor"],
		Command: []string{fmt.Sprintf("./%s", command)},
		Args:    []string{fmt.Sprintf("/etc/%s/%s.json", name, p.names["configmap"])},
		MinCPU:  p.opssight.Spec.DefaultCPU,
		MinMem:  p.opssight.Spec.DefaultMem,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: int32(3001),
		Protocol:      horizonapi.ProtocolTCP,
	})
	err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      name,
		MountPath: fmt.Sprintf("/etc/%s", name),
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddEnv(horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, FromName: util.GetResourceName(p.opssight.Name, util.OpsSightName, "blackduck")})

	return container, nil
}

// PerceptorService creates a service for perceptor
func (p *SpecConfig) PerceptorService() (*components.Service, error) {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["perceptor"]),
		Namespace: p.opssight.Spec.Namespace,
		Type:      horizonapi.ServiceTypeServiceIP,
	})

	err := service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(3001),
		TargetPort: fmt.Sprintf("%d", 3001),
		Protocol:   horizonapi.ProtocolTCP,
		Name:       fmt.Sprintf("port-%s", p.names["perceptor"]),
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	service.AddLabels(map[string]string{"component": p.names["perceptor"], "app": "opssight", "name": p.opssight.Name})
	service.AddSelectors(map[string]string{"component": p.names["perceptor"], "app": "opssight", "name": p.opssight.Name})

	return service, nil
}

// PerceptorNodePortService creates a nodeport service for perceptor
func (p *SpecConfig) PerceptorNodePortService() (*components.Service, error) {
	name := fmt.Sprintf("%s-exposed", p.names["perceptor"])
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      util.GetResourceName(p.opssight.Name, util.OpsSightName, name),
		Namespace: p.opssight.Spec.Namespace,
		Type:      horizonapi.ServiceTypeNodePort,
	})

	err := service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(3001),
		TargetPort: fmt.Sprintf("%d", 3001),
		Protocol:   horizonapi.ProtocolTCP,
		Name:       fmt.Sprintf("port-%s", name),
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	service.AddLabels(map[string]string{"component": p.names["perceptor"], "app": "opssight", "name": p.opssight.Name})
	service.AddSelectors(map[string]string{"component": p.names["perceptor"], "app": "opssight", "name": p.opssight.Name})

	return service, nil
}

// PerceptorLoadBalancerService creates a loadbalancer service for perceptor
func (p *SpecConfig) PerceptorLoadBalancerService() (*components.Service, error) {
	name := fmt.Sprintf("%s-exposed", p.names["perceptor"])
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      util.GetResourceName(p.opssight.Name, util.OpsSightName, name),
		Namespace: p.opssight.Spec.Namespace,
		Type:      horizonapi.ServiceTypeLoadBalancer,
	})

	err := service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(3001),
		TargetPort: fmt.Sprintf("%d", 3001),
		Protocol:   horizonapi.ProtocolTCP,
		Name:       fmt.Sprintf("port-%s", name),
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	service.AddLabels(map[string]string{"component": p.names["perceptor"], "app": "opssight", "name": p.opssight.Name})
	service.AddSelectors(map[string]string{"component": p.names["perceptor"], "app": "opssight", "name": p.opssight.Name})

	return service, nil
}

// PerceptorSecret create a secret for perceptor
func (p *SpecConfig) PerceptorSecret() (*components.Secret, error) {
	secretConfig := horizonapi.SecretConfig{
		Name:      util.GetResourceName(p.opssight.Name, util.OpsSightName, "blackduck"),
		Namespace: p.opssight.Spec.Namespace,
		Type:      horizonapi.SecretTypeOpaque,
	}
	secret := components.NewSecret(secretConfig)

	// empty data fields that will be overwritten
	emptyHosts := make(map[string]*opssightapi.Host)
	bytes, err := json.Marshal(emptyHosts)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal OpsSight's Host struct: %+v", err)
	}
	secret.AddData(map[string][]byte{p.opssight.Spec.Blackduck.ConnectionsEnvironmentVariableName: bytes})

	emptySecuredRegistries := make(map[string]*opssightapi.RegistryAuth)
	bytes, err = json.Marshal(emptySecuredRegistries)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal secured registries struct: %+v", err)
	}
	secret.AddData(map[string][]byte{"securedRegistries.json": bytes})

	secret.AddLabels(map[string]string{"component": "blackduck", "app": "opssight", "name": p.opssight.Name})
	return secret, nil
}

// GetPerceptorOpenShiftRoute creates the OpenShift route component for the perceptor model
func (p *SpecConfig) GetPerceptorOpenShiftRoute() *api.Route {
	namespace := p.opssight.Spec.Namespace
	if strings.ToUpper(p.opssight.Spec.Perceptor.Expose) == util.OPENSHIFT {
		return &api.Route{
			Name:               util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["perceptor"]),
			Namespace:          namespace,
			Kind:               "Service",
			ServiceName:        util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["perceptor"]),
			PortName:           fmt.Sprintf("port-%s", p.names["perceptor"]),
			Labels:             map[string]string{"app": "opssight", "name": p.opssight.Name, "component": fmt.Sprintf("%s-ui", p.names["perceptor"])},
			TLSTerminationType: routev1.TLSTerminationEdge,
		}
	}
	return nil
}
