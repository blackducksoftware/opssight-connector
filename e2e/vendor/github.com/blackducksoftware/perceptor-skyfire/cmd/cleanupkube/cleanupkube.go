/*
Copyright (C) 2018 Black Duck Software, Inc.

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

package main

import (
	"os"

	kube "github.com/blackducksoftware/perceptor-skyfire/pkg/kube"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	configPath := os.Args[1]
	config := ReadConfig(configPath)
	logLevel, err := config.GetLogLevel()
	if err != nil {
		panic(err)
	}
	log.SetLevel(logLevel)
	kubeClient, err := kube.NewKubeClient(config.KubeClientConfig())
	if err != nil {
		panic(err)
	}
	err = kubeClient.CleanupAllPods()
	if err != nil {
		panic(err)
	}
}

type Config struct {
	UseInClusterConfig bool
	MasterURL          string
	KubeConfigPath     string
	LogLevel           string
}

func (config *Config) KubeClientConfig() *kube.KubeClientConfig {
	if config.UseInClusterConfig {
		return nil
	}
	return &kube.KubeClientConfig{
		KubeConfigPath: config.KubeConfigPath,
		MasterURL:      config.MasterURL,
	}
}

func (config *Config) GetLogLevel() (log.Level, error) {
	return log.ParseLevel(config.LogLevel)
}

func ReadConfig(configPath string) *Config {
	viper.SetConfigFile(configPath)
	config := &Config{}
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	err = viper.Unmarshal(config)
	if err != nil {
		panic(err)
	}
	return config
}
