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
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	HubHost                 string
	HubUser                 string
	HubUserPasswordEnvVar   string
	HubPort                 int
	HubClientTimeoutSeconds int

	LogLevel string
	Port     int

	ImageFacadePort int

	PerceptorHost string
	PerceptorPort int
}

func (config *Config) GetLogLevel() (log.Level, error) {
	return log.ParseLevel(config.LogLevel)
}

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

	// Ports must be reachable
	if config.Port == 0 || config.PerceptorPort == 0 || config.ImageFacadePort == 0 {
		err = fmt.Errorf("Need non zero numbers for Port (got %d), PerceptorPort (got %d), HubPort (got %d), and ImageFacadePort (got %d)",
			config.Port,
			config.PerceptorPort,
			config.HubPort,
			config.ImageFacadePort)
	}
	return config, err
}
