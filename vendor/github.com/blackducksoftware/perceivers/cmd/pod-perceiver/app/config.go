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

package app

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// PerceptorConfig contains Perceptor config
type PerceptorConfig struct {
	Host string
	Port int
}

// PodPerceiverConfig contains config specific to pod perceivers
type PodPerceiverConfig struct {
	NamespaceFilter string
}

// PerceiverConfig contains general Perceiver config
type PerceiverConfig struct {
	AnnotationIntervalSeconds int
	DumpIntervalMinutes       int
	Port                      int
	Pod                       PodPerceiverConfig
}

// Config contains all configuration for a PodPerceiver
type Config struct {
	Perceptor PerceptorConfig
	Perceiver PerceiverConfig
}

// GetConfig returns a configuration object to configure a PodPerceiver
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
	return cfg, nil
}

// StartWatch will start watching the PodPerceiver configuration file and
// call the passed handler function when the configuration file has changed
func (p *Config) StartWatch(handler func(fsnotify.Event)) {
	viper.WatchConfig()
	viper.OnConfigChange(handler)
}
