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

package imagefacade

import (
	"strings"

	"github.com/blackducksoftware/perceptor-scanner/pkg/docker"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// ImageFacadeConfig ...
type ImageFacadeConfig struct {
	// These allow images to be pulled from registries that require authentication
	PrivateDockerRegistries []docker.RegistryAuth

	CreateImagesOnly bool
	Port             int
}

// Config ...
type Config struct {
	LogLevel    string
	ImageFacade ImageFacadeConfig
}

// GetLogLevel ...
func (config *Config) GetLogLevel() (log.Level, error) {
	return log.ParseLevel(config.LogLevel)
}

// GetConfig returns a configuration object
func GetConfig(configPath string) (*Config, error) {
	var config *Config

	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.SetEnvPrefix("PCP")
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

		viper.BindEnv("ImageFacade_PrivateDockerRegistries")
		viper.BindEnv("ImageFacade_Port")
		viper.BindEnv("ImageFacade_CreateImagesOnly")
		viper.BindEnv("LogLevel")

		viper.AutomaticEnv()
	}

	err := viper.ReadInConfig()
	if err != nil {
		return nil, errors.Annotate(err, "failed to ReadInConfig")
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, errors.Annotate(err, "failed to unmarshal config")
	}

	return config, nil
}
