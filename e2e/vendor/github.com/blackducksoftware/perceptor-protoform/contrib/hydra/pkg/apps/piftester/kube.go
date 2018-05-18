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

package piftester

import (
	"github.com/blackducksoftware/perceptor-protoform/contrib/hydra/pkg/model"
	"k8s.io/api/core/v1"
)

type Kube struct {
	Config *Config
	// model objects
	PifTester    *model.PifTester
	PodPerceiver *model.PodPerceiver
	ImageFacade  *model.Imagefacade
	Prometheus   *model.Prometheus
	// kubernetes resources
	ReplicationControllers []*v1.ReplicationController
	ConfigMaps             []*v1.ConfigMap
	Services               []*v1.Service
	Secrets                []*v1.Secret
}

func NewKube(config *Config) *Kube {
	kube := &Kube{Config: config}
	kube.createResources()
	return kube
}

func (kube *Kube) createResources() {
	config := kube.Config

	pifTester := model.NewPifTester()
	pifTester.Config = config.PifTesterConfig()

	podPerceiverReplicationCount := 1
	podPerceiver := model.NewPodPerceiver(config.AuxConfig.PodPerceiverServiceAccountName, podPerceiverReplicationCount)
	podPerceiver.Config = config.PodPerceiverConfig()
	podPerceiver.Config.PerceptorHost = pifTester.ServiceName

	imageFacade := model.NewImagefacade(config.AuxConfig.ImageFacadeServiceAccountName)
	imageFacade.Config = config.ImagefacadeConfig()
	imageFacade.PodName = "perceptor-imagefacade"

	prometheus := model.NewPrometheus()
	prometheus.AddTarget(&model.PrometheusTarget{Host: pifTester.ServiceName, Port: config.PifTesterPort})
	prometheus.AddTarget(&model.PrometheusTarget{Host: imageFacade.ServiceName, Port: config.ImageFacadePort})
	prometheus.AddTarget(&model.PrometheusTarget{Host: podPerceiver.ServiceName, Port: config.PodPerceiverPort})

	kube.ReplicationControllers = []*v1.ReplicationController{
		podPerceiver.ReplicationController(),
		pifTester.ReplicationController(),
		imageFacade.ReplicationController(),
	}
	kube.Services = []*v1.Service{
		podPerceiver.Service(),
		pifTester.Service(),
		imageFacade.Service(),
	}
	kube.ConfigMaps = []*v1.ConfigMap{
		podPerceiver.ConfigMap(),
		pifTester.ConfigMap(),
		imageFacade.ConfigMap(),
		prometheus.ConfigMap(),
	}
	kube.Secrets = []*v1.Secret{}

	kube.PifTester = pifTester
	kube.ImageFacade = imageFacade
	kube.PodPerceiver = podPerceiver
}

func (kube *Kube) GetConfigMaps() []*v1.ConfigMap {
	return kube.ConfigMaps
}

func (kube *Kube) GetServices() []*v1.Service {
	return kube.Services
}

func (kube *Kube) GetSecrets() []*v1.Secret {
	return kube.Secrets
}

func (kube *Kube) GetReplicationControllers() []*v1.ReplicationController {
	return kube.ReplicationControllers
}
