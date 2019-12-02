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

const (
	// AlertCRDName is the name of an Alert CRD
	AlertCRDName = "alerts.synopsys.com"
	// BlackDuckCRDName is the name of the Black Duck CRD
	BlackDuckCRDName = "blackducks.synopsys.com"
	// OpsSightCRDName is the name of an OpsSight CRD
	OpsSightCRDName = "opssights.synopsys.com"
	// PrmCRDName is the name of the Polaris Reporting Module CRD
	PrmCRDName = "prms.synopsys.com"

	// OperatorName is the name of an Operator
	OperatorName = "synopsys-operator"
	// AlertName is the name of an Alert app
	AlertName = "alert"
	// BlackDuckName is the name of the Black Duck app
	BlackDuckName = "blackduck"
	// OpsSightName is the name of an OpsSight app
	OpsSightName = "opssight"
	// PrmName is the name of the Prm app
	PrmName = "prm"
)

// GetSampleDefaultValue creates a sample crd configuration object with defaults
//func GetSampleDefaultValue() *samplev1.SampleSpec {
//	return &samplev1.SampleSpec{
//		Namespace:   "namesapce",
//		SampleValue: "Value",
//	}
//}

// GetBlackDuckTemplate returns the required fields for Black Duck
func GetBlackDuckTemplate() *blackduckv1.BlackduckSpec {
	return &blackduckv1.BlackduckSpec{
		Size:              "Small",
		CertificateName:   "default",
		PersistentStorage: false,
		ExposeService:     NONE,
	}
}

// GetBlackDuckDefaultPersistentStorageLatest creates a Black Duck crd configuration object
// with defaults and persistent storage
func GetBlackDuckDefaultPersistentStorageLatest() *blackduckv1.BlackduckSpec {
	return &blackduckv1.BlackduckSpec{
		Namespace:         "blackduck-pvc",
		Size:              "Small",
		CertificateName:   "default",
		LivenessProbes:    false,
		PersistentStorage: true,
		ExposeService:     NONE,
		Environs:          []string{},
		ImageRegistries:   []string{},
		PVC:               []blackduckv1.PVC{},
	}
}

// GetBlackDuckDefaultExternalPersistentStorageLatest creates a BlackDuck crd configuration object
// with defaults and external persistent storage for latest BlackDuck
func GetBlackDuckDefaultExternalPersistentStorageLatest() *blackduckv1.BlackduckSpec {
	return &blackduckv1.BlackduckSpec{
		Namespace:         "synopsys-operator",
		Size:              "small",
		LivenessProbes:    false,
		PersistentStorage: true,
		PVC:               []blackduckv1.PVC{},
		CertificateName:   "default",
		Type:              "Artifacts",
		ExposeService:     NONE,
		Environs:          []string{},
		ImageRegistries:   []string{},
	}
}

// GetBlackDuckDefaultPersistentStorageV1 creates a BlackDuck crd configuration object
// with defaults and persistent storage for V1 BlackDuck
func GetBlackDuckDefaultPersistentStorageV1() *blackduckv1.BlackduckSpec {
	return &blackduckv1.BlackduckSpec{
		Namespace:         "synopsys-operator",
		Size:              "small",
		LivenessProbes:    false,
		PersistentStorage: true,
		PVC:               []blackduckv1.PVC{},
		CertificateName:   "default",
		Type:              "Artifacts",
		ExposeService:     NONE,
		Environs:          []string{},
		ImageRegistries:   []string{},
	}
}

// GetBlackDuckDefaultExternalPersistentStorageV1 creates a BlackDuck crd configuration object
// with defaults and external persistent storage for V1 BlackDuck
func GetBlackDuckDefaultExternalPersistentStorageV1() *blackduckv1.BlackduckSpec {
	return &blackduckv1.BlackduckSpec{
		Namespace:         "synopsys-operator",
		Size:              "small",
		LivenessProbes:    false,
		PersistentStorage: true,
		PVC:               []blackduckv1.PVC{},
		Type:              "Artifacts",
		ExposeService:     NONE,
	}
}

// GetBlackDuckDefaultBDBA returns a BlackDuck with BDBA
func GetBlackDuckDefaultBDBA() *blackduckv1.BlackduckSpec {
	return &blackduckv1.BlackduckSpec{
		Namespace:       "blackduck-bdba",
		CertificateName: "default",
		Environs: []string{
			"USE_BINARY_UPLOADS:1",
		},
		LivenessProbes:    false,
		PersistentStorage: false,
		Size:              "small",
		ExposeService:     NONE,
	}
}

