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

	alertv1 "github.com/blackducksoftware/perceptor-protoform/pkg/api/alert/v1"
	hubv1 "github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
	opssightv1 "github.com/blackducksoftware/perceptor-protoform/pkg/api/opssight/v1"
	"github.com/blackducksoftware/perceptor-protoform/pkg/crds/alert"
	"github.com/blackducksoftware/perceptor-protoform/pkg/crds/hub"
	"github.com/blackducksoftware/perceptor-protoform/pkg/crds/opssight"
	"github.com/blackducksoftware/perceptor-protoform/pkg/protoform"
)

func main() {
	configPath := os.Args[1]
	fmt.Printf("Config path: %s", configPath)
	runProtoform(configPath)
}

func runProtoform(configPath string) {
	deployer, err := protoform.NewController(configPath)
	if err != nil {
		panic(err)
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	alertController, err := alert.NewController(&alert.Config{
		Config:        deployer.Config,
		KubeConfig:    deployer.KubeConfig,
		KubeClientSet: deployer.KubeClientSet,
		Defaults:      GetAlertDefaultValue(),
		Threadiness:   deployer.Config.Threadiness,
		StopCh:        stopCh,
	})
	deployer.AddController(alertController)

	hubController, err := hub.NewController(&hub.ProtoformConfig{
		Config:        deployer.Config,
		KubeConfig:    deployer.KubeConfig,
		KubeClientSet: deployer.KubeClientSet,
		Defaults:      GetHubDefaultValue(),
		Threadiness:   deployer.Config.Threadiness,
		StopCh:        stopCh,
	})
	deployer.AddController(hubController)

	opssSightController, err := opssight.NewController(&opssight.Config{
		Config:        deployer.Config,
		KubeConfig:    deployer.KubeConfig,
		KubeClientSet: deployer.KubeClientSet,
		Defaults:      GetOpsSightDefaultValue(),
		Threadiness:   deployer.Config.Threadiness,
		StopCh:        stopCh,
	})
	deployer.AddController(opssSightController)

	deployer.Deploy()

	<-stopCh

}

// GetAlertDefaultValue creates a alert crd configuration object with defaults
func GetAlertDefaultValue() *alertv1.AlertSpec {
	port := 8443
	hubPort := 443
	standAlone := true

	return &alertv1.AlertSpec{
		Port:           &port,
		HubPort:        &hubPort,
		StandAlone:     &standAlone,
		AlertMemory:    "512M",
		CfsslMemory:    "640M",
		AlertImageName: "blackduck-alert",
		CfsslImageName: "hub-cfssl",
	}
}

// GetHubDefaultValue creates a hub crd configuration object with defaults
func GetHubDefaultValue() *hubv1.HubSpec {
	return &hubv1.HubSpec{
		Flavor:          "small",
		DockerRegistry:  "docker.io",
		DockerRepo:      "blackducksoftware",
		HubVersion:      "4.8.2",
		DbPrototype:     "empty",
		CertificateName: "default",
		HubType:         "worker",
		Environs:        []hubv1.Environs{},
		ImagePrefix:     "hub",
	}
}

// GetOpsSightDefaultValue creates a perceptor crd configuration object with defaults
func GetOpsSightDefaultValue() *opssightv1.OpsSightSpec {
	defaultPerceptorPort := 3001
	defaultPerceiverPort := 3002
	defaultScannerPort := 3003
	defaultIFPort := 3004
	defaultSkyfirePort := 3005
	defaultAnnotationInterval := 30
	defaultDumpInterval := 30
	defaultHubPort := 443
	defaultPerceptorHubClientTimeout := 100000
	defaultScannerHubClientTimeout := 600
	defaultScanLimit := 2
	defaultTotalScanLimit := 1000
	defaultCheckForStalledScansPauseHours := 999999
	defaultStalledScanClientTimeoutHours := 999999
	defaultModelMetricsPauseSeconds := 15
	defaultUnknownImagePauseMilliseconds := 15000
	defaultPodPerceiverEnabled := true
	defaultImagePerceiverEnabled := false
	defaultMetricsEnabled := true
	defaultPerceptorSkyfire := false
	defaultUseMockMode := false

	return &opssightv1.OpsSightSpec{
		PerceptorPort:             &defaultPerceptorPort,
		PerceiverPort:             &defaultPerceiverPort,
		ScannerPort:               &defaultScannerPort,
		ImageFacadePort:           &defaultIFPort,
		SkyfirePort:               &defaultSkyfirePort,
		InternalRegistries:        []opssightv1.RegistryAuth{},
		AnnotationIntervalSeconds: &defaultAnnotationInterval,
		DumpIntervalMinutes:       &defaultDumpInterval,
		HubUser:                   "sysadmin",
		HubPort:                   &defaultHubPort,
		HubClientTimeoutPerceptorMilliseconds: &defaultPerceptorHubClientTimeout,
		HubClientTimeoutScannerSeconds:        &defaultScannerHubClientTimeout,
		ConcurrentScanLimit:                   &defaultScanLimit,
		TotalScanLimit:                        &defaultTotalScanLimit,
		CheckForStalledScansPauseHours:        &defaultCheckForStalledScansPauseHours,
		StalledScanClientTimeoutHours:         &defaultStalledScanClientTimeoutHours,
		ModelMetricsPauseSeconds:              &defaultModelMetricsPauseSeconds,
		UnknownImagePauseMilliseconds:         &defaultUnknownImagePauseMilliseconds,
		DefaultVersion:                        "latest",
		Registry:                              "docker.io",
		ImagePath:                             "blackducksoftware",
		PerceptorImageName:                    "opssight-core",
		ScannerImageName:                      "opssight-scanner",
		ImagePerceiverImageName:               "opssight-image-processor",
		PodPerceiverImageName:                 "opssight-pod-processor",
		ImageFacadeImageName:                  "opssight-image-getter",
		SkyfireImageName:                      "skyfire",
		PodPerceiver:                          &defaultPodPerceiverEnabled,
		ImagePerceiver:                        &defaultImagePerceiverEnabled,
		Metrics:                               &defaultMetricsEnabled,
		PerceptorSkyfire:                      &defaultPerceptorSkyfire,
		DefaultCPU:                            "300m",
		DefaultMem:                            "1300Mi",
		LogLevel:                              "debug",
		HubUserPasswordEnvVar:                 "PCP_HUBUSERPASSWORD",
		SecretName:                            "blackduck-secret",
		UseMockMode:                           &defaultUseMockMode,
		ServiceAccounts: map[string]string{
			// WARNING: These service accounts need to exist !
			"pod-perceiver":          "opssight-processor",
			"image-perceiver":        "opssight-processor",
			"perceptor-image-facade": "opssight-scanner",
			"skyfire":                "skyfire",
		},
		ContainerNames: map[string]string{
			"perceiver":              "opssight-processor",
			"pod-perceiver":          "opssight-pod-processor",
			"image-perceiver":        "opssight-image-processor",
			"perceptor":              "opssight-core",
			"perceptor-image-facade": "opssight-image-getter",
			"perceptor-scanner":      "opssight-scanner",
			"skyfire":                "skyfire",
		},
	}
}
