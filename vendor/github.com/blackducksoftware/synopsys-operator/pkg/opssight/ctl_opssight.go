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

	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	crddefaults "github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Ctl type provides functionality for an OpsSight
// for the Synopsysctl tool
type Ctl struct {
	Spec                                             *opssightv1.OpsSightSpec
	PerceptorName                                    string
	PerceptorImage                                   string
	PerceptorPort                                    int
	PerceptorCheckForStalledScansPauseHours          int
	PerceptorStalledScanClientTimeoutHours           int
	PerceptorModelMetricsPauseSeconds                int
	PerceptorUnknownImagePauseMilliseconds           int
	PerceptorClientTimeoutMilliseconds               int
	ScannerPodName                                   string
	ScannerPodScannerName                            string
	ScannerPodScannerImage                           string
	ScannerPodScannerPort                            int
	ScannerPodScannerClientTimeoutSeconds            int
	ScannerPodImageFacadeName                        string
	ScannerPodImageFacadeImage                       string
	ScannerPodImageFacadePort                        int
	ScannerPodImageFacadeInternalRegistriesJSONSlice []string
	ScannerPodImageFacadeImagePullerType             string
	ScannerPodImageFacadeServiceAccount              string
	ScannerPodReplicaCount                           int
	ScannerPodImageDirectory                         string
	PerceiverEnableImagePerceiver                    bool
	PerceiverEnablePodPerceiver                      bool
	PerceiverImagePerceiverName                      string
	PerceiverImagePerceiverImage                     string
	PerceiverPodPerceiverName                        string
	PerceiverPodPerceiverImage                       string
	PerceiverPodPerceiverNamespaceFilter             string
	PerceiverAnnotationIntervalSeconds               int
	PerceiverDumpIntervalMinutes                     int
	PerceiverServiceAccount                          string
	PerceiverPort                                    int
	ConfigMapName                                    string
	SecretName                                       string
	DefaultCPU                                       string
	DefaultMem                                       string
	ScannerCPU                                       string
	ScannerMem                                       string
	LogLevel                                         string
	EnableMetrics                                    bool
	PrometheusName                                   string
	PrometheusImage                                  string
	PrometheusPort                                   int
	EnableSkyfire                                    bool
	SkyfireName                                      string
	SkyfireImage                                     string
	SkyfirePort                                      int
	SkyfirePrometheusPort                            int
	SkyfireServiceAccount                            string
	SkyfireHubClientTimeoutSeconds                   int
	SkyfireHubDumpPauseSeconds                       int
	SkyfireKubeDumpIntervalSeconds                   int
	SkyfirePerceptorDumpIntervalSeconds              int
	BlackduckExternalHostsJSON                       []string
	BlackduckConnectionsEnvironmentVaraiableName     string
	BlackduckTLSVerification                         bool
	BlackduckPasswordEnvVar                          string
	BlackduckInitialCount                            int
	BlackduckMaxCount                                int
	BlackduckDeleteBlackduckThresholdPercentage      int
}

