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
	"net/http"
	"os"

	"github.com/blackducksoftware/perceptor-scanner/pkg/scanner"

	"github.com/prometheus/client_golang/prometheus"

	log "github.com/sirupsen/logrus"
)

// PerceptorScanner handles scanning containers
type PerceptorScanner struct {
	scannerManager *scanner.Scanner
	config         *scanner.Config
}

// NewPerceptorScanner creates a new PerceptorScanner object
func NewPerceptorScanner(configPath string) (*PerceptorScanner, error) {
	config, err := scanner.GetConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to load configuration: %v", err)
	}

	level, err := config.GetLogLevel()
	if err != nil {
		return nil, err
	}
	log.SetLevel(level)

	sm, err := scanner.NewScanner(config)
	if err != nil {
		return nil, fmt.Errorf("unable to instantiate scanner: %v", err)
	}

	return &PerceptorScanner{scannerManager: sm, config: config}, nil
}

// Run starts the PerceptorScanner looking for scan jobs
func (ps *PerceptorScanner) Run(stopCh <-chan struct{}) {
	prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
	prometheus.Unregister(prometheus.NewGoCollector())

	ps.scannerManager.StartRequestingScanJobs()

	http.Handle("/metrics", prometheus.Handler())

	addr := fmt.Sprintf(":%d", ps.config.Port)
	log.Infof("successfully instantiated scanner %+v, serving on %s", ps.scannerManager, addr)
	http.ListenAndServe(addr, nil)

	<-stopCh
}
