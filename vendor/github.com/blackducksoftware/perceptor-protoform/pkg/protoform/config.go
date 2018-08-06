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

package protoform

import (
	"encoding/json"
	"fmt"

	"github.com/blackducksoftware/perceptor-protoform/pkg/api"
	"github.com/koki/short/util/floatstr"

	"k8s.io/apimachinery/pkg/api/resource"
)

type protoformConfig struct {
	// Dry run wont actually install, but will print the objects definitions out.
	DryRun bool

	// CONTAINER CONFIGS
	// These are sed replaced into the config maps for the containers.
	PerceptorPort                    int
	ScannerPort                      int
	PerceiverPort                    int
	ImageFacadePort                  int
	InternalDockerRegistries         []string
	AnnotationIntervalSeconds        int
	DumpIntervalMinutes              int
	HubHost                          string
	HubUser                          string
	HubUserPassword                  string
	HubPort                          int
	HubClientTimeoutPerceptorSeconds int
	HubClientTimeoutScannerSeconds   int
	ConcurrentScanLimit              int
	Namespace                        string
	DefaultVersion                   string

	// CONTAINER PULL CONFIG
	// These are for defining docker registry and image location and versions
	Registry  string
	ImagePath string

	PerceptorImageName      string
	ScannerImageName        string
	PodPerceiverImageName   string
	ImagePerceiverImageName string
	ImageFacadeImageName    string
	SkyfireImageName        string

	PerceptorContainerVersion   string
	ScannerContainerVersion     string
	PerceiverContainerVersion   string
	ImageFacadeContainerVersion string
	SkyfireContainerVersion     string

	// AUTH CONFIGS
	// These are given to containers through secrets or other mechanisms.
	// Not necessarily a one-to-one text replacement.
	// TODO Lets try to have this injected on serviceaccount
	// at pod startup, eventually Service accounts.
	DockerPasswordOrToken string
	DockerUsername        string

	ServiceAccounts map[string]string
	Openshift       bool

	// CPU and memory configurations
	DefaultCPU string // Should be passed like: e.g. "300m"
	DefaultMem string // Should be passed like: e.g "1300Mi"

	// Log level
	LogLevel string

	// Viper secrets
	ViperSecret string

	// Environment Variables
	HubUserPasswordEnvVar string

	// Automate test
	PerceptorSkyfire bool
}

type envSecret struct {
	EnvName       string
	SecretName    string
	KeyFromSecret string
}

// ReplicationController defines the replication controller or pod configuration. Those configurations will be used for the creation of replication controller or pod
type ReplicationController struct {
	ConfigMapMounts map[string]string
	EmptyDirMounts  map[string]string
	Name            string
	Image           string
	Port            int32
	Cmd             []string
	Arg             []floatstr.FloatOrString
	Replicas        int32
	Env             []envSecret

	// key:value = name:mountPath
	EmptyDirVolumeMounts map[string]string

	// if true, then container is privileged /var/run/docker.sock.
	DockerSocket bool

	ServiceAccount     string
	ServiceAccountName string

	Memory resource.Quantity
	CPU    resource.Quantity
}

// NewDefaultsObj returns a ProtoformDefaults object with sane default for most options
func NewDefaultsObj() *api.ProtoformDefaults {
	return &api.ProtoformDefaults{
		PerceptorPort:                    3001,
		PerceiverPort:                    3002,
		ScannerPort:                      3003,
		ImageFacadePort:                  3004,
		AnnotationIntervalSeconds:        30,
		DumpIntervalMinutes:              30,
		HubClientTimeoutPerceptorSeconds: 5,
		HubClientTimeoutScannerSeconds:   30,
		HubHost:                 "nginx-webapp-logstash",
		HubPort:                 8443,
		DockerUsername:          "admin",
		ConcurrentScanLimit:     7,
		DefaultVersion:          "master",
		PerceptorImageName:      "perceptor",
		ScannerImageName:        "perceptor-scanner",
		ImagePerceiverImageName: "image-perceiver",
		PodPerceiverImageName:   "pod-perceiver",
		ImageFacadeImageName:    "perceptor-imagefacade",
		SkyfireImageName:        "skyfire",
		LogLevel:                "debug",
	}
}

