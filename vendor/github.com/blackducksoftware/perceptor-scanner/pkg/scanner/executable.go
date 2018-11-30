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
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// RunScanner ...
func RunScanner(configPath string, stop <-chan struct{}) {
	config, err := GetConfig(configPath)
	if err != nil {
		panic(fmt.Errorf("Failed to load configuration: %v", err))
	}

	level, err := config.GetLogLevel()
	if err != nil {
		panic(err)
	}
	log.SetLevel(level)

	prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
	prometheus.Unregister(prometheus.NewGoCollector())

	manager, err := NewManager(config, stop)
	if err != nil {
		panic(err)
	}
	manager.StartRequestingScanJobs()

	http.Handle("/metrics", prometheus.Handler())

	addr := fmt.Sprintf(":%d", config.Scanner.Port)
	log.Infof("successfully instantiated manager %+v, serving on %s", manager, addr)
	go func() {
		http.ListenAndServe(addr, nil)
	}()

	<-stop
}
