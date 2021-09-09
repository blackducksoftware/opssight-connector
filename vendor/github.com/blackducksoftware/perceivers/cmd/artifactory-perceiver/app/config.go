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

package app

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/blackducksoftware/perceivers/pkg/utils"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// PerceptorConfig contains Perceptor config
type PerceptorConfig struct {
	Host string
	Port int
}

// ArtifactoryPerceiverConfig contains config specific to pod perceivers
type ArtifactoryPerceiverConfig struct {
	Dumper bool
}

// PerceiverConfig contains general Perceiver config
type PerceiverConfig struct {
	Certificate               string
	CertificateKey            string
	AnnotationIntervalSeconds int
	DumpIntervalMinutes       int
	Port                      int
	Artifactory               ArtifactoryPerceiverConfig
}

// Config contains the ArtifactoryPerceiver configurations
type Config struct {
	LogLevel                string
	Perceptor               PerceptorConfig
	Perceiver               PerceiverConfig
	PrivateDockerRegistries []*utils.RegistryAuth
}

// GetConfig returns a configuration object to configure a ArtifactoryPerceiver
func GetConfig(configPath string) (*Config, error) {
	var cfg *Config

	viper.SetConfigFile(configPath)

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	err = cfg.getPrivateDockerRegistries()
	if err != nil {
		return nil, fmt.Errorf("failed to find private docker repo credentials: %v", err)
	}

	return cfg, nil
}

// GetLogLevel returns the log level set in Opssight Spec Config
func (config *Config) GetLogLevel() (log.Level, error) {
	return log.ParseLevel(config.LogLevel)
}

// StartWatch will start watching the ArtifactoryPerceiver configuration file and
// call the passed handler function when the configuration file has changed
func (config *Config) StartWatch(handler func(fsnotify.Event)) {
	viper.WatchConfig()
	viper.OnConfigChange(handler)
}

// getPrivateDockerRegistries will get the private Docker registries credential
func (config *Config) getPrivateDockerRegistries() error {
	credentials, ok := os.LookupEnv("securedRegistries.json")
	if !ok {
		return fmt.Errorf("cannot find Private Docker Registries: environment variable securedRegistries not found")
	}

	privateDockerRegistries := map[string]*utils.RegistryAuth{}
	err := json.Unmarshal([]byte(credentials), &privateDockerRegistries)
	if err != nil {
		return fmt.Errorf("unable to unmarshall Private Docker registries due to %+v", err)
	}

	dockerRegistries := []*utils.RegistryAuth{}
	for _, privatedockerRegistry := range privateDockerRegistries {
		dockerRegistries = append(dockerRegistries, privatedockerRegistry)
	}

	config.PrivateDockerRegistries = dockerRegistries

	return nil
}