// GetBlackDuckDefaultEphemeral returns a BlackDuck with ephemeral storage
func GetBlackDuckDefaultEphemeral() *blackduckv1.BlackduckSpec {
	return &blackduckv1.BlackduckSpec{
		Namespace:         "blackduck-ephemeral",
		CertificateName:   "default",
		LivenessProbes:    false,
		PersistentStorage: false,
		Size:              "small",
		Type:              "worker",
		ExposeService:     NONE,
	}
}

// GetBlackDuckDefaultEphemeralCustomAuthCA returns a BlackDuck with ephemeral storage
// using custom auth CA
func GetBlackDuckDefaultEphemeralCustomAuthCA() *blackduckv1.BlackduckSpec {
	return &blackduckv1.BlackduckSpec{
		Namespace:         "blackduck-auth-ca",
		CertificateName:   "default",
		LivenessProbes:    false,
		PersistentStorage: false,
		Size:              "Small",
		ExposeService:     NONE,
		AuthCustomCA:      "-----BEGIN CERTIFICATE-----\r\nMIIE1DCCArwCCQCuw9TgaoBKVDANBgkqhkiG9w0BAQsFADAsMQswCQYDVQQGEwJV\r\nUzELMAkGA1UECgwCYmQxEDAOBgNVBAMMB1JPT1QgQ0EwHhcNMTkwMjA2MDAzMjM3\r\nWhcNMjExMTI2MDAzMjM3WjAsMQswCQYDVQQGEwJVUzELMAkGA1UECgwCYmQxEDAO\r\nBgNVBAMMB1JPT1QgQ0EwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQCr\r\nIctvPVoqRS3Ti38uFRVfJDovyi0p9PIaOmja3tMvkfecCsCVYHMo/vAy/fm9qiJI\r\nKutTwX9aLuiLO0tsDDUNwv0CrbXvuHpWvASOAdKyl6uxiYl0fq0cyBZSdKlsdDGk\r\nivENpN2gKHxDSUgAo74wUskfBrKvfKLhJhOmKCbN/NvxlsGMM5DgPgFGNegmw5r0\r\nZlDTXlWn3J/8C80dfGjT5hLr6Jtl0KTqxSREVTLT0fDk7bt9BHH/TCtNs9UwR1UI\r\nJVjjzW6pgS1DmGZ7Mfg2WBhhdDBuN0gxk/bcoiV2tfI0MLQyeVP+qWmdUXSNn9CT\r\nmpYdKezMfi5ieSy40fy23n+D1C+Xm5pnFErm3BwZYdN9gI633IBPQa0ELo28ZxhI\r\nIclGGyhUubZJ+ybNvGOIrgypTXYrZqvyWMV3qiMZb1EzpKdqAzGfsN1zmF+o4Rc3\r\ntBa2EF/lNSVCClUeFBA2UXvD/K9QA84cbLNJwpBZ9Bc6CZyvRTYGzXtAuZUVvNju\r\nMcWhsqXWzhVkChTyYicOdT8ZB+7/eC3tFyjAKSszIA5xuO8NtuIZBAc2AzRrkoE5\r\nCgHEUxNA3tbRUjYnH5HcgaQveFQtFwBWqIMxPeJixSLk2KYJSsWpTPC1x6s1IBLO\r\nITWhedDbtbs/FT9+cXd9K+/L+6UgR31oHaY/hYai1QIDAQABMA0GCSqGSIb3DQEB\r\nCwUAA4ICAQAz7aK5m9yPE/tTFQJfZRr35ug8ikBuGFvzb5s3fWYlQ1QbKUPBp9Q/\r\n1kUGJF2niOULUp5Gig6urz+E1m3wE5jgYRwZjgTmoEQEmN0/VQWTus72isWhTsZ5\r\nJKDSzcKGRJnHzO91gA3ZP1Cxoin5GX6w8eqEA2vh1hc7+GyKPTOsxu8hYMYI1yId\r\nfWAjqEUobLZZoijf+c3AqBVcf4tOpFMRTy4au3H+v7TNjc/fAeZUeAz7BswfqEV9\r\n0QNNTpezq5IS+pSPShRatL9k/BaE3MaF0Ossfnv3UPV80Yrup+9pRV8Lu6EXrdg5\r\n3L2+KK2Nz9A+iF2u9VqUw9lcJCIjgY+APf6Tf2AKQxNCA/pV1z0I8aQAlSLolgpx\r\nSMLwMecpjAcHPWF5ut3Re+8PfeyLGzeXCVyhZc9Aj9KaTNLRa/kb21KNVbcGGTu/\r\nuiGMEJXq1a1fKzMKTPnARz70XCS7nLJ7qEK3TuvrMhCqEEdFUf/S4yAmmWaEO9Fr\r\nUBk9ACW9UYBFtowqbJkbJm3KEXMMFP5cs33j/HEA1IkKDVT9Hi7NEK2/Y7e9afv7\r\no1UGNrGgU1rK8K+/2htOH9JhlPFWHQkk+wvGL6fFI7p+6TGes0KILN4WioOEKY0t\r\n0V1Zr8bejDW49cu1Awy443SrauhFLOInubZLA8S9ZvwTVIvpmTDjdQ==\r\n-----END CERTIFICATE-----",
	}
}

