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

package standardperceptor

import (
	model "github.com/blackducksoftware/perceptor-protoform/contrib/hydra/pkg/model"
	"github.com/spf13/viper"
)

type Config struct {
	// general protoform config
	MasterURL      string
	KubeConfigPath string

	// perceptor config
	HubHost              string
	HubUser              string
	HubUserPassword      string
	HubPort              int32
	ConcurrentScanLimit  int
	PerceptorPort        int32
	UseMockPerceptorMode bool

	// Perceivers config
	ImagePerceiverPort           int32
	PodPerceiverPort             int32
	PodPerceiverReplicationCount int
	AnnotationIntervalSeconds    int
	DumpIntervalMinutes          int

	// Scanner config
	ScannerReplicationCount int32
	ScannerPort             int32
	ImageFacadePort         int32
	ScannerMemory           string
	JavaInitialHeapSizeMBs  int
	JavaMaxHeapSizeMBs      int

	// Skyfire config
	SkyfirePort int32

	// Secret config
	HubPasswordSecretName string
	HubPasswordSecretKey  string

	LogLevel string

	AuxConfig *AuxiliaryConfig
}

func ReadConfig(configPath string) *Config {
	viper.SetConfigFile(configPath)
	pc := &Config{}
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	viper.Unmarshal(pc)
	return pc
}

func (pc *Config) PodPerceiverConfig() model.PodPerceiverConfigMap {
	return model.PodPerceiverConfigMap{
		AnnotationIntervalSeconds: pc.AnnotationIntervalSeconds,
		DumpIntervalMinutes:       pc.DumpIntervalMinutes,
		PerceptorHost:             "", // must be filled in elsewhere
		PerceptorPort:             pc.PerceptorPort,
		Port:                      pc.PodPerceiverPort,
	}
}

func (pc *Config) ImagePerceiverConfig() model.ImagePerceiverConfigMap {
	return model.ImagePerceiverConfigMap{
		AnnotationIntervalSeconds: pc.AnnotationIntervalSeconds,
		DumpIntervalMinutes:       pc.DumpIntervalMinutes,
		PerceptorHost:             "", // must be filled in elsewhere
		PerceptorPort:             pc.PerceptorPort,
		Port:                      pc.ImagePerceiverPort,
	}
}

func (pc *Config) ScannerConfig() model.ScannerConfigMap {
	return model.ScannerConfigMap{
		HubHost:                 pc.HubHost,
		HubUser:                 pc.HubUser,
		HubUserPasswordEnvVar:   "SCANNER_HUBUSERPASSWORD",
		HubPort:                 pc.HubPort,
		HubClientTimeoutSeconds: 120,
		JavaInitialHeapSizeMBs:  pc.JavaInitialHeapSizeMBs,
		JavaMaxHeapSizeMBs:      pc.JavaMaxHeapSizeMBs,
		LogLevel:                pc.LogLevel,
		Port:                    pc.ScannerPort,
		PerceptorHost:           "", // must be filled in elsewhere
		PerceptorPort:           pc.PerceptorPort,
		ImageFacadePort:         pc.ImageFacadePort,
	}
}

func (pc *Config) ImagefacadeConfig() model.ImagefacadeConfigMap {
	return model.ImagefacadeConfigMap{
		DockerPassword:           pc.AuxConfig.DockerPassword,
		DockerUser:               pc.AuxConfig.DockerUsername,
		InternalDockerRegistries: pc.AuxConfig.InternalDockerRegistries,
		CreateImagesOnly:         false,
		Port:                     pc.ImageFacadePort,
		LogLevel:                 pc.LogLevel,
	}
}

func (pc *Config) PerceptorConfig() model.PerceptorConfigMap {
	return model.PerceptorConfigMap{
		ConcurrentScanLimit:   pc.ConcurrentScanLimit,
		HubHost:               pc.HubHost,
		HubUser:               pc.HubUser,
		HubUserPasswordEnvVar: "PERCEPTOR_HUBUSERPASSWORD",
		HubPort:               int(pc.HubPort),
		UseMockMode:           pc.UseMockPerceptorMode,
		Port:                  pc.PerceptorPort,
		LogLevel:              pc.LogLevel,
	}
}

func (pc *Config) SkyfireConfig() model.SkyfireConfigMap {
	return model.SkyfireConfigMap{
		HubHost:     pc.HubHost,
		HubUser:     pc.HubUser,
		HubPassword: pc.HubUserPassword,
		// TODO pc.HubPort ?
		KubeDumpIntervalSeconds:      15,
		PerceptorDumpIntervalSeconds: 15,
		HubDumpPauseSeconds:          30,
		LogLevel:                     pc.LogLevel,
		PerceptorHost:                "", // must be filled in elsewhere
		PerceptorPort:                pc.PerceptorPort,
		Port:                         pc.SkyfirePort,
		UseInClusterConfig:           true,
	}
}
