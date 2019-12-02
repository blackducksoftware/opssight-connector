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
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config type will be used for protoform config that bootstraps everything
type Config struct {
	DryRun                        bool
	LogLevel                      string
	Namespace                     string
	CrdNamespace                  string
	Threadiness                   int
	PostgresRestartInMins         int64
	HubFederatorConfig            *HubFederatorConfig
	PodWaitTimeoutSeconds         int64
	ResyncIntervalInSeconds       int64
	TerminationGracePeriodSeconds int64
	AdmissionWebhookListener      bool
	CrdNames                      string
	IsClusterScoped               bool
	IsOpenshift                   bool
	Version                       string
}

// SelfSetDefaults ...
func (config *Config) SelfSetDefaults() {
	config.HubFederatorConfig = &HubFederatorConfig{}
	config.HubFederatorConfig.HubConfig = &HubConfig{}
}

// HubFederatorConfig will have the configuration related to hub federator
type HubFederatorConfig struct {
	HubConfig    *HubConfig
	UseMockMode  bool
	Port         int
	Registry     string
	ImagePath    string
	ImageName    string
	ImageVersion string
}

// HubConfig will have the configuration related to Blackduck
type HubConfig struct {
	User                         string
	PasswordEnvVar               string
	ClientTimeoutMilliseconds    int
	Port                         int
	FetchAllProjectsPauseSeconds int
}

// GetLogLevel will set the log level
func (config *Config) GetLogLevel() (log.Level, error) {
	return log.ParseLevel(config.LogLevel)
}

// GetConfig will read the config file and set in the Viper
func GetConfig(configPath string, version string) (*Config, error) {
	var config *Config

	if len(configPath) > 0 {
		viper.SetConfigFile(configPath)
		err := viper.ReadInConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %v", err)
		}
	} else {
		viper.SetEnvPrefix("SO")
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		viper.BindEnv("DryRun")
		viper.BindEnv("LogLevel")
		viper.BindEnv("Namespace")
		viper.BindEnv("Threadiness")
		viper.BindEnv("PostgresRestartInMins")
		viper.BindEnv("PodWaitTimeoutSeconds")
		viper.BindEnv("ResyncIntervalInSeconds")
		viper.BindEnv("TerminationGracePeriodSeconds")
		viper.BindEnv("AdmissionWebhookListener")
		viper.BindEnv("CrdNames")
		viper.BindEnv("IsClusterScoped")
		viper.AutomaticEnv()
	}

	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// set the operator version
	config.Version = version

	if !config.IsClusterScoped {
		config.CrdNamespace = config.Namespace
	}

	return config, nil
}