func generateStringFromStringArr(strArr []string) string {
	str, _ := json.Marshal(strArr)
	return string(str)
}

func (p *protoformConfig) toMap() map[string]map[string]string {
	configs := map[string]map[string]string{
		"prometheus":            {"prometheus.yml": fmt.Sprint(`{"global":{"scrape_interval":"5s"},"scrape_configs":[{"job_name":"perceptor-scrape","scrape_interval":"5s","static_configs":[{"targets":["`, p.PerceptorImageName, `:`, p.PerceptorPort, `","`, p.ScannerImageName, `:`, p.ScannerPort, `","`, p.ImagePerceiverImageName, `:`, p.PerceiverPort, `","`, p.PodPerceiverImageName, `:`, p.PerceiverPort, `","`, p.ImageFacadeImageName, `:`, p.ImageFacadePort, `","skyfire:`, "3005", `"]}]}]}`)},
		"perceptor":             {"perceptor.yaml": fmt.Sprint(`{"HubHost": "`, p.HubHost, `","HubPort": "`, p.HubPort, `","HubUser": "`, p.HubUser, `","HubUserPasswordEnvVar": "`, p.HubUserPasswordEnvVar, `","HubClientTimeoutSeconds": "`, p.HubClientTimeoutPerceptorSeconds, `","ConcurrentScanLimit": "`, p.ConcurrentScanLimit, `","Port": "`, p.PerceptorPort, `","LogLevel": "`, p.LogLevel, `"}`)},
		"perceiver":             {"perceiver.yaml": fmt.Sprint(`{"PerceptorHost": "`, p.PerceptorImageName, `","PerceptorPort": "`, p.PerceptorPort, `","AnnotationIntervalSeconds": "`, p.AnnotationIntervalSeconds, `","DumpIntervalMinutes": "`, p.DumpIntervalMinutes, `","Port": "`, p.PerceiverPort, `","LogLevel": "`, p.LogLevel, `"}`)},
		"perceptor-scanner":     {"perceptor_scanner.yaml": fmt.Sprint(`{"HubHost": "`, p.HubHost, `","HubPort": "`, p.HubPort, `","HubUser": "`, p.HubUser, `","HubUserPasswordEnvVar": "`, p.HubUserPasswordEnvVar, `","HubClientTimeoutSeconds": "`, p.HubClientTimeoutScannerSeconds, `","Port": "`, p.ScannerPort, `","PerceptorHost": "`, p.PerceptorImageName, `","PerceptorPort": "`, p.PerceptorPort, `","ImageFacadeHost": "`, p.ImageFacadeImageName, `","ImageFacadePort": "`, p.ImageFacadePort, `","LogLevel": "`, p.LogLevel, `"}`)},
		"perceptor-imagefacade": {"perceptor_imagefacade.yaml": fmt.Sprint(`{"DockerUser": "`, p.DockerUsername, `","DockerPassword": "`, p.DockerPasswordOrToken, `","Port": "`, p.ImageFacadePort, `","InternalDockerRegistries": `, generateStringFromStringArr(p.InternalDockerRegistries), `,"LogLevel": "`, p.LogLevel, `"}`)},
	}

	if p.PerceptorSkyfire {
		configs["skyfire"] = map[string]string{"skyfire.yaml": fmt.Sprint(`{"UseInClusterConfig": "`, "true", `","Port": "`, "3005", `","HubHost": "`, p.HubHost, `","HubPort": "`, p.HubPort, `","HubUser": "`, p.HubUser, `","HubUserPasswordEnvVar": "`, p.HubUserPasswordEnvVar, `","HubClientTimeoutSeconds": "`, p.HubClientTimeoutScannerSeconds, `","PerceptorHost": "`, p.PerceptorImageName, `","PerceptorPort": "`, p.PerceptorPort, `","KubeDumpIntervalSeconds": "`, "15", `","PerceptorDumpIntervalSeconds": "`, "15", `","HubDumpPauseSeconds": "`, "30", `","ImageFacadePort": "`, p.ImageFacadePort, `","LogLevel": "`, p.LogLevel, `"}`)}
	}

	return configs
}
