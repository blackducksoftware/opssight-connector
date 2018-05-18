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

package scannertester

import (
	"github.com/blackducksoftware/perceptor-protoform/contrib/hydra/pkg/model"
	"github.com/spf13/viper"
)

type Config struct {
	MasterURL      string
	KubeConfigPath string

	HubHost         string
	HubUser         string
	HubUserPassword string
	HubPort         int32

	ScannerMemory          string
	JavaInitialHeapSizeMBs int
	JavaMaxHeapSizeMBs     int

	PerceptorHost   string
	PerceptorPort   int32
	ImageFacadePort int32
	ScannerPort     int32

	// Secret config
	HubPasswordSecretName string
	HubPasswordSecretKey  string

	LogLevel string

	AuxConfig *AuxiliaryConfig
}

func ReadConfig(configPath string) *Config {
	viper.SetConfigFile(configPath)
	config := &Config{}
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	viper.Unmarshal(config)
	return config
}

func (config *Config) ScannerConfig() model.ScannerConfigMap {
	return model.ScannerConfigMap{
		HubClientTimeoutSeconds: 120,
		HubHost:                 config.HubHost,
		HubUser:                 config.HubUser,
		HubUserPasswordEnvVar:   "SCANNER_HUBUSERPASSWORD",
		HubPort:                 config.HubPort,
		JavaInitialHeapSizeMBs:  config.JavaInitialHeapSizeMBs,
		JavaMaxHeapSizeMBs:      config.JavaMaxHeapSizeMBs,
		ImageFacadePort:         config.ImageFacadePort,
		PerceptorHost:           config.PerceptorHost,
		PerceptorPort:           config.PerceptorPort,
		Port:                    config.ScannerPort,
		LogLevel:                config.LogLevel,
	}
}

func (config *Config) MockImagefacadeConfig() model.MockImagefacadeConfigMap {
	return model.MockImagefacadeConfigMap{
		Port: config.ImageFacadePort,
	}
}

func (config *Config) PerceptorConfig() model.PerceptorConfigMap {
	return model.PerceptorConfigMap{
		Port:                  config.PerceptorPort,
		UseMockMode:           true,
		LogLevel:              config.LogLevel,
		HubUserPasswordEnvVar: "SCANNER_HUBPASSWORD",
	}
}
