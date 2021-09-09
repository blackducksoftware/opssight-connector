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
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/blackducksoftware/perceivers/pkg/annotator"
	"github.com/blackducksoftware/perceivers/pkg/webhook"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// QuayPerceiver handles watching and annotating Images
type QuayPerceiver struct {
	annotator          *annotator.QuayAnnotator
	webhook            *webhook.QuayWebhook
	annotationInterval time.Duration
	dumpInterval       time.Duration
	metricsURL         string
}

// NewQuayPerceiver creates a new ImagePerceiver object
func NewQuayPerceiver(configPath string) (*QuayPerceiver, error) {
	config, err := GetConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %v", err)
	}

	// Configure prometheus for metrics
	prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
	prometheus.Unregister(prometheus.NewGoCollector())
	http.Handle("/metrics", prometheus.Handler())

	// Set log level
	level, err := config.GetLogLevel()
	if err != nil {
		level = log.DebugLevel
	}
	log.SetLevel(level)

	perceptorURL := fmt.Sprintf("http://%s:%d", config.Perceptor.Host, config.Perceptor.Port)
	qp := QuayPerceiver{
		annotator:          annotator.NewQuayAnnotator(perceptorURL, config.PrivateDockerRegistries),
		webhook:            webhook.NewQuayWebhook(perceptorURL, config.PrivateDockerRegistries, config.Perceiver.Certificate, config.Perceiver.CertificateKey),
		annotationInterval: time.Second * time.Duration(config.Perceiver.AnnotationIntervalSeconds),
		dumpInterval:       time.Minute * time.Duration(config.Perceiver.DumpIntervalMinutes),
		metricsURL:         fmt.Sprintf(":%d", config.Perceiver.Port),
	}
	return &qp, nil
}

// Run starts the QuayPerceiver watching and annotating Images
func (qp *QuayPerceiver) Run(stopCh <-chan struct{}) {
	log.Infof("starting quay controllers")
	go qp.annotator.Run(qp.annotationInterval, stopCh)
	go qp.webhook.Run()
	<-stopCh
}