// NewOpsSightCtl creates a new Ctl struct
func NewOpsSightCtl() *Ctl {
	return &Ctl{
		Spec:                                             &opssightv1.OpsSightSpec{},
		PerceptorName:                                    "",
		PerceptorImage:                                   "",
		PerceptorPort:                                    0,
		PerceptorCheckForStalledScansPauseHours:          0,
		PerceptorStalledScanClientTimeoutHours:           0,
		PerceptorModelMetricsPauseSeconds:                0,
		PerceptorUnknownImagePauseMilliseconds:           0,
		PerceptorClientTimeoutMilliseconds:               0,
		ScannerPodName:                                   "",
		ScannerPodScannerName:                            "",
		ScannerPodScannerImage:                           "",
		ScannerPodScannerPort:                            0,
		ScannerPodScannerClientTimeoutSeconds:            0,
		ScannerPodImageFacadeName:                        "",
		ScannerPodImageFacadeImage:                       "",
		ScannerPodImageFacadePort:                        0,
		ScannerPodImageFacadeInternalRegistriesJSONSlice: []string{},
		ScannerPodImageFacadeImagePullerType:             "",
		ScannerPodImageFacadeServiceAccount:              "",
		ScannerPodReplicaCount:                           0,
		ScannerPodImageDirectory:                         "",
		PerceiverEnableImagePerceiver:                    false,
		PerceiverEnablePodPerceiver:                      false,
		PerceiverImagePerceiverName:                      "",
		PerceiverImagePerceiverImage:                     "",
		PerceiverPodPerceiverName:                        "",
		PerceiverPodPerceiverImage:                       "",
		PerceiverPodPerceiverNamespaceFilter:             "",
		PerceiverAnnotationIntervalSeconds:               0,
		PerceiverDumpIntervalMinutes:                     0,
		PerceiverServiceAccount:                          "",
		PerceiverPort:                                    0,
		ConfigMapName:                                    "",
		SecretName:                                       "",
		DefaultCPU:                                       "",
		DefaultMem:                                       "",
		ScannerCPU:                                       "",
		ScannerMem:                                       "",
		LogLevel:                                         "",
		EnableMetrics:                                    false,
		PrometheusName:                                   "",
		PrometheusImage:                                  "",
		PrometheusPort:                                   0,
		EnableSkyfire:                                    false,
		SkyfireName:                                      "",
		SkyfireImage:                                     "",
		SkyfirePort:                                      0,
		SkyfirePrometheusPort:                            0,
		SkyfireServiceAccount:                            "",
		SkyfireHubClientTimeoutSeconds:                   0,
		SkyfireHubDumpPauseSeconds:                       0,
		SkyfireKubeDumpIntervalSeconds:                   0,
		SkyfirePerceptorDumpIntervalSeconds:              0,
		BlackduckExternalHostsJSON:                       []string{},
		BlackduckConnectionsEnvironmentVaraiableName:     "",
		BlackduckTLSVerification:                         false,
		BlackduckInitialCount:                            0,
		BlackduckMaxCount:                                0,
		BlackduckDeleteBlackduckThresholdPercentage:      0,
	}
}

// GetSpec returns the Spec for the resource
func (ctl *Ctl) GetSpec() interface{} {
	return *ctl.Spec
}

// SetSpec sets the Spec for the resource
func (ctl *Ctl) SetSpec(spec interface{}) error {
	convertedSpec, ok := spec.(opssightv1.OpsSightSpec)
	if !ok {
		return fmt.Errorf("Error setting OpsSight Spec")
	}
	ctl.Spec = &convertedSpec
	return nil
}

// CheckSpecFlags returns an error if a user input was invalid
func (ctl *Ctl) CheckSpecFlags() error {
	for _, registryJSON := range ctl.ScannerPodImageFacadeInternalRegistriesJSONSlice {
		registry := &opssightv1.RegistryAuth{}
		err := json.Unmarshal([]byte(registryJSON), registry)
		if err != nil {
			return fmt.Errorf("Invalid Registry Format")
		}
	}
	return nil
}

// SwitchSpec switches the OpsSight's Spec to a different predefined spec
func (ctl *Ctl) SwitchSpec(createOpsSightSpecType string) error {
	switch createOpsSightSpecType {
	case "empty":
		ctl.Spec = &opssightv1.OpsSightSpec{}
	case "disabledBlackduck":
		ctl.Spec = crddefaults.GetOpsSightDefaultValueWithDisabledHub()
	case "default":
		ctl.Spec = crddefaults.GetOpsSightDefaultValue()
	default:
		return fmt.Errorf("OpsSight Spec Type %s does not match: empty, disabledBlackduck, default", createOpsSightSpecType)
	}
	return nil
}

