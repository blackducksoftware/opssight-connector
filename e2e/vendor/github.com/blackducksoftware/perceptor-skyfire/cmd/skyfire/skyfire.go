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
	"fmt"
	"net/http"
	"os"

	skyfire "github.com/blackducksoftware/perceptor-skyfire/pkg/skyfire"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

func main() {
	configPath := os.Args[1]
	fmt.Printf("Config path: %s", configPath)
	config, err := skyfire.ReadConfig(configPath)
	if err != nil {
		panic(err)
	}

	logLevel, err := config.GetLogLevel()
	if err != nil {
		panic(err)
	}
	log.SetLevel(logLevel)

	log.Infof("received config %+v", config)

	skyfire, err := skyfire.NewSkyfire(config)
	if err != nil {
		panic(err)
	}
	log.Infof("instantiated skyfire: %+v", skyfire)

	http.Handle("/metrics", prometheus.Handler())
	addr := fmt.Sprintf(":%d", config.Port)
	go http.ListenAndServe(addr, nil)
	log.Infof("running http server on %s", addr)

	select {}
}
