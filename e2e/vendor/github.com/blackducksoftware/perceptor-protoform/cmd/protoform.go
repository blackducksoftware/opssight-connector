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

package main

import (
	"fmt"
	"os"

	"github.com/blackducksoftware/perceptor-protoform/pkg/api"
	"github.com/blackducksoftware/perceptor-protoform/pkg/protoform"
)

func main() {
	configPath := os.Args[1]
	fmt.Printf("Config path: %s", configPath)
	runProtoform(configPath)
}

func runProtoform(configPath string) {
	defaults := createDefaults()
	installer := protoform.NewInstaller(defaults, configPath)
	installer.Run()
}

func createDefaults() *api.ProtoformDefaults {
	d := protoform.NewDefaultsObj()
	d.HubUser = "sysadmin"
	d.HubHost = "nginx-webapp-logstash"
	d.HubPort = 8443
	d.InternalDockerRegistries = []string{"docker-registry.default.svc:5000", "172.1.1.0:5000"}
	d.DefaultVersion = "master"
	d.Registry = "gcr.io"
	d.ImagePath = "gke-verification/blackducksoftware"
	d.Namespace = "perceptor"
	d.LogLevel = "info"
	d.DefaultCPU = "300m"
	d.DefaultMem = "1300Mi"
	d.HubClientTimeoutPerceptorSeconds = 5
	d.HubClientTimeoutScannerSeconds = 120
	d.ConcurrentScanLimit = 7

	return d
}
