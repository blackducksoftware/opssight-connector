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

// PodPerceiverConfig contains all configuration for a PodPerceiver
type PodPerceiverConfig struct {
	PerceptorHost             string
	PerceptorPort             int
	AnnotationIntervalSeconds int
	DumpIntervalMinutes       int
	Port                      int
}

// GetPodPerceiverConfig returns a configuration object to configure a PodPerceiver
func GetPodPerceiverConfig() (*PodPerceiverConfig, error) {
	var cfg *PodPerceiverConfig

	viper.SetConfigName("perceiver")
	viper.AddConfigPath("/etc/perceiver")

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
func (p *PodPerceiverConfig) StartWatch(handler func(fsnotify.Event)) {
	viper.WatchConfig()
	viper.OnConfigChange(handler)
}
