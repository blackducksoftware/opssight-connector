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
	"time"

	"github.com/blackducksoftware/perceivers/image/pkg/annotator"
	"github.com/blackducksoftware/perceivers/image/pkg/controller"
	"github.com/blackducksoftware/perceivers/image/pkg/dumper"
	"github.com/blackducksoftware/perceivers/pkg/annotations"

	"k8s.io/client-go/rest"

	imagev1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
)

// ImagePerceiver handles watching and annotating Images
type ImagePerceiver struct {
	client *imagev1.ImageV1Client

	ImageController *controller.ImageController

	ImageAnnotator     *annotator.ImageAnnotator
	annotationInterval time.Duration

	ImageDumper  *dumper.ImageDumper
	dumpInterval time.Duration

	metricsURL string
}

// NewImagePerceiver creates a new ImagePerceiver object
func NewImagePerceiver(handler annotations.ImageAnnotatorHandler) (*ImagePerceiver, error) {
	config, err := GetImagePerceiverConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %v", err)
	}

	// Create a kube client from in cluster configuration
	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to build config from cluster: %v", err)
	}
	imageClient, err := imagev1.NewForConfig(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create image client: %v", err)
	}

	// Configure prometheus for metrics
	prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
	prometheus.Unregister(prometheus.NewGoCollector())
	http.Handle("/metrics", prometheus.Handler())

	perceptorURL := fmt.Sprintf("http://%s:%d", config.PerceptorHost, config.PerceptorPort)
	p := ImagePerceiver{
		ImageController:    controller.NewImageController(imageClient, perceptorURL, handler),
		ImageAnnotator:     annotator.NewImageAnnotator(imageClient, perceptorURL, handler),
		annotationInterval: time.Second * time.Duration(config.AnnotationIntervalSeconds),
		ImageDumper:        dumper.NewImageDumper(imageClient, perceptorURL),
		dumpInterval:       time.Minute * time.Duration(config.DumpIntervalMinutes),
		metricsURL:         fmt.Sprintf(":%d", config.Port),
	}

	return &p, nil
}

// Run starts the ImagePerceiver watching and annotating Images
func (ip *ImagePerceiver) Run(stopCh <-chan struct{}) {
	log.Infof("starting image controllers")
	go ip.ImageController.Run(5, stopCh)
	go ip.ImageAnnotator.Run(ip.annotationInterval, stopCh)
	go ip.ImageDumper.Run(ip.dumpInterval, stopCh)

	log.Infof("starting prometheus on %d", ip.metricsURL)
	http.ListenAndServe(ip.metricsURL, nil)

	<-stopCh
}
