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

package util

import (
	alertv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	blackduckv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	//samplev1 "github.com/blackducksoftware/synopsys-operator/pkg/api/sample/v1"
)

// GetSampleDefaultValue creates a sample crd configuration object with defaults
//func GetSampleDefaultValue() *samplev1.SampleSpec {
//	return &samplev1.SampleSpec{
//		Namespace:   "namesapce",
//		SampleValue: "Value",
//	}
//}

// GetHubDefaultValue creates a hub crd configuration object with defaults
func GetHubDefaultValue() *blackduckv1.BlackduckSpec {
	return &blackduckv1.BlackduckSpec{
		Size:            "Small",
		DbPrototype:     "",
		CertificateName: "default",
		Type:            "",
	}
}

// GetHubDefaultPersistentStorage creates a hub crd configuration object
// with defaults and persistent storage
func GetHubDefaultPersistentStorage() *blackduckv1.BlackduckSpec {
	return &blackduckv1.BlackduckSpec{
		Namespace:         "synopsys-operator",
		Size:              "small",
		PVCStorageClass:   "",
		LivenessProbes:    false,
		PersistentStorage: true,
		PVC: []blackduckv1.PVC{
			{
				Name: "blackduck-postgres",
				Size: "200Gi",
			},
			{
				Name: "blackduck-authentication",
				Size: "2Gi",
			},
			{
				Name: "blackduck-cfssl",
				Size: "2Gi",
			},
			{
				Name: "blackduck-registration",
				Size: "2Gi",
			},
			{
				Name: "blackduck-solr",
				Size: "2Gi",
			},
			{
				Name: "blackduck-webapp",
				Size: "2Gi",
			},
			{
				Name: "blackduck-logstash",
				Size: "20Gi",
			},
			{
				Name: "blackduck-zookeeper-data",
				Size: "2Gi",
			},
			{
				Name: "blackduck-zookeeper-datalog",
				Size: "2Gi",
			},
		},
		CertificateName: "default",
		Type:            "Artifacts",
		Environs:        []string{},
		ImageRegistries: []string{},
		LicenseKey:      "",
	}
}

// GetOpsSightDefaultValue creates a perceptor crd configuration object with defaults
func GetOpsSightDefaultValue() *opssightv1.OpsSightSpec {
	return &opssightv1.OpsSightSpec{
		Perceptor: &opssightv1.Perceptor{
			Name:                           "perceptor",
			Port:                           3001,
			Image:                          "gcr.io/saas-hub-stg/blackducksoftware/perceptor:master",
			CheckForStalledScansPauseHours: 999999,
			StalledScanClientTimeoutHours:  999999,
			ModelMetricsPauseSeconds:       15,
			UnknownImagePauseMilliseconds:  15000,
			ClientTimeoutMilliseconds:      100000,
		},
		Perceiver: &opssightv1.Perceiver{
			EnableImagePerceiver: false,
			EnablePodPerceiver:   true,
			Port:                 3002,
			ImagePerceiver: &opssightv1.ImagePerceiver{
				Name:  "image-perceiver",
				Image: "gcr.io/saas-hub-stg/blackducksoftware/image-perceiver:master",
			},
			PodPerceiver: &opssightv1.PodPerceiver{
				Name:  "pod-perceiver",
				Image: "gcr.io/saas-hub-stg/blackducksoftware/pod-perceiver:master",
			},
			ServiceAccount:            "perceiver",
			AnnotationIntervalSeconds: 30,
			DumpIntervalMinutes:       30,
		},
		ScannerPod: &opssightv1.ScannerPod{
			Name: "perceptor-scanner",
			ImageFacade: &opssightv1.ImageFacade{
				Port:           3004,
				Image:          "gcr.io/saas-hub-stg/blackducksoftware/perceptor-imagefacade:master",
				ServiceAccount: "perceptor-scanner",
				Name:           "perceptor-imagefacade",
			},
			Scanner: &opssightv1.Scanner{
				Name:                 "perceptor-scanner",
				Port:                 3003,
				Image:                "gcr.io/saas-hub-stg/blackducksoftware/perceptor-scanner:master",
				ClientTimeoutSeconds: 600,
			},
			ReplicaCount:   1,
			ImageDirectory: "/var/images",
		},
		Prometheus: &opssightv1.Prometheus{
			Name:  "prometheus",
			Image: "docker.io/prom/prometheus:v2.1.0",
			Port:  9090,
		},
		Skyfire: &opssightv1.Skyfire{
			Image:                        "gcr.io/saas-hub-stg/blackducksoftware/pyfire:master",
			Name:                         "skyfire",
			Port:                         3005,
			PrometheusPort:               3006,
			ServiceAccount:               "skyfire",
			HubClientTimeoutSeconds:      100,
			HubDumpPauseSeconds:          240,
			KubeDumpIntervalSeconds:      60,
			PerceptorDumpIntervalSeconds: 60,
		},
		Blackduck: &opssightv1.Blackduck{
			InitialCount:                       0,
			MaxCount:                           0,
			ConnectionsEnvironmentVariableName: "blackduck.json",
			TLSVerification:                    false,
			DeleteBlackduckThresholdPercentage: 50,
			BlackduckSpec:                      GetHubDefaultValue(),
		},
		EnableMetrics: true,
		EnableSkyfire: false,
		DefaultCPU:    "300m",
		DefaultMem:    "1300Mi",
		LogLevel:      "debug",
		SecretName:    "perceptor",
		ConfigMapName: "opssight",
		DesiredState:  "START",
	}
}

