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

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/juju/errors"
)

// PerceiverConfig stores the Perceiver configuration
type PerceiverConfig struct {
	Certificate               string
	CertificateKey            string
	AnnotationIntervalSeconds int
	DumpIntervalMinutes       int
	Port                      int
	Pod                       *PodPerceiverConfig
	Image                     *ImagePerceiverConfig
	Artifactory               *ArtifactoryPerceiverConfig
}

// ImagePerceiverConfig stores the Image Perceiver configuration
type ImagePerceiverConfig struct{}

// ArtifactoryPerceiverConfig stores the Artifactory Perceiver configuration
type ArtifactoryPerceiverConfig struct {
	Dumper bool
}

// PodPerceiverConfig stores the Pod Perceiver configuration
type PodPerceiverConfig struct {
	NamespaceFilter string
}

// BlackDuckConfig stores the Black Duck configuration
type BlackDuckConfig struct {
	ConnectionsEnvironmentVariableName string
	TLSVerification                    bool
}

// PerceptorTimingsConfig stores the Perceptor configuration
type PerceptorTimingsConfig struct {
	CheckForStalledScansPauseHours int
	StalledScanClientTimeoutHours  int
	ModelMetricsPauseSeconds       int
	UnknownImagePauseMilliseconds  int
	ClientTimeoutMilliseconds      int
}

// PerceptorConfig stores the Perceptor configuration
type PerceptorConfig struct {
	Timings     *PerceptorTimingsConfig
	UseMockMode bool
	Host        string
	Port        int
}

// ScannerConfig stores the Perceptor Scanner configuration
type ScannerConfig struct {
	Port                          int
	ImageDirectory                string
	BlackDuckClientTimeoutSeconds int
}

// ImageFacadeConfig stores the Perceptor Image Facade configuration
type ImageFacadeConfig struct {
	Host             string
	Port             int
	ImagePullerType  string
	CreateImagesOnly bool
}

// SkyfireConfig stores the Skyfire configuration
type SkyfireConfig struct {
	UseInClusterConfig            bool
	Port                          int
	PrometheusPort                int
	BlackDuckClientTimeoutSeconds int
	KubeDumpIntervalSeconds       int
	PerceptorDumpIntervalSeconds  int
	BlackDuckDumpPauseSeconds     int
}

// MainOpssightConfigMap stores the opssight configmap
type MainOpssightConfigMap struct {
	Perceiver   *PerceiverConfig
	BlackDuck   *BlackDuckConfig
	Perceptor   *PerceptorConfig
	Scanner     *ScannerConfig
	ImageFacade *ImageFacadeConfig
	Skyfire     *SkyfireConfig
	LogLevel    string
}

func (cm *MainOpssightConfigMap) jsonString() (string, error) {
	bytes, err := json.MarshalIndent(cm, "", "  ")
	if err != nil {
		return "", errors.Annotate(err, "unable to serialize to json")
	}
	return string(bytes), nil
}

func (cm *MainOpssightConfigMap) horizonConfigMap(name string, namespace string, filename string) (*components.ConfigMap, error) {
	configMap := components.NewConfigMap(horizonapi.ConfigMapConfig{
		Name:      name,
		Namespace: namespace,
	})
	configMapString, err := cm.jsonString()
	if err != nil {
		return nil, errors.Trace(err)
	}
	configMap.AddLabels(map[string]string{"component": name, "app": "opssight"})
	configMap.AddData(map[string]string{filename: configMapString})

	return configMap, nil
}
