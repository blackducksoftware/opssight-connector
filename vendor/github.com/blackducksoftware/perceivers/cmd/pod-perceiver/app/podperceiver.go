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

	"github.com/blackducksoftware/perceivers/pkg/annotations"
	"github.com/blackducksoftware/perceivers/pkg/annotator"
	"github.com/blackducksoftware/perceivers/pkg/controller"
	"github.com/blackducksoftware/perceivers/pkg/dumper"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
)

// PodPerceiver handles watching and annotating pods
type PodPerceiver struct {
	podController *controller.PodController

	podAnnotator       *annotator.PodAnnotator
	annotationInterval time.Duration

	podDumper    *dumper.PodDumper
	dumpInterval time.Duration

	metricsURL string
}

// NewPodPerceiver creates a new PodPerceiver object
func NewPodPerceiver(handler annotations.PodAnnotatorHandler, configPath string) (*PodPerceiver, error) {
	config, err := GetPodPerceiverConfig(configPath)
	if err != nil {
		panic(fmt.Errorf("failed to read config: %v", err))
	}

	// Create a kube client from in cluster configuration
	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to build config from cluster: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create kubernetes client: %v", err)
	}

	// Configure prometheus for metrics
	prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
	prometheus.Unregister(prometheus.NewGoCollector())
	http.Handle("/metrics", prometheus.Handler())

	perceptorURL := fmt.Sprintf("http://%s:%d", config.PerceptorHost, config.PerceptorPort)
	p := PodPerceiver{
		podController:      controller.NewPodController(clientset, perceptorURL, config.RequireLabel, handler),
		podAnnotator:       annotator.NewPodAnnotator(clientset.CoreV1(), perceptorURL, handler),
		annotationInterval: time.Second * time.Duration(config.AnnotationIntervalSeconds),
		podDumper:          dumper.NewPodDumper(clientset.CoreV1(), perceptorURL, config.RequireLabel),
		dumpInterval:       time.Minute * time.Duration(config.DumpIntervalMinutes),
		metricsURL:         fmt.Sprintf(":%d", config.Port),
	}

	return &p, nil
}

// Run starts the PodPerceiver watching and annotating pods
func (pp *PodPerceiver) Run(stopCh <-chan struct{}) {
	log.Infof("starting pod controllers")
	go pp.podController.Run(5, stopCh)
	go pp.podAnnotator.Run(pp.annotationInterval, stopCh)
	go pp.podDumper.Run(pp.dumpInterval, stopCh)

	log.Infof("starting prometheus on %s", pp.metricsURL)
	http.ListenAndServe(pp.metricsURL, nil)

	<-stopCh
}
