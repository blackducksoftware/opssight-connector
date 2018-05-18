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

package model

import (
	"encoding/json"
	"fmt"

	"k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type PrometheusTarget struct {
	Host string
	Port int32
}

type Prometheus struct {
	Name                string
	PodName             string
	Image               string
	ReplicaCount        int32
	DataVolumeName      string
	DataVolumeMountPath string
	ConfigMapName       string
	ConfigMapSource     string
	ConfigMapMountPath  string
	ServiceName         string
	Port                int32
	Targets             []*PrometheusTarget
}

func NewPrometheus() *Prometheus {
	return &Prometheus{
		Name:                "prometheus",
		PodName:             "prometheus-pod",
		Image:               "prom/prometheus:v2.1.0",
		ReplicaCount:        1,
		DataVolumeName:      "data",
		DataVolumeMountPath: "/data",
		ConfigMapName:       "config-volume",
		ConfigMapSource:     "prometheus",
		ConfigMapMountPath:  "/etc/prometheus",
		ServiceName:         "prometheus",
		Port:                9090,
	}
}

func (prom *Prometheus) AddTarget(target *PrometheusTarget) {
	prom.Targets = append(prom.Targets, target)
}

func (prom *Prometheus) Deployment() *v1beta1.Deployment {
	return &v1beta1.Deployment{
		ObjectMeta: v1meta.ObjectMeta{Name: prom.PodName},
		Spec: v1beta1.DeploymentSpec{
			Replicas: &prom.ReplicaCount,
			Selector: &v1meta.LabelSelector{
				MatchLabels: map[string]string{"app": prom.PodName},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: v1meta.ObjectMeta{
					Name:   prom.Name,
					Labels: map[string]string{"app": prom.PodName}},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						{
							Name:         prom.DataVolumeName,
							VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{}},
						},
						{
							Name: prom.ConfigMapName,
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{Name: prom.ConfigMapSource},
								},
							},
						},
					},
					Containers: []v1.Container{*prom.container()},
				}}}}
}

func (prom *Prometheus) container() *v1.Container {
	return &v1.Container{
		Name:  "prometheus",
		Image: prom.Image,
		Args: []string{
			"--log.level=debug",
			"--config.file=/etc/prometheus/prometheus.yml",
			"--storage.tsdb.path=/tmp/data/",
		},
		Ports: []v1.ContainerPort{
			{
				Name:          "web",
				ContainerPort: prom.Port,
			},
		},
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      prom.DataVolumeName,
				MountPath: prom.DataVolumeMountPath,
			},
			{
				Name:      prom.ConfigMapName,
				MountPath: prom.ConfigMapMountPath,
			},
		},
	}
}

func (prom *Prometheus) Service() *v1.Service {
	return &v1.Service{
		ObjectMeta: v1meta.ObjectMeta{
			Annotations: map[string]string{"prometheus.io/scrape": "true"},
			Labels:      map[string]string{"name": "prometheus"},
			Name:        prom.ServiceName},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeNodePort,
			Ports: []v1.ServicePort{
				{
					Name:       prom.ServiceName,
					Port:       prom.Port,
					Protocol:   "TCP",
					TargetPort: intstr.IntOrString{IntVal: prom.Port},
				},
			},
			Selector: map[string]string{"name": prom.PodName}}}
}

func (prom *Prometheus) ConfigMap() *v1.ConfigMap {
	targets := []string{}
	for _, target := range prom.Targets {
		targets = append(targets, fmt.Sprintf("%s:%d", target.Host, target.Port))
	}
	targetsBytes, err := json.Marshal(targets)
	if err != nil {
		panic(err)
	}
	jsonString := `
  {
    "global": {
      "scrape_interval": "5s"
    },
    "scrape_configs": [
      {
        "job_name": "perceptor-scrape",
        "scrape_interval": "5s",
        "static_configs": [
          {
            "targets": %s
          }
        ]
      }
    ]
  }
  `
	paramString := fmt.Sprintf(jsonString, string(targetsBytes))
	return MakeConfigMap("prometheus", "prometheus.yml", paramString)
}
