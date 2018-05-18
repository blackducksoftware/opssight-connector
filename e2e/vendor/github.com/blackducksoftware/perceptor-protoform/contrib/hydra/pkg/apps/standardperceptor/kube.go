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

package standardperceptor

import (
	"github.com/blackducksoftware/perceptor-protoform/contrib/hydra/pkg/model"
	"k8s.io/api/core/v1"
	// v1beta1 "k8s.io/api/extensions/v1beta1"

	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Kube struct {
	Config *Config
	// model objects
	Perceptor    *model.Perceptor
	PodPerceiver *model.PodPerceiver
	Scanner      *model.Scanner
	ImageFacade  *model.Imagefacade
	ScannerPod   *model.ScannerPod
	Prometheus   *model.Prometheus
	Skyfire      *model.Skyfire
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

	perceptor := model.NewPerceptor()
	perceptor.Config = config.PerceptorConfig()
	perceptor.HubPasswordSecretName = config.HubPasswordSecretName
	perceptor.HubPasswordSecretKey = config.HubPasswordSecretKey

	podPerceiver := model.NewPodPerceiver(config.AuxConfig.PodPerceiverServiceAccountName, config.PodPerceiverReplicationCount)
	podPerceiver.Config = config.PodPerceiverConfig()
	podPerceiver.Config.PerceptorHost = perceptor.ServiceName

	perceptorScanner := model.NewScanner(config.ScannerMemory)
	perceptorScanner.Config = config.ScannerConfig()
	perceptorScanner.Config.PerceptorHost = perceptor.ServiceName
	perceptorScanner.HubPasswordSecretKey = config.HubPasswordSecretKey
	perceptorScanner.HubPasswordSecretName = config.HubPasswordSecretName

	imageFacade := model.NewImagefacade(config.AuxConfig.ImageFacadeServiceAccountName)
	imageFacade.Config = config.ImagefacadeConfig()

	skyfire := model.NewSkyfire()
	skyfire.Config = config.SkyfireConfig()
	skyfire.Config.PerceptorHost = perceptor.ServiceName

	prometheus := model.NewPrometheus()
	prometheus.AddTarget(&model.PrometheusTarget{Host: perceptor.ServiceName, Port: config.PerceptorPort})
	prometheus.AddTarget(&model.PrometheusTarget{Host: perceptorScanner.ServiceName, Port: config.ScannerPort})
	prometheus.AddTarget(&model.PrometheusTarget{Host: imageFacade.ServiceName, Port: config.ImageFacadePort})
	prometheus.AddTarget(&model.PrometheusTarget{Host: podPerceiver.ServiceName, Port: config.PodPerceiverPort})
	prometheus.AddTarget(&model.PrometheusTarget{Host: skyfire.ServiceName, Port: config.SkyfirePort})
	//	prometheus.Config = config.PrometheusConfig() // TODO ?

	scanner := model.NewScannerPod(perceptorScanner, imageFacade)
	scanner.ReplicaCount = config.ScannerReplicationCount

	kube.ReplicationControllers = []*v1.ReplicationController{
		perceptor.ReplicationController(),
		podPerceiver.ReplicationController(),
		scanner.ReplicationController(),
		skyfire.ReplicationController(),
	}
	kube.Services = []*v1.Service{
		perceptor.Service(),
		podPerceiver.Service(),
		perceptorScanner.Service(),
		imageFacade.Service(),
		skyfire.Service(),
		//		prometheus.Service(),
	}
	kube.ConfigMaps = []*v1.ConfigMap{
		perceptor.ConfigMap(),
		podPerceiver.ConfigMap(),
		perceptorScanner.ConfigMap(),
		imageFacade.ConfigMap(),
		prometheus.ConfigMap(),
		skyfire.ConfigMap(),
	}
	kube.Secrets = []*v1.Secret{
		{
			ObjectMeta: v1meta.ObjectMeta{
				Name: config.HubPasswordSecretName,
			},
			Type: v1.SecretTypeOpaque,
			StringData: map[string]string{
				config.HubPasswordSecretKey: config.HubUserPassword,
			},
		},
	}

	kube.Perceptor = perceptor
	kube.Scanner = perceptorScanner
	kube.ScannerPod = scanner
	kube.ImageFacade = imageFacade
	kube.PodPerceiver = podPerceiver
	kube.Prometheus = prometheus
	kube.Skyfire = skyfire
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
