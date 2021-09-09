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
	"github.com/blackducksoftware/perceivers/pkg/controller"
	"github.com/blackducksoftware/perceivers/pkg/webhook"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// ArtifactoryPerceiver handles watching and annotating Images
type ArtifactoryPerceiver struct {
	controller         *controller.ArtifactoryController
	annotator          *annotator.ArtifactoryAnnotator
	webhook            *webhook.ArtifactoryWebhook
	annotationInterval time.Duration
	dumpInterval       time.Duration
	metricsURL         string
	dumper             bool
}

// NewArtifactoryPerceiver creates a new ArtifactoryPerceiver object
func NewArtifactoryPerceiver(configPath string) (*ArtifactoryPerceiver, error) {
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
	ap := ArtifactoryPerceiver{
		controller:         controller.NewArtifactoryController(perceptorURL, config.PrivateDockerRegistries),
		annotator:          annotator.NewArtifactoryAnnotator(perceptorURL, config.PrivateDockerRegistries),
		webhook:            webhook.NewArtifactoryWebhook(perceptorURL, config.PrivateDockerRegistries, config.Perceiver.Certificate, config.Perceiver.CertificateKey),
		annotationInterval: time.Second * time.Duration(config.Perceiver.AnnotationIntervalSeconds),
		dumpInterval:       time.Minute * time.Duration(config.Perceiver.DumpIntervalMinutes),
		metricsURL:         fmt.Sprintf(":%d", config.Perceiver.Port),
		dumper:             config.Perceiver.Artifactory.Dumper,
	}
	return &ap, nil
}

// Run starts the ArtifactoryPerceiver watching and annotating Images
func (ap *ArtifactoryPerceiver) Run(stopCh <-chan struct{}) {
	log.Infof("starting artifactory controllers")
	// Only run if config set
	if ap.dumper {
		go ap.controller.Run(ap.dumpInterval, stopCh)
	}
	go ap.annotator.Run(ap.annotationInterval, stopCh)
	go ap.webhook.Run()
	<-stopCh
}
