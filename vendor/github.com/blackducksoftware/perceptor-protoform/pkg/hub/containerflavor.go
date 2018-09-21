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

package hub

import (
	"strings"

	"github.com/blackducksoftware/perceptor-protoform/pkg/util"
)

// Flavor will determine the size of the Black Duck Hub
type Flavor string

const (
	// SMALL Black Duck Hub
	SMALL Flavor = "SMALL"
	// MEDIUM Black Duck Hub
	MEDIUM Flavor = "MEDIUM"
	// LARGE Black Duck Hub
	LARGE Flavor = "LARGE"
	// OPSSIGHT Black Duck Hub
	OPSSIGHT Flavor = "OPSSIGHT"
)

// ContainerFlavor configuration will have the settings for flavored Black Duck Hub
type ContainerFlavor struct {
	WebserverMemoryLimit       string
	SolrMemoryLimit            string
	WebappCPULimit             string
	WebappMemoryLimit          string
	WebappHubMaxMemory         string
	ScanReplicas               *int32
	ScanMemoryLimit            string
	ScanHubMaxMemory           string
	JobRunnerReplicas          *int32
	JobRunnerMemoryLimit       string
	JobRunnerHubMaxMemory      string
	CfsslMemoryLimit           string
	LogstashMemoryLimit        string
	RegistrationMemoryLimit    string
	ZookeeperMemoryLimit       string
	AuthenticationMemoryLimit  string
	AuthenticationHubMaxMemory string
	DocumentationMemoryLimit   string
	PostgresMemoryLimit        string
	PostgresCPULimit           string
}