// GetBlackDuckDefaultExternalDB returns a BlackDuck with an external Data Base
func GetBlackDuckDefaultExternalDB() *blackduckv1.BlackduckSpec {
	return &blackduckv1.BlackduckSpec{
		Namespace:         "blackduck-externaldb",
		CertificateName:   "default",
		Size:              "small",
		PersistentStorage: false,
		ExternalPostgres: &blackduckv1.PostgresExternalDBConfig{
			PostgresHost:          "<<IP/FQDN>>",
			PostgresPort:          5432,
			PostgresAdmin:         "blackduck",
			PostgresUser:          "blackduck_user",
			PostgresSsl:           false,
			PostgresAdminPassword: "<<PASSWORD>>",
			PostgresUserPassword:  "<<PASSWORD>>",
		},
		Type:          "worker",
		ExposeService: NONE,
	}
}

// GetBlackDuckDefaultIPV6Disabled returns a BlackDuck with IPV6 Disabled
func GetBlackDuckDefaultIPV6Disabled() *blackduckv1.BlackduckSpec {
	return &blackduckv1.BlackduckSpec{
		Namespace:       "blackduck-ipv6disabled",
		CertificateName: "default",
		Environs: []string{
			"IPV4_ONLY:1",
			"BLACKDUCK_HUB_SERVER_ADDRESS:0.0.0.0",
		},
		Size:              "small",
		PersistentStorage: false,
		Type:              "worker",
		ExposeService:     NONE,
	}
}

// GetOpsSightUpstream returns the required fields for an upstream OpsSight
func GetOpsSightUpstream() *opssightv1.OpsSightSpec {
	return &opssightv1.OpsSightSpec{
		IsUpstream: true,
		Perceptor: &opssightv1.Perceptor{
			CheckForStalledScansPauseHours: 999999,
			StalledScanClientTimeoutHours:  999999,
			ModelMetricsPauseSeconds:       15,
			UnknownImagePauseMilliseconds:  15000,
			ClientTimeoutMilliseconds:      100000,
			Expose:                         NONE,
		},
		Perceiver: &opssightv1.Perceiver{
			EnableImagePerceiver:       false,
			EnableArtifactoryPerceiver: false,
			EnablePodPerceiver:         true,
			PodPerceiver:               &opssightv1.PodPerceiver{},
			AnnotationIntervalSeconds:  30,
			DumpIntervalMinutes:        30,
		},
		ScannerPod: &opssightv1.ScannerPod{
			ImageFacade: &opssightv1.ImageFacade{},
			Scanner: &opssightv1.Scanner{
				ClientTimeoutSeconds: 600,
			},
			ReplicaCount:   1,
			ImageDirectory: "/var/images",
		},
		Prometheus: &opssightv1.Prometheus{
			Expose: NONE,
		},
		Skyfire: &opssightv1.Skyfire{
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
			BlackduckSpec:                      GetBlackDuckTemplate(),
		},
		EnableMetrics: true,
		EnableSkyfire: false,
		DefaultCPU:    "300m",
		DefaultMem:    "1300Mi",
		ScannerCPU:    "300m",
		ScannerMem:    "1300Mi",
		LogLevel:      "debug",
	}
}