// AddSpecFlags adds flags for the OpsSight's Spec to the command
// master - if false, doesn't add flags that all Users shouldn't use
func (ctl *Ctl) AddSpecFlags(cmd *cobra.Command, master bool) {
	if master {
		cmd.Flags().StringVar(&ctl.PerceptorName, "perceptor-name", ctl.PerceptorName, "Name of the Perceptor")
		cmd.Flags().StringVar(&ctl.ScannerPodName, "scannerpod-name", ctl.ScannerPodName, "Name of the ScannerPod")
		cmd.Flags().StringVar(&ctl.ScannerPodScannerName, "scannerpod-scanner-name", ctl.ScannerPodScannerName, "Name of the ScannerPod's Scanner Container")
		cmd.Flags().StringVar(&ctl.ScannerPodImageFacadeName, "scannerpod-imagefacade-name", ctl.ScannerPodImageFacadeName, "Name of the ScannerPod's ImageFacade Container")
		cmd.Flags().StringVar(&ctl.PerceiverImagePerceiverName, "imageperceiver-name", ctl.PerceiverImagePerceiverName, "Name of the ImagePerceiver")
		cmd.Flags().StringVar(&ctl.PerceiverPodPerceiverName, "podperceiver-name", ctl.PerceiverPodPerceiverName, "Name of the PodPerceiver")
		cmd.Flags().StringVar(&ctl.PerceiverServiceAccount, "perceiver-service-account", ctl.PerceiverServiceAccount, "TODO")
		cmd.Flags().StringVar(&ctl.PrometheusName, "prometheus-name", ctl.PrometheusName, "Name of Prometheus")
		cmd.Flags().StringVar(&ctl.SkyfireName, "skyfire-name", ctl.SkyfireName, "Name of Skyfire")
		cmd.Flags().StringVar(&ctl.SkyfireServiceAccount, "skyfire-service-account", ctl.SkyfireServiceAccount, "Service Account for Skyfire")
		cmd.Flags().StringVar(&ctl.BlackduckConnectionsEnvironmentVaraiableName, "blackduck-connections-environment-variable-name", ctl.BlackduckConnectionsEnvironmentVaraiableName, "TODO")
		cmd.Flags().StringVar(&ctl.ConfigMapName, "config-map-name", ctl.ConfigMapName, "Name of the config map for OpsSight")
		cmd.Flags().StringVar(&ctl.SecretName, "secret-name", ctl.SecretName, "Name of the secret for OpsSight")
	}
	cmd.Flags().StringVar(&ctl.PerceptorImage, "perceptor-image", ctl.PerceptorImage, "Image of the Perceptor")
	cmd.Flags().IntVar(&ctl.PerceptorPort, "perceptor-port", ctl.PerceptorPort, "Port for the Perceptor")
	cmd.Flags().IntVar(&ctl.PerceptorCheckForStalledScansPauseHours, "perceptor-check-scan-hours", ctl.PerceptorCheckForStalledScansPauseHours, "Hours the Percpetor waits between checking for scans")
	cmd.Flags().IntVar(&ctl.PerceptorStalledScanClientTimeoutHours, "perceptor-scan-client-timeout-hours", ctl.PerceptorStalledScanClientTimeoutHours, "Hours until Perceptor stops checking for scans")
	cmd.Flags().IntVar(&ctl.PerceptorModelMetricsPauseSeconds, "perceptor-metrics-pause-seconds", ctl.PerceptorModelMetricsPauseSeconds, "TODO")
	cmd.Flags().IntVar(&ctl.PerceptorUnknownImagePauseMilliseconds, "perceptor-unknown-image-pause-milliseconds", ctl.PerceptorUnknownImagePauseMilliseconds, "TODO")
	cmd.Flags().IntVar(&ctl.PerceptorClientTimeoutMilliseconds, "perceptor-client-timeout-milliseconds", ctl.PerceptorClientTimeoutMilliseconds, "TODO")
	cmd.Flags().StringVar(&ctl.ScannerPodScannerImage, "scannerpod-scanner-image", ctl.ScannerPodScannerImage, "Scanner Container's image")
	cmd.Flags().IntVar(&ctl.ScannerPodScannerPort, "scannerpod-scanner-port", ctl.ScannerPodScannerPort, "Scanner Container's port")
	cmd.Flags().IntVar(&ctl.ScannerPodScannerClientTimeoutSeconds, "scannerpod-scanner-client-timeout-seconds", ctl.ScannerPodScannerClientTimeoutSeconds, "TODO")
	cmd.Flags().StringVar(&ctl.ScannerPodImageFacadeImage, "scannerpod-imagefacade-image", ctl.ScannerPodImageFacadeImage, "ImageFacade Container's image")
	cmd.Flags().IntVar(&ctl.ScannerPodImageFacadePort, "scannerpod-imagefacade-port", ctl.ScannerPodImageFacadePort, "ImageFacade Container's port")
	cmd.Flags().StringSliceVar(&ctl.ScannerPodImageFacadeInternalRegistriesJSONSlice, "scannerpod-imagefacade-internal-registries", ctl.ScannerPodImageFacadeInternalRegistriesJSONSlice, "TODO")
	cmd.Flags().StringVar(&ctl.ScannerPodImageFacadeImagePullerType, "scannerpod-imagefacade-image-puller-type", ctl.ScannerPodImageFacadeImagePullerType, "Type of ImageFacade's Image Puller - docker, skopeo")
	cmd.Flags().StringVar(&ctl.ScannerPodImageFacadeServiceAccount, "scannerpod-imagefacade-service-account", ctl.ScannerPodImageFacadeServiceAccount, "Service Account for the ImageFacade")
	cmd.Flags().IntVar(&ctl.ScannerPodReplicaCount, "scannerpod-replica-count", ctl.ScannerPodReplicaCount, "TODO")
	cmd.Flags().StringVar(&ctl.ScannerPodImageDirectory, "scannerpod-image-directory", ctl.ScannerPodImageDirectory, "TODO")
	cmd.Flags().BoolVar(&ctl.PerceiverEnableImagePerceiver, "enable-image-perceiver", ctl.PerceiverEnableImagePerceiver, "TODO")
	cmd.Flags().BoolVar(&ctl.PerceiverEnablePodPerceiver, "enable-pod-perceiver", ctl.PerceiverEnablePodPerceiver, "TODO")
	cmd.Flags().StringVar(&ctl.PerceiverImagePerceiverImage, "imageperceiver-image", ctl.PerceiverImagePerceiverImage, "Image of the ImagePerceiver")
	cmd.Flags().StringVar(&ctl.PerceiverPodPerceiverImage, "podperceiver-image", ctl.PerceiverPodPerceiverImage, "Image of the PodPerceiver")
	cmd.Flags().StringVar(&ctl.PerceiverPodPerceiverNamespaceFilter, "podperceiver-namespace-filter", ctl.PerceiverPodPerceiverNamespaceFilter, "TODO")
	cmd.Flags().IntVar(&ctl.PerceiverAnnotationIntervalSeconds, "perceiver-annotation-interval-seconds", ctl.PerceiverAnnotationIntervalSeconds, "TODO")
	cmd.Flags().IntVar(&ctl.PerceiverDumpIntervalMinutes, "perceiver-dump-interval-minutes", ctl.PerceiverDumpIntervalMinutes, "TODO")
	cmd.Flags().IntVar(&ctl.PerceiverPort, "perceiver-port", ctl.PerceiverPort, "Port for the Perceiver")
	cmd.Flags().StringVar(&ctl.ConfigMapName, "configmapname", ctl.ConfigMapName, "TODO")
	cmd.Flags().StringVar(&ctl.SecretName, "secretname", ctl.SecretName, "TODO")
	cmd.Flags().StringVar(&ctl.DefaultCPU, "defaultcpu", ctl.DefaultCPU, "TODO")
	cmd.Flags().StringVar(&ctl.DefaultMem, "defaultmem", ctl.DefaultMem, "TODO")
	cmd.Flags().StringVar(&ctl.ScannerCPU, "scannercpu", ctl.ScannerCPU, "TODO")
	cmd.Flags().StringVar(&ctl.ScannerMem, "scannermem", ctl.ScannerMem, "TODO")
	cmd.Flags().StringVar(&ctl.LogLevel, "log-level", ctl.LogLevel, "TODO")
	cmd.Flags().BoolVar(&ctl.EnableMetrics, "enable-metrics", ctl.EnableMetrics, "TODO")
	cmd.Flags().StringVar(&ctl.PrometheusImage, "prometheus-image", ctl.PrometheusImage, "Image for Prometheus")
	cmd.Flags().IntVar(&ctl.PrometheusPort, "prometheus-port", ctl.PrometheusPort, "Port for Prometheus")
	cmd.Flags().BoolVar(&ctl.EnableSkyfire, "enable-skyfire", ctl.EnableSkyfire, "Enables Skyfire Pod if true")
	cmd.Flags().StringVar(&ctl.SkyfireImage, "skyfire-image", ctl.SkyfireImage, "Image of Skyfire")
	cmd.Flags().IntVar(&ctl.SkyfirePort, "skyfire-port", ctl.SkyfirePort, "Port of Skyfire")
	cmd.Flags().IntVar(&ctl.SkyfirePrometheusPort, "skyfire-prometheus-port", ctl.SkyfirePrometheusPort, "Skyfire's Prometheus port")
	cmd.Flags().IntVar(&ctl.SkyfireHubClientTimeoutSeconds, "skyfire-hub-client-timeout-seconds", ctl.SkyfireHubClientTimeoutSeconds, "TODO")
	cmd.Flags().IntVar(&ctl.SkyfireHubDumpPauseSeconds, "skyfire-hub-dump-pause-seconds", ctl.SkyfireHubDumpPauseSeconds, "Seconds Skyfire waits between querying Blackducks")
	cmd.Flags().IntVar(&ctl.SkyfireKubeDumpIntervalSeconds, "skyfire-kube-dump-interval-seconds", ctl.SkyfireKubeDumpIntervalSeconds, "Seconds Skyfire waits between querying the KubeAPI")
	cmd.Flags().IntVar(&ctl.SkyfirePerceptorDumpIntervalSeconds, "skyfire-perceptor-dump-interval-seconds", ctl.SkyfirePerceptorDumpIntervalSeconds, "Seconds Skyfire waits between querying the Perceptor Model")
	cmd.Flags().StringSliceVar(&ctl.BlackduckExternalHostsJSON, "blackduck-external-hosts", ctl.BlackduckExternalHostsJSON, "List of Blackduck External Hosts")
	cmd.Flags().BoolVar(&ctl.BlackduckTLSVerification, "blackduck-TLS-verification", ctl.BlackduckTLSVerification, "TODO")
	cmd.Flags().IntVar(&ctl.BlackduckInitialCount, "blackduck-initial-count", ctl.BlackduckInitialCount, "Initial number of Blackducks to create")
	cmd.Flags().IntVar(&ctl.BlackduckMaxCount, "blackduck-max-count", ctl.BlackduckMaxCount, "Maximum number of Blackducks that can be created")
	cmd.Flags().IntVar(&ctl.BlackduckDeleteBlackduckThresholdPercentage, "blackduck-delete-blackduck-threshold-percentage", ctl.BlackduckDeleteBlackduckThresholdPercentage, "TODO")
}