// GetContainersFlavor will return the default settings for the flavored Black Duck Hub
func GetContainersFlavor(flavor string) *ContainerFlavor {
	switch Flavor(strings.ToUpper(flavor)) {
	case SMALL:
		return &ContainerFlavor{
			WebserverMemoryLimit:       smallWebServerMemoryLimit,
			SolrMemoryLimit:            smallSolrMemoryLimit,
			WebappCPULimit:             smallWebappCPULimit,
			WebappMemoryLimit:          smallWebappMemoryLimit,
			WebappHubMaxMemory:         smallWebappHubMaxMemory,
			ScanReplicas:               util.IntToInt32(smallScanReplicas),
			ScanMemoryLimit:            smallScanMemoryLimit,
			ScanHubMaxMemory:           smallScanHubMaxMemory,
			JobRunnerReplicas:          util.IntToInt32(smallJobRunnerReplicas),
			JobRunnerMemoryLimit:       smallJobRunnerMemoryLimit,
			JobRunnerHubMaxMemory:      smallJobRunnerHubMaxMemory,
			CfsslMemoryLimit:           cfsslMemoryLimit,
			LogstashMemoryLimit:        logstashMemoryLimit,
			RegistrationMemoryLimit:    registrationMemoryLimit,
			ZookeeperMemoryLimit:       zookeeperMemoryLimit,
			AuthenticationMemoryLimit:  authenticationMemoryLimit,
			AuthenticationHubMaxMemory: authenticationHubMaxMemory,
			DocumentationMemoryLimit:   documentationMemoryLimit,
			PostgresCPULimit:           smallPostgresCPULimit,
			PostgresMemoryLimit:        smallPostgresMemoryLimit,
		}
	case MEDIUM:
		return &ContainerFlavor{
			WebserverMemoryLimit:       mediumWebServerMemoryLimit,
			SolrMemoryLimit:            mediumSolrMemoryLimit,
			WebappCPULimit:             mediumWebappCPULimit,
			WebappMemoryLimit:          mediumWebappMemoryLimit,
			WebappHubMaxMemory:         mediumWebappHubMaxMemory,
			ScanReplicas:               util.IntToInt32(mediumScanReplicas),
			ScanMemoryLimit:            mediumScanMemoryLimit,
			ScanHubMaxMemory:           mediumScanHubMaxMemory,
			JobRunnerReplicas:          util.IntToInt32(mediumJobRunnerReplicas),
			JobRunnerMemoryLimit:       mediumJobRunnerMemoryLimit,
			JobRunnerHubMaxMemory:      mediumJobRunnerHubMaxMemory,
			CfsslMemoryLimit:           cfsslMemoryLimit,
			LogstashMemoryLimit:        logstashMemoryLimit,
			RegistrationMemoryLimit:    registrationMemoryLimit,
			ZookeeperMemoryLimit:       zookeeperMemoryLimit,
			AuthenticationMemoryLimit:  authenticationMemoryLimit,
			AuthenticationHubMaxMemory: authenticationHubMaxMemory,
			DocumentationMemoryLimit:   documentationMemoryLimit,
			PostgresCPULimit:           mediumPostgresCPULimit,
			PostgresMemoryLimit:        mediumPostgresMemoryLimit,
		}
	case LARGE:
		return &ContainerFlavor{
			WebserverMemoryLimit:       largeWebServerMemoryLimit,
			SolrMemoryLimit:            largeSolrMemoryLimit,
			WebappCPULimit:             largeWebappCPULimit,
			WebappMemoryLimit:          largeWebappMemoryLimit,
			WebappHubMaxMemory:         largeWebappHubMaxMemory,
			ScanReplicas:               util.IntToInt32(largeScanReplicas),
			ScanMemoryLimit:            largeScanMemoryLimit,
			ScanHubMaxMemory:           largeScanHubMaxMemory,
			JobRunnerReplicas:          util.IntToInt32(largeJobRunnerReplicas),
			JobRunnerMemoryLimit:       largeJobRunnerMemoryLimit,
			JobRunnerHubMaxMemory:      largeJobRunnerHubMaxMemory,
			CfsslMemoryLimit:           cfsslMemoryLimit,
			LogstashMemoryLimit:        logstashMemoryLimit,
			RegistrationMemoryLimit:    registrationMemoryLimit,
			ZookeeperMemoryLimit:       zookeeperMemoryLimit,
			AuthenticationMemoryLimit:  authenticationMemoryLimit,
			AuthenticationHubMaxMemory: authenticationHubMaxMemory,
			DocumentationMemoryLimit:   documentationMemoryLimit,
			PostgresCPULimit:           largePostgresCPULimit,
			PostgresMemoryLimit:        largePostgresMemoryLimit,
		}
	case OPSSIGHT:
		return &ContainerFlavor{
			WebserverMemoryLimit:       opsSightWebServerMemoryLimit,
			SolrMemoryLimit:            opsSightSolrMemoryLimit,
			WebappCPULimit:             opsSightWebappCPULimit,
			WebappMemoryLimit:          opsSightWebappMemoryLimit,
			WebappHubMaxMemory:         opsSightWebappHubMaxMemory,
			ScanReplicas:               util.IntToInt32(opsSightScanReplicas),
			ScanMemoryLimit:            opsSightScanMemoryLimit,
			ScanHubMaxMemory:           opsSightScanHubMaxMemory,
			JobRunnerReplicas:          util.IntToInt32(opsSightJobRunnerReplicas),
			JobRunnerMemoryLimit:       opsSightJobRunnerMemoryLimit,
			JobRunnerHubMaxMemory:      opsSightJobRunnerHubMaxMemory,
			CfsslMemoryLimit:           cfsslMemoryLimit,
			LogstashMemoryLimit:        logstashMemoryLimit,
			RegistrationMemoryLimit:    registrationMemoryLimit,
			ZookeeperMemoryLimit:       zookeeperMemoryLimit,
			AuthenticationMemoryLimit:  authenticationMemoryLimit,
			AuthenticationHubMaxMemory: authenticationHubMaxMemory,
			DocumentationMemoryLimit:   documentationMemoryLimit,
			PostgresCPULimit:           opsSightPostgresCPULimit,
			PostgresMemoryLimit:        opsSightPostgresMemoryLimit,
		}
	default:
		return nil
	}
}
