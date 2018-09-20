/*
Copyright (C) 2018 Synopsys, Inc.

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

package opssight

import (
	"fmt"

	"github.com/blackducksoftware/perceptor-protoform/pkg/api"
	"github.com/blackducksoftware/perceptor-protoform/pkg/api/opssight/v1"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
)

// SpecConfig will contain the specification of OpsSight
type SpecConfig struct {
	config *v1.OpsSightSpec
}

// NewSpecConfig will create the OpsSight object
func NewSpecConfig(config *v1.OpsSightSpec) *SpecConfig {
	return &SpecConfig{config: config}
}

// GetComponents will return the list of components
func (p *SpecConfig) GetComponents() (*api.ComponentList, error) {
	p.configServiceAccounts()
	err := p.sanityCheckServices()
	if err != nil {
		return nil, errors.Annotate(err, "Please set the service accounts correctly")
	}

	p.substituteDefaultImageVersion()

	components := &api.ComponentList{}

	// Add Perceptor
	rc, err := p.PerceptorReplicationController()
	if err != nil {
		return nil, errors.Trace(err)
	}
	components.ReplicationControllers = append(components.ReplicationControllers, rc)
	service, err := p.PerceptorService()
	if err != nil {
		return nil, errors.Trace(err)
	}
	components.Services = append(components.Services, service)
	cm, err := p.PerceptorConfigMap()
	if err != nil {
		return nil, errors.Trace(err)
	}
	components.ConfigMaps = append(components.ConfigMaps, cm)
	components.Secrets = append(components.Secrets, p.PerceptorSecret())

	// Add Perceptor Scanner
	scannerRC, err := p.ScannerReplicationController()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create scanner replication controller")
	}
	components.ReplicationControllers = append(components.ReplicationControllers, scannerRC)
	components.Services = append(components.Services, p.ScannerService(), p.ImageFacadeService())
	scannerCM, err := p.ScannerConfigMap()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create scanner replication controller")
	}
	ifCM, err := p.ImageFacadeConfigMap()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create scanner replication controller")
	}
	components.ConfigMaps = append(components.ConfigMaps, scannerCM, ifCM)
	log.Debugf("image facade configmap: %+v", ifCM.GetObj())
	components.ServiceAccounts = append(components.ServiceAccounts, p.ScannerServiceAccount())
	components.ClusterRoleBindings = append(components.ClusterRoleBindings, p.ScannerClusterRoleBinding())

	if p.config.PodPerceiver != nil && *p.config.PodPerceiver {
		rc, err := p.PodPerceiverReplicationController()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create pod perceiver")
		}
		components.ReplicationControllers = append(components.ReplicationControllers, rc)
		components.Services = append(components.Services, p.PodPerceiverService())
		perceiverConfigMap, err := p.PerceiverConfigMap()
		if err != nil {
			return nil, errors.Trace(err)
		}
		components.ConfigMaps = append(components.ConfigMaps, perceiverConfigMap)
		components.ServiceAccounts = append(components.ServiceAccounts, p.PodPerceiverServiceAccount())
		cr := p.PodPerceiverClusterRole()
		components.ClusterRoles = append(components.ClusterRoles, cr)
		components.ClusterRoleBindings = append(components.ClusterRoleBindings, p.PodPerceiverClusterRoleBinding(cr))
	}

	if p.config.ImagePerceiver != nil && *p.config.ImagePerceiver {
		rc, err := p.ImagePerceiverReplicationController()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create image perceiver")
		}
		components.ReplicationControllers = append(components.ReplicationControllers, rc)
		components.Services = append(components.Services, p.ImagePerceiverService())
		perceiverConfigMap, err := p.PerceiverConfigMap()
		if err != nil {
			return nil, errors.Trace(err)
		}
		components.ConfigMaps = append(components.ConfigMaps, perceiverConfigMap)
		components.ServiceAccounts = append(components.ServiceAccounts, p.ImagePerceiverServiceAccount())
		cr := p.ImagePerceiverClusterRole()
		components.ClusterRoles = append(components.ClusterRoles, cr)
		components.ClusterRoleBindings = append(components.ClusterRoleBindings, p.ImagePerceiverClusterRoleBinding(cr))
	}

	if p.config.PerceptorSkyfire != nil && *p.config.PerceptorSkyfire {
		rc, err := p.PerceptorSkyfireReplicationController()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create skyfire")
		}
		components.ReplicationControllers = append(components.ReplicationControllers, rc)
		components.Services = append(components.Services, p.PerceptorSkyfireService())
		components.ConfigMaps = append(components.ConfigMaps, p.PerceptorSkyfireConfigMap())
		components.ServiceAccounts = append(components.ServiceAccounts, p.PerceptorSkyfireServiceAccount())
		cr := p.PerceptorSkyfireClusterRole()
		components.ClusterRoles = append(components.ClusterRoles, cr)
		components.ClusterRoleBindings = append(components.ClusterRoleBindings, p.PerceptorSkyfireClusterRoleBinding(cr))
	}

	if p.config.Metrics != nil && *p.config.Metrics {
		dep, err := p.PerceptorMetricsDeployment()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create metrics")
		}
		components.Deployments = append(components.Deployments, dep)
		components.Services = append(components.Services, p.PerceptorMetricsService())
		components.ConfigMaps = append(components.ConfigMaps, p.PerceptorMetricsConfigMap())
	}

	return components, nil
}

func (p *SpecConfig) substituteDefaultImageVersion() {
	if len(p.config.PerceptorImageVersion) == 0 {
		p.config.PerceptorImageVersion = p.config.DefaultVersion
	}
	if len(p.config.ScannerImageVersion) == 0 {
		p.config.ScannerImageVersion = p.config.DefaultVersion
	}
	if len(p.config.PerceiverImageVersion) == 0 {
		p.config.PerceiverImageVersion = p.config.DefaultVersion
	}
	if len(p.config.ImageFacadeImageVersion) == 0 {
		p.config.ImageFacadeImageVersion = p.config.DefaultVersion
	}
	if len(p.config.SkyfireImageVersion) == 0 {
		p.config.SkyfireImageVersion = p.config.DefaultVersion
	}
}

func (p *SpecConfig) configServiceAccounts() {
	// TODO Viperize these env vars.
	if len(p.config.ServiceAccounts) == 0 {
		svcAccounts := map[string]string{
			// WARNING: These service accounts need to exist !
			"pod-perceiver":          "perceiver",
			"image-perceiver":        "perceiver",
			"perceptor-image-facade": "perceptor-scanner",
			"skyfire":                "skyfire",
		}
		p.config.ServiceAccounts = svcAccounts
	}
}

// TODO programatically validate rather then sanity check.
func (p *SpecConfig) sanityCheckServices() error {
	serviceAccountNames := map[string]bool{
		"perceptor":              true,
		"pod-perceiver":          true,
		"image-perceiver":        true,
		"perceptor-scanner":      true,
		"perceptor-image-facade": true,
		"skyfire":                true,
	}
	for cn := range p.config.ServiceAccounts {
		if _, ok := serviceAccountNames[cn]; !ok {
			return fmt.Errorf("invalid service account name <%s>", cn)
		}
	}
	return nil
}
