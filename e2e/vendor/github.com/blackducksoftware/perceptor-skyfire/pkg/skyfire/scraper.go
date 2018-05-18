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

package skyfire

import (
	"os"
	"time"

	"github.com/blackducksoftware/perceptor-skyfire/pkg/hub"
	"github.com/blackducksoftware/perceptor-skyfire/pkg/kube"
	"github.com/blackducksoftware/perceptor-skyfire/pkg/perceptor"
	log "github.com/sirupsen/logrus"
)

type Scraper struct {
	KubeDumper                   *kube.KubeClient
	KubeDumps                    chan *kube.Dump
	KubeDumpIntervalSeconds      int
	PerceptorDumper              *perceptor.PerceptorDumper
	PerceptorDumps               chan *perceptor.Dump
	PerceptorDumpIntervalSeconds int
	HubDumper                    *hub.HubDumper
	HubDumps                     chan *hub.Dump
	HubDumpPauseSeconds          int
}

func NewScraper(config *Config) (*Scraper, error) {
	kubeDumper, err := kube.NewKubeClient(config.KubeClientConfig())
	if err != nil {
		return nil, err
	}

	perceptorDumper := perceptor.NewPerceptorDumper(config.PerceptorHost, config.PerceptorPort)

	hubDumper, err := hub.NewHubDumper(config.HubHost, config.HubUser, os.Getenv(config.HubUserPasswordEnvVar))
	if err != nil {
		return nil, err
	}

	scraper := &Scraper{
		KubeDumper:                   kubeDumper,
		KubeDumps:                    make(chan *kube.Dump),
		KubeDumpIntervalSeconds:      config.KubeDumpIntervalSeconds,
		PerceptorDumper:              perceptorDumper,
		PerceptorDumps:               make(chan *perceptor.Dump),
		PerceptorDumpIntervalSeconds: config.PerceptorDumpIntervalSeconds,
		HubDumper:                    hubDumper,
		HubDumps:                     make(chan *hub.Dump),
		HubDumpPauseSeconds:          config.HubDumpPauseSeconds,
	}

	scraper.StartScraping()

	return scraper, nil
}

func (sc *Scraper) StartHubScrapes() {
	for {
		hubDump, err := sc.HubDumper.Dump()
		if err == nil {
			sc.HubDumps <- hubDump
			recordEvent("hub dump")
		} else {
			recordError("unable to get perceptor dump")
			log.Errorf("unable to get hub dump: %s", err.Error())
		}
		time.Sleep(time.Duration(sc.HubDumpPauseSeconds) * time.Second)
	}
}

func (sc *Scraper) StartKubeScrapes() {
	for {
		kubeDump, err := sc.KubeDumper.Dump()
		if err == nil {
			sc.KubeDumps <- kubeDump
			recordEvent("kube dump")
		} else {
			recordError("unable to get kube dump")
			log.Errorf("unable to get kube dump: %s", err.Error())
		}
		time.Sleep(time.Duration(sc.KubeDumpIntervalSeconds) * time.Second)
	}
}

func (sc *Scraper) StartPerceptorScrapes() {
	for {
		perceptorDump, err := sc.PerceptorDumper.Dump()
		if err == nil {
			sc.PerceptorDumps <- perceptorDump
			recordEvent("perceptor dump")
		} else {
			recordError("unable to get perceptor dump")
			log.Errorf("unable to get perceptor dump: %s", err.Error())
		}
		time.Sleep(time.Duration(sc.PerceptorDumpIntervalSeconds) * time.Second)
	}
}

func (sc *Scraper) StartScraping() {
	go sc.StartHubScrapes()
	go sc.StartKubeScrapes()
	go sc.StartPerceptorScrapes()
}
