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

package scanner

import (
	"strings"

	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// BlackDuckConfig stores the Black Duck configuration
type BlackDuckConfig struct {
	ConnectionsEnvironmentVariableName string
	TLSVerification                    bool
}

// ImageFacadeConfig stores the image facade configuration
type ImageFacadeConfig struct {
	Host string
	Port int
}

// GetHost return the image facade host
func (ifc *ImageFacadeConfig) GetHost() string {
	if ifc.Host == "" {
		return "localhost"
	}
	return ifc.Host
}

// PerceptorConfig stores the perceptor client configuration
type PerceptorConfig struct {
	Host string
	Port int
}

// ScannerConfig stores the scanner configuration
type ScannerConfig struct {
	ImageDirectory       string
	Port                 int
	ClientTimeoutSeconds int
}

// Config stores the input scanner configurqtion
type Config struct {
	BlackDuck   *BlackDuckConfig
	ImageFacade *ImageFacadeConfig
	Perceptor   *PerceptorConfig
	Scanner     *ScannerConfig

	LogLevel string
}

// GetImageDirectory return the image directory to store the pulled artifactory
func (config *ScannerConfig) GetImageDirectory() string {
	if config.ImageDirectory == "" {
		return "/var/images"
	}
	return config.ImageDirectory
}

// GetLogLevel return the log level
func (config *Config) GetLogLevel() (log.Level, error) {
	return log.ParseLevel(config.LogLevel)
}

// GetConfig returns the input configuration for Scanner pod
func GetConfig(configPath string) (*Config, error) {
	var config *Config

	if configPath != "" {
		viper.SetConfigFile(configPath)
		err := viper.ReadInConfig()
		if err != nil {
			return nil, errors.Annotatef(err, "failed to ReadInConfig")
		}
	} else {
		viper.SetEnvPrefix("PCP")
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

		viper.BindEnv("ImageFacade.Host")
		viper.BindEnv("ImageFacade.Port")

		viper.BindEnv("Perceptor.Host")
		viper.BindEnv("Perceptor.Port")

		viper.BindEnv("BlackDuck.ConnectionsEnvironmentVariableName")
		viper.BindEnv("BlackDuck.TLSVerification")

		viper.BindEnv("Scanner.Port")
		viper.BindEnv("Scanner.ImageDirectory")
		viper.BindEnv("Scanner.HubClientTimeoutSeconds")

		viper.BindEnv("LogLevel")

		viper.AutomaticEnv()
	}

	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to unmarshal config")
	}

	return config, nil
}
