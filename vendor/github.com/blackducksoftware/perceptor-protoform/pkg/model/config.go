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

package model

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config type will used for protoform config
type Config struct {
	DryRun             bool
	LogLevel           string
	Namespace          string
	Threadiness        int
	HubFederatorConfig HubFederatorConfig
}

// HubFederatorConfig will have the configuration related to hub federator
type HubFederatorConfig struct {
	HubConfig   HubConfig
	UseMockMode bool
	Port        int
}

// HubConfig will have the configuration related to Hub
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
func GetConfig(configPath string) (*Config, error) {
	var config *Config

	viper.SetConfigFile(configPath)

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	return config, nil
}
