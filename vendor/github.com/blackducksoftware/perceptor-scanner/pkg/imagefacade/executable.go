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

package imagefacade

import (
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"

	log "github.com/sirupsen/logrus"
)

func RunImageFacade(configPath string, stop <-chan struct{}) {
	config, err := GetConfig(configPath)
	if err != nil {
		log.Errorf("unable to read config: %s", err.Error())
		panic(err)
	}

	level, err := config.GetLogLevel()
	if err != nil {
		log.Errorf("unable to set log level: %s", err.Error())
		panic(err)
	}
	log.SetLevel(level)

	prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
	prometheus.Unregister(prometheus.NewGoCollector())

	imageFacade := NewImageFacade(config.PrivateDockerRegistries, config.CreateImagesOnly)

	log.Infof("successfully instantiated imagefacade -- %+v", imageFacade)

	addr := fmt.Sprintf(":%d", config.Port)
	log.Infof("starting HTTP server on %s", addr)
	go func() {
		http.ListenAndServe(addr, nil)
	}()

	<-stop
}