// GetOpsSightDefault returns the required fields for OpsSight
func GetOpsSightDefault() *opssightv1.OpsSightSpec {
	return &opssightv1.OpsSightSpec{
		Namespace: "opssight-test",
		Perceptor: &opssightv1.Perceptor{
			CheckForStalledScansPauseHours: 999999,
			StalledScanClientTimeoutHours:  999999,
			ModelMetricsPauseSeconds:       15,
			UnknownImagePauseMilliseconds:  15000,
			ClientTimeoutMilliseconds:      100000,
			Expose:                         NONE,
		},
		ScannerPod: &opssightv1.ScannerPod{
			Scanner: &opssightv1.Scanner{
				ClientTimeoutSeconds: 600,
			},
			ImageFacade: &opssightv1.ImageFacade{
				InternalRegistries: []*opssightv1.RegistryAuth{},
				ImagePullerType:    "skopeo",
			},
			ReplicaCount:   1,
			ImageDirectory: "/var/images",
		},
		Perceiver: &opssightv1.Perceiver{
			EnableImagePerceiver:       false,
			EnableArtifactoryPerceiver: false,
			EnablePodPerceiver:         false,
			PodPerceiver:               &opssightv1.PodPerceiver{},
			AnnotationIntervalSeconds:  30,
			DumpIntervalMinutes:        30,
		},
		Prometheus: &opssightv1.Prometheus{
			Expose: NONE,
		},
		EnableSkyfire: false,
		Skyfire: &opssightv1.Skyfire{
			HubClientTimeoutSeconds:      120,
			HubDumpPauseSeconds:          240,
			KubeDumpIntervalSeconds:      60,
			PerceptorDumpIntervalSeconds: 60,
		},
		EnableMetrics: false,
		DefaultCPU:    "300m",
		DefaultMem:    "1300Mi",
		ScannerCPU:    "300m",
		ScannerMem:    "1300Mi",
		LogLevel:      "debug",
		Blackduck: &opssightv1.Blackduck{
			InitialCount:                       0,
			MaxCount:                           0,
			ConnectionsEnvironmentVariableName: "blackduck.json",
			TLSVerification:                    false,
			DeleteBlackduckThresholdPercentage: 50,
			BlackduckSpec:                      GetBlackDuckTemplate(),
		},
	}
}

// GetOpsSightDefaultWithIPV6DisabledBlackDuck retuns an OpsSight with a BlackDuck and
// IPV6 disabled
func GetOpsSightDefaultWithIPV6DisabledBlackDuck() *opssightv1.OpsSightSpec {
	return &opssightv1.OpsSightSpec{
		Namespace: "opssight-test",
		Perceptor: &opssightv1.Perceptor{
			CheckForStalledScansPauseHours: 999999,
			StalledScanClientTimeoutHours:  999999,
			ModelMetricsPauseSeconds:       15,
			UnknownImagePauseMilliseconds:  15000,
			ClientTimeoutMilliseconds:      100000,
			Expose:                         NONE,
		},
		ScannerPod: &opssightv1.ScannerPod{
			Scanner: &opssightv1.Scanner{
				ClientTimeoutSeconds: 600,
			},
			ImageFacade: &opssightv1.ImageFacade{
				InternalRegistries: []*opssightv1.RegistryAuth{},
				ImagePullerType:    "skopeo",
			},
			ReplicaCount:   1,
			ImageDirectory: "/var/images",
		},
		Perceiver: &opssightv1.Perceiver{
			EnableImagePerceiver:       false,
			EnableArtifactoryPerceiver: false,
			EnablePodPerceiver:         true,
			PodPerceiver:               &opssightv1.PodPerceiver{},
			AnnotationIntervalSeconds:  30,
			DumpIntervalMinutes:        30,
		},
		Prometheus: &opssightv1.Prometheus{
			Expose: NONE,
		},
		EnableSkyfire: false,
		Skyfire: &opssightv1.Skyfire{
			HubClientTimeoutSeconds:      120,
			HubDumpPauseSeconds:          240,
			KubeDumpIntervalSeconds:      60,
			PerceptorDumpIntervalSeconds: 60,
		},
		EnableMetrics: true,
		DefaultCPU:    "300m",
		DefaultMem:    "1300Mi",
		ScannerCPU:    "300m",
		ScannerMem:    "1300Mi",
		LogLevel:      "debug",
		Blackduck: &opssightv1.Blackduck{
			InitialCount:                       0,
			MaxCount:                           0,
			ConnectionsEnvironmentVariableName: "blackduck.json",
			TLSVerification:                    false,
			DeleteBlackduckThresholdPercentage: 50,
			BlackduckSpec: &blackduckv1.BlackduckSpec{
				PersistentStorage: false,
				CertificateName:   "default",
				Environs: []string{
					"IPV4_ONLY:1",
					"BLACKDUCK_HUB_SERVER_ADDRESS:0.0.0.0",
				},
				Size: "small",
				Type: "worker",
			},
		},
	}
}

// GetAlertTemplate returns the required fields for Alert
func GetAlertTemplate() *alertv1.AlertSpec {
	return &alertv1.AlertSpec{}
}

// GetAlertDefault creates an Alert crd configuration object with defaults
func GetAlertDefault() *alertv1.AlertSpec {
	standAlone := false

	return &alertv1.AlertSpec{
		ExposeService:     NONE,
		Port:              IntToInt32(8443),
		PersistentStorage: false,
		PVCName:           "alert-pvc",
		StandAlone:        &standAlone,
		PVCSize:           "5G",
		AlertMemory:       "2560M",
		CfsslMemory:       "640M",
		Environs: []string{
			"ALERT_SERVER_PORT:8443",
			"PUBLIC_HUB_WEBSERVER_HOST:localhost",
			"PUBLIC_HUB_WEBSERVER_PORT:443",
		},
	}
}
