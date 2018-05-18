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

	"github.com/blackducksoftware/perceptor-scanner/pkg/imagefacade"
	"github.com/prometheus/client_golang/prometheus"

	log "github.com/sirupsen/logrus"
)

// PerceptorImageFacade handles retrieving images to scan
type PerceptorImageFacade struct {
	imageFacade *imagefacade.ImageFacade
	config      *imagefacade.Config
}

// NewPerceptorImageFacade creates a new PerceptorImageFacade
func NewPerceptorImageFacade(configPath string) (*PerceptorImageFacade, error) {
	config, err := imagefacade.GetConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to load configuration: %v", err)
	}

	level, err := config.GetLogLevel()
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	log.SetLevel(level)

	prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
	prometheus.Unregister(prometheus.NewGoCollector())

	imageFacade := imagefacade.NewImageFacade(
		config.DockerUser,
		config.DockerPassword,
		config.InternalDockerRegistries,
		config.CreateImagesOnly,
	)

	return &PerceptorImageFacade{imageFacade: imageFacade, config: config}, nil
}

// Run starts the PerceptorImageFacade listening for requests
func (pif *PerceptorImageFacade) Run(stopCh chan struct{}) {
	log.Infof("successfully instantiated imagefacade -- %+v", pif.imageFacade)

	addr := fmt.Sprintf(":%d", pif.config.Port)
	log.Infof("starting HTTP server on %s", addr)
	http.ListenAndServe(addr, nil)

	<-stopCh
}
