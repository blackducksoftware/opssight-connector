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
	"k8s.io/api/core/v1"

	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ScannerPod struct {
	PodName string

	ReplicaCount int32

	DockerSocketName string
	DockerSocketPath string

	ImagesMountName string
	ImagesMountPath string

	Scanner     *Scanner
	Imagefacade *Imagefacade
}

func NewScannerPod(scanner *Scanner, imagefacade *Imagefacade) *ScannerPod {
	scannerPod := &ScannerPod{
		PodName: "perceptor-scanner",

		ReplicaCount: 0,

		DockerSocketName: "dir-docker-socket",
		DockerSocketPath: "/var/run/docker.sock",

		ImagesMountName: "var-images",
		ImagesMountPath: "/var/images",

		Scanner:     scanner,
		Imagefacade: imagefacade,
	}

	scanner.ImagesMountName = scannerPod.ImagesMountName
	scanner.ImagesMountPath = scannerPod.ImagesMountPath

	scanner.PodName = scannerPod.PodName

	imagefacade.ImagesMountName = scannerPod.ImagesMountName
	imagefacade.ImagesMountPath = scannerPod.ImagesMountPath

	imagefacade.DockerSocketName = scannerPod.DockerSocketName
	imagefacade.DockerSocketPath = scannerPod.DockerSocketPath

	imagefacade.PodName = scannerPod.PodName

	return scannerPod
}

func (sc *ScannerPod) ReplicationController() *v1.ReplicationController {
	return &v1.ReplicationController{
		ObjectMeta: v1meta.ObjectMeta{Name: sc.PodName},
		Spec: v1.ReplicationControllerSpec{
			Replicas: &sc.ReplicaCount,
			Selector: map[string]string{"name": sc.PodName},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: v1meta.ObjectMeta{Labels: map[string]string{"name": sc.PodName}},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						{
							Name: sc.Scanner.ConfigMapName,
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{Name: sc.Scanner.ConfigMapName},
								},
							},
						},
						{
							Name: sc.Imagefacade.ConfigMapName,
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{Name: sc.Imagefacade.ConfigMapName},
								},
							},
						},
						{
							Name:         sc.ImagesMountName,
							VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{}},
						},
						{
							Name: sc.DockerSocketName,
							VolumeSource: v1.VolumeSource{
								HostPath: &v1.HostPathVolumeSource{Path: sc.DockerSocketPath},
							},
						},
					},
					Containers:         []v1.Container{*sc.Scanner.Container(), *sc.Imagefacade.Container()},
					ServiceAccountName: sc.Imagefacade.ServiceAccountName,
					// TODO: RestartPolicy?  terminationGracePeriodSeconds? dnsPolicy?
				}}}}
}