// SetChangedFlags visits every flag and calls setFlag to update
// the resource's spec
func (ctl *Ctl) SetChangedFlags(flagset *pflag.FlagSet) {
	flagset.VisitAll(ctl.SetFlag)
}

// SetFlag sets an OpsSights's Spec field if its flag was changed
func (ctl *Ctl) SetFlag(f *pflag.Flag) {
	if f.Changed {
		log.Debugf("Flag %s: CHANGED\n", f.Name)
		switch f.Name {
		case "perceptor-name":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightv1.Perceptor{}
			}
			ctl.Spec.Perceptor.Name = ctl.PerceptorName
		case "perceptor-image":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightv1.Perceptor{}
			}
			ctl.Spec.Perceptor.Image = ctl.PerceptorImage
		case "perceptor-port":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightv1.Perceptor{}
			}
			ctl.Spec.Perceptor.Port = ctl.PerceptorPort
		case "perceptor-check-scan-hours":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightv1.Perceptor{}
			}
			ctl.Spec.Perceptor.CheckForStalledScansPauseHours = ctl.PerceptorCheckForStalledScansPauseHours
		case "perceptor-scan-client-timeout-hours":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightv1.Perceptor{}
			}
			ctl.Spec.Perceptor.StalledScanClientTimeoutHours = ctl.PerceptorStalledScanClientTimeoutHours
		case "perceptor-metrics-pause-seconds":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightv1.Perceptor{}
			}
			ctl.Spec.Perceptor.ModelMetricsPauseSeconds = ctl.PerceptorModelMetricsPauseSeconds
		case "perceptor-unknown-image-pause-milliseconds":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightv1.Perceptor{}
			}
			ctl.Spec.Perceptor.UnknownImagePauseMilliseconds = ctl.PerceptorUnknownImagePauseMilliseconds
		case "perceptor-client-timeout-milliseconds":
			if ctl.Spec.Perceptor == nil {
				ctl.Spec.Perceptor = &opssightv1.Perceptor{}
			}
			ctl.Spec.Perceptor.ClientTimeoutMilliseconds = ctl.PerceptorClientTimeoutMilliseconds
		case "scannerpod-name":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			ctl.Spec.ScannerPod.Name = ctl.ScannerPodName
		case "scannerpod-scanner-name":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.Scanner == nil {
				ctl.Spec.ScannerPod.Scanner = &opssightv1.Scanner{}
			}
			ctl.Spec.ScannerPod.Scanner.Name = ctl.ScannerPodScannerName
		case "scannerpod-scanner-image":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.Scanner == nil {
				ctl.Spec.ScannerPod.Scanner = &opssightv1.Scanner{}
			}
			ctl.Spec.ScannerPod.Scanner.Image = ctl.ScannerPodScannerImage
		case "scannerpod-scanner-port":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.Scanner == nil {
				ctl.Spec.ScannerPod.Scanner = &opssightv1.Scanner{}
			}
			ctl.Spec.ScannerPod.Scanner.Port = ctl.ScannerPodScannerPort
		case "scannerpod-scanner-client-timeout-seconds":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.Scanner == nil {
				ctl.Spec.ScannerPod.Scanner = &opssightv1.Scanner{}
			}
			ctl.Spec.ScannerPod.Scanner.ClientTimeoutSeconds = ctl.ScannerPodScannerClientTimeoutSeconds
		case "scannerpod-imagefacade-name":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.ImageFacade == nil {
				ctl.Spec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			ctl.Spec.ScannerPod.ImageFacade.Name = ctl.ScannerPodImageFacadeName
		case "scannerpod-imagefacade-image":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.ImageFacade == nil {
				ctl.Spec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			ctl.Spec.ScannerPod.ImageFacade.Image = ctl.ScannerPodImageFacadeImage
		case "scannerpod-imagefacade-port":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.ImageFacade == nil {
				ctl.Spec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			ctl.Spec.ScannerPod.ImageFacade.Port = ctl.ScannerPodImageFacadePort
		case "scannerpod-imagefacade-internal-registries":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.ImageFacade == nil {
				ctl.Spec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			for _, registryJSON := range ctl.ScannerPodImageFacadeInternalRegistriesJSONSlice {
				registry := &opssightv1.RegistryAuth{}
				json.Unmarshal([]byte(registryJSON), registry)
				ctl.Spec.ScannerPod.ImageFacade.InternalRegistries = append(ctl.Spec.ScannerPod.ImageFacade.InternalRegistries, registry)
			}
		case "scannerpod-imagefacade-image-puller-type":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.ImageFacade == nil {
				ctl.Spec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			ctl.Spec.ScannerPod.ImageFacade.ImagePullerType = ctl.ScannerPodImageFacadeImagePullerType
		case "scannerpod-imagefacade-service-account":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			if ctl.Spec.ScannerPod.ImageFacade == nil {
				ctl.Spec.ScannerPod.ImageFacade = &opssightv1.ImageFacade{}
			}
			ctl.Spec.ScannerPod.ImageFacade.ServiceAccount = ctl.ScannerPodImageFacadeServiceAccount
		case "scannerpod-replica-count":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			ctl.Spec.ScannerPod.ReplicaCount = ctl.ScannerPodReplicaCount
		case "scannerpod-image-directory":
			if ctl.Spec.ScannerPod == nil {
				ctl.Spec.ScannerPod = &opssightv1.ScannerPod{}
			}
			ctl.Spec.ScannerPod.ImageDirectory = ctl.ScannerPodImageDirectory
		case "enable-image-perceiver":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			ctl.Spec.Perceiver.EnableImagePerceiver = ctl.PerceiverEnableImagePerceiver
		case "enable-pod-perceiver":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			ctl.Spec.Perceiver.EnablePodPerceiver = ctl.PerceiverEnablePodPerceiver
		case "imageperceiver-name":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			if ctl.Spec.Perceiver.ImagePerceiver == nil {
				ctl.Spec.Perceiver.ImagePerceiver = &opssightv1.ImagePerceiver{}
			}
			ctl.Spec.Perceiver.ImagePerceiver.Name = ctl.PerceiverImagePerceiverName
		case "imageperceiver-image":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			if ctl.Spec.Perceiver.ImagePerceiver == nil {
				ctl.Spec.Perceiver.ImagePerceiver = &opssightv1.ImagePerceiver{}
			}
			ctl.Spec.Perceiver.ImagePerceiver.Image = ctl.PerceiverImagePerceiverImage
		case "podperceiver-name":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			if ctl.Spec.Perceiver.PodPerceiver == nil {
				ctl.Spec.Perceiver.PodPerceiver = &opssightv1.PodPerceiver{}
			}
			ctl.Spec.Perceiver.PodPerceiver.Name = ctl.PerceiverPodPerceiverName
		case "podperceiver-image":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			if ctl.Spec.Perceiver.PodPerceiver == nil {
				ctl.Spec.Perceiver.PodPerceiver = &opssightv1.PodPerceiver{}
			}
			ctl.Spec.Perceiver.PodPerceiver.Image = ctl.PerceiverPodPerceiverImage
		case "podperceiver-namespace-filter":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			if ctl.Spec.Perceiver.PodPerceiver == nil {
				ctl.Spec.Perceiver.PodPerceiver = &opssightv1.PodPerceiver{}
			}
			ctl.Spec.Perceiver.PodPerceiver.NamespaceFilter = ctl.PerceiverPodPerceiverNamespaceFilter
		case "perceiver-annotation-interval-seconds":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			ctl.Spec.Perceiver.AnnotationIntervalSeconds = ctl.PerceiverAnnotationIntervalSeconds
		case "perceiver-dump-interval-minutes":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			ctl.Spec.Perceiver.DumpIntervalMinutes = ctl.PerceiverDumpIntervalMinutes
		case "perceiver-service-account":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			ctl.Spec.Perceiver.ServiceAccount = ctl.PerceiverServiceAccount
		case "perceiver-port":
			if ctl.Spec.Perceiver == nil {
				ctl.Spec.Perceiver = &opssightv1.Perceiver{}
			}
			ctl.Spec.Perceiver.Port = ctl.PerceiverPort
		case "configmapname":
			ctl.Spec.ConfigMapName = ctl.ConfigMapName
		case "secretname":
			ctl.Spec.SecretName = ctl.SecretName
		case "defaultcpu":
			ctl.Spec.DefaultCPU = ctl.DefaultCPU
		case "defaultmem":
			ctl.Spec.DefaultMem = ctl.DefaultMem
		case "scannercpu":
			ctl.Spec.ScannerCPU = ctl.ScannerCPU
		case "scannermem":
			ctl.Spec.ScannerMem = ctl.ScannerMem
		case "log-level":
			ctl.Spec.LogLevel = ctl.LogLevel
		case "enable-metrics":
			ctl.Spec.EnableMetrics = ctl.EnableMetrics
		case "prometheus-name":
			if ctl.Spec.Prometheus == nil {
				ctl.Spec.Prometheus = &opssightv1.Prometheus{}
			}
			ctl.Spec.Prometheus.Name = ctl.PrometheusName
		case "prometheus-image":
			if ctl.Spec.Prometheus == nil {
				ctl.Spec.Prometheus = &opssightv1.Prometheus{}
			}
			ctl.Spec.Prometheus.Image = ctl.PrometheusImage
		case "prometheus-port":
			if ctl.Spec.Prometheus == nil {
				ctl.Spec.Prometheus = &opssightv1.Prometheus{}
			}
			ctl.Spec.Prometheus.Port = ctl.PrometheusPort
		case "enable-skyfire":
			ctl.Spec.EnableSkyfire = ctl.EnableSkyfire
		case "skyfire-name":
			if ctl.Spec.Skyfire == nil {
				ctl.Spec.Skyfire = &opssightv1.Skyfire{}
			}
			ctl.Spec.Skyfire.Name = ctl.SkyfireName
		case "skyfire-image":
			if ctl.Spec.Skyfire == nil {
				ctl.Spec.Skyfire = &opssightv1.Skyfire{}
			}
			ctl.Spec.Skyfire.Image = ctl.SkyfireImage
		case "skyfire-port":
			if ctl.Spec.Skyfire == nil {
				ctl.Spec.Skyfire = &opssightv1.Skyfire{}
			}
			ctl.Spec.Skyfire.Port = ctl.SkyfirePort
		case "skyfire-prometheus-port":
			if ctl.Spec.Skyfire == nil {
				ctl.Spec.Skyfire = &opssightv1.Skyfire{}
			}
			ctl.Spec.Skyfire.PrometheusPort = ctl.SkyfirePrometheusPort
		case "skyfire-service-account":
			if ctl.Spec.Skyfire == nil {
				ctl.Spec.Skyfire = &opssightv1.Skyfire{}
			}
			ctl.Spec.Skyfire.ServiceAccount = ctl.SkyfireServiceAccount
		case "skyfire-hub-client-timeout-seconds":
			if ctl.Spec.Skyfire == nil {
				ctl.Spec.Skyfire = &opssightv1.Skyfire{}
			}
			ctl.Spec.Skyfire.HubClientTimeoutSeconds = ctl.SkyfireHubClientTimeoutSeconds
		case "skyfire-hub-dump-pause-seconds":
			if ctl.Spec.Skyfire == nil {
				ctl.Spec.Skyfire = &opssightv1.Skyfire{}
			}
			ctl.Spec.Skyfire.HubDumpPauseSeconds = ctl.SkyfireHubDumpPauseSeconds
		case "skyfire-kube-dump-interval-seconds":
			if ctl.Spec.Skyfire == nil {
				ctl.Spec.Skyfire = &opssightv1.Skyfire{}
			}
			ctl.Spec.Skyfire.KubeDumpIntervalSeconds = ctl.SkyfireKubeDumpIntervalSeconds
		case "skyfire-perceptor-dump-interval-seconds":
			if ctl.Spec.Skyfire == nil {
				ctl.Spec.Skyfire = &opssightv1.Skyfire{}
			}
			ctl.Spec.Skyfire.PerceptorDumpIntervalSeconds = ctl.SkyfirePerceptorDumpIntervalSeconds
		case "blackduck-external-hosts":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightv1.Blackduck{}
			}
			for _, hostJSON := range ctl.BlackduckExternalHostsJSON {
				host := &opssightv1.Host{}
				json.Unmarshal([]byte(hostJSON), host)
				ctl.Spec.Blackduck.ExternalHosts = append(ctl.Spec.Blackduck.ExternalHosts, host)
			}
		case "blackduck-connections-environment-variable-name":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightv1.Blackduck{}
			}
			ctl.Spec.Blackduck.ConnectionsEnvironmentVariableName = ctl.BlackduckConnectionsEnvironmentVaraiableName
		case "blackduck-TLS-verification":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightv1.Blackduck{}
			}
			ctl.Spec.Blackduck.TLSVerification = ctl.BlackduckTLSVerification
		case "blackduck-initial-count":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightv1.Blackduck{}
			}
			ctl.Spec.Blackduck.InitialCount = ctl.BlackduckInitialCount
		case "blackduck-max-count":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightv1.Blackduck{}
			}
			ctl.Spec.Blackduck.MaxCount = ctl.BlackduckMaxCount
		case "blackduck-delete-blackduck-threshold-percentage":
			if ctl.Spec.Blackduck == nil {
				ctl.Spec.Blackduck = &opssightv1.Blackduck{}
			}
			ctl.Spec.Blackduck.DeleteBlackduckThresholdPercentage = ctl.BlackduckDeleteBlackduckThresholdPercentage
		default:
			log.Debugf("Flag %s: Not Found\n", f.Name)
		}
	} else {
		log.Debugf("Flag %s: UNCHANGED\n", f.Name)
	}
}

// SpecIsValid verifies the spec has necessary fields to deploy
func (ctl *Ctl) SpecIsValid() (bool, error) {
	return true, nil
}

// CanUpdate checks if a user has permission to modify based on the spec
func (ctl *Ctl) CanUpdate() (bool, error) {
	return true, nil
}
