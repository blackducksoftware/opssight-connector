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

package alert

import (
	"fmt"

	"github.com/blackducksoftware/perceptor-protoform/pkg/api"
	"github.com/blackducksoftware/perceptor-protoform/pkg/api/alert/v1"
)

// SpecConfig will contain the specification of Alert
type SpecConfig struct {
	config *v1.AlertSpec
}

// NewAlert will create the Alert object
func NewAlert(config *v1.AlertSpec) *SpecConfig {
	return &SpecConfig{config: config}
}

// GetComponents will return the list of components for alert
func (a *SpecConfig) GetComponents() (*api.ComponentList, error) {
	components := &api.ComponentList{}

	// Add alert
	dep, err := a.alertDeployment()
	if err != nil {
		return nil, fmt.Errorf("failed to create alert deployment: %v", err)
	}
	components.Deployments = append(components.Deployments, dep)
	components.Services = append(components.Services, a.alertService())
	components.Services = append(components.Services, a.alertExposedService())
	components.ConfigMaps = append(components.ConfigMaps, a.alertConfigMap())

	// Add cfssl if running in stand alone mode
	if *a.config.StandAlone {
		dep, err := a.cfsslDeployment()
		if err != nil {
			return nil, fmt.Errorf("failed to create cfssl deployment: %v", err)
		}
		components.Deployments = append(components.Deployments, dep)
		components.Services = append(components.Services, a.cfsslService())
	}

	return components, nil
}