// GetOpsSightDefaultValueWithDisabledHub creates an opssight crd configuration
// with defaults and a Disabled Hub
func GetOpsSightDefaultValueWithDisabledHub() *opssightv1.OpsSightSpec {
	blackduck := opssightv1.Blackduck{
		ExternalHosts:                      []*opssightv1.Host{},
		ConnectionsEnvironmentVariableName: "",
		TLSVerification:                    true,
		InitialCount:                       0,
		MaxCount:                           0,
		BlackduckSpec: &blackduckv1.BlackduckSpec{
			Size:              "small",
			DbPrototype:       "",
			PersistentStorage: false,
			CertificateName:   "default",
			Type:              "worker",
			Environs: []string{
				"HUB_VERSION:2018.12.2",
			},
			ImageRegistries: []string{
				"docker.io/blackducksoftware/blackduck-authentication:2018.12.2",
				"docker.io/blackducksoftware/blackduck-documentation:2018.12.2",
				"docker.io/blackducksoftware/blackduck-jobrunner:2018.12.2",
				"docker.io/blackducksoftware/blackduck-registration:2018.12.2",
				"docker.io/blackducksoftware/blackduck-scan:2018.12.2",
				"docker.io/blackducksoftware/blackduck-webapp:2018.12.2",
				"docker.io/blackducksoftware/blackduck-cfssl:1.0.0",
				"docker.io/blackducksoftware/blackduck-logstash:1.0.2",
				"docker.io/blackducksoftware/blackduck-nginx:1.0.0",
				"docker.io/blackducksoftware/blackduck-solr:1.0.0",
				"docker.io/blackducksoftware/blackduck-zookeeper:1.0.0",
			},
			LicenseKey: "LICENSE_KEY",
		},
	}

	return &opssightv1.OpsSightSpec{
		Namespace: "opssight-test",
		Perceptor: &opssightv1.Perceptor{
			Name:                           "opssight-core",
			Image:                          "docker.io/blackducksoftware/opssight-core:master",
			Port:                           3001,
			CheckForStalledScansPauseHours: 999999,
			StalledScanClientTimeoutHours:  999999,
			ModelMetricsPauseSeconds:       15,
			UnknownImagePauseMilliseconds:  15000,
			ClientTimeoutMilliseconds:      100000,
		},
		ScannerPod: &opssightv1.ScannerPod{
			Name: "opssight-scanner",
			Scanner: &opssightv1.Scanner{
				Name:                 "opssight-scanner",
				Image:                "docker.io/blackducksoftware/opssight-scanner:master",
				Port:                 3003,
				ClientTimeoutSeconds: 600,
			},
			ImageFacade: &opssightv1.ImageFacade{
				Name:               "opssight-image-getter",
				Image:              "docker.io/blackducksoftware/opssight-image-getter:master",
				Port:               3004,
				InternalRegistries: []*opssightv1.RegistryAuth{},
				ImagePullerType:    "skopeo",
				ServiceAccount:     "opssight-scanner",
			},
			ReplicaCount: 1,
		},
		Perceiver: &opssightv1.Perceiver{
			EnableImagePerceiver: false,
			EnablePodPerceiver:   true,
			ImagePerceiver: &opssightv1.ImagePerceiver{
				Name:  "opssight-image-processor",
				Image: "docker.io/blackducksoftware/opssight-image-processor:master",
			},
			PodPerceiver: &opssightv1.PodPerceiver{
				Name:  "opssight-pod-processor",
				Image: "docker.io/blackducksoftware/opssight-pod-processor:master",
			},
			AnnotationIntervalSeconds: 30,
			DumpIntervalMinutes:       30,
			ServiceAccount:            "opssight-processor",
			Port:                      3002,
		},
		Prometheus: &opssightv1.Prometheus{
			Name:  "prometheus",
			Image: "docker.io/prom/prometheus:v2.1.0",
			Port:  9090,
		},
		EnableSkyfire: false,
		Skyfire: &opssightv1.Skyfire{
			Name:                         "skyfire",
			Image:                        "gcr.io/saas-hub-stg/blackducksoftware/pyfire:master",
			Port:                         3005,
			PrometheusPort:               3006,
			ServiceAccount:               "skyfire",
			HubClientTimeoutSeconds:      120,
			HubDumpPauseSeconds:          240,
			KubeDumpIntervalSeconds:      60,
			PerceptorDumpIntervalSeconds: 60,
		},
		Blackduck:     &blackduck,
		EnableMetrics: true,
		DefaultCPU:    "300m",
		DefaultMem:    "1300Mi",
		LogLevel:      "debug",
		SecretName:    "blackduck",
		DesiredState:  "START",
	}
}

// GetAlertDefaultValue creates an Alert crd configuration object with defaults
func GetAlertDefaultValue() *alertv1.AlertSpec {
	port := 8443
	standAlone := true

	return &alertv1.AlertSpec{
		Namespace:            "alert-test",
		Registry:             "docker.io",
		ImagePath:            "blackducksoftware",
		AlertImageName:       "blackduck-alert",
		AlertImageVersion:    "3.1.0",
		CfsslImageName:       "blackduck-cfssl",
		CfsslImageVersion:    "1.0.0",
		ExposeService:        "NODEPORT",
		Port:                 &port,
		EncryptionPassword:   "",
		EncryptionGlobalSalt: "",
		PersistentStorage:    true,
		PVCName:              "alert-pvc",
		StandAlone:           &standAlone,
		PVCSize:              "5G",
		PVCStorageClass:      "",
		AlertMemory:          "2560M",
		CfsslMemory:          "640M",
		Environs: []string{
			"ALERT_SERVER_PORT:8443",
			"PUBLIC_HUB_WEBSERVER_HOST:localhost",
			"PUBLIC_HUB_WEBSERVER_PORT:443",
		},
	}
}
