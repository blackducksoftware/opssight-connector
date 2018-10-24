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

const (
	gcloudAuthFilePath = "/var/secrets/google/key.json"

	// SHARED VALUES
	cfsslMemoryLimit           = "1G"
	logstashMemoryLimit        = "1G"
	registrationMemoryLimit    = "640M"
	zookeeperMemoryLimit       = "640M"
	authenticationMemoryLimit  = "1024M"
	authenticationHubMaxMemory = "512m"
	documentationMemoryLimit   = "512M"

	registrationMinCPUUsage = "1"
	zookeeperMinCPUUsage    = "1"
	jonRunnerMinCPUUsage    = "1"
	jonRunnerMaxCPUUsage    = "1"

	// Ports
	cfsslPort          = "8888"
	webserverPort      = "8443"
	documentationPort  = "8443"
	solrPort           = "8983"
	registrationPort   = "8443"
	zookeeperPort      = "2181"
	jobRunnerPort      = "3001"
	scannerPort        = "8443"
	authenticationPort = "8443"
	webappPort         = "8443"
	logstashPort       = "5044"
	postgresPort       = "5432"

	// Small Flavor
	smallWebServerMemoryLimit = "1024M"

	smallSolrMemoryLimit = "1024M"

	smallWebappCPULimit     = "2"
	smallWebappMemoryLimit  = "5120M"
	smallWebappHubMaxMemory = "4096m"

	smallScanReplicas     = 1
	smallScanMemoryLimit  = "5120M"
	smallScanHubMaxMemory = "4096m"

	smallJobRunnerReplicas     = 1
	smallJobRunnerMemoryLimit  = "7168M"
	smallJobRunnerHubMaxMemory = "6144m"

	smallPostgresCPULimit    = "1"
	smallPostgresMemoryLimit = "4096M"

	// Medium Flavor
	mediumWebServerMemoryLimit = "2048M"

	mediumSolrMemoryLimit = "1024M"

	mediumWebappCPULimit     = "2"
	mediumWebappMemoryLimit  = "5120M"
	mediumWebappHubMaxMemory = "4096m"

	mediumScanReplicas     = 2
	mediumScanMemoryLimit  = "5120M"
	mediumScanHubMaxMemory = "4096m"

	mediumJobRunnerReplicas     = 4
	mediumJobRunnerMemoryLimit  = "7168M"
	mediumJobRunnerHubMaxMemory = "6144m"

	mediumPostgresCPULimit    = "2"
	mediumPostgresMemoryLimit = "8192M"

	// Large Flavor
	largeWebServerMemoryLimit = "2048M"

	largeSolrMemoryLimit = "1024M"

	largeWebappCPULimit     = "1"
	largeWebappMemoryLimit  = "9728M"
	largeWebappHubMaxMemory = "8192m"

	largeScanReplicas     = 3
	largeScanMemoryLimit  = "9728M"
	largeScanHubMaxMemory = "8192m"

	largeJobRunnerReplicas     = 6
	largeJobRunnerMemoryLimit  = "13824M"
	largeJobRunnerHubMaxMemory = "12288m"

	largePostgresCPULimit    = "2"
	largePostgresMemoryLimit = "12288M"

	// OpsSight Flavor
	opsSightWebServerMemoryLimit = "2048M"

	opsSightSolrMemoryLimit = "1024M"

	opsSightWebappCPULimit     = "3"
	opsSightWebappMemoryLimit  = "19728M"
	opsSightWebappHubMaxMemory = "8192m"

	opsSightScanReplicas     = 5
	opsSightScanMemoryLimit  = "9728M"
	opsSightScanHubMaxMemory = "8192m"

	opsSightJobRunnerReplicas     = 10
	opsSightJobRunnerMemoryLimit  = "13824M"
	opsSightJobRunnerHubMaxMemory = "12288m"

	opsSightPostgresCPULimit    = "3"
	opsSightPostgresMemoryLimit = "12288M"
)
