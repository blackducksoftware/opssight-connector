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

type ScannerTester struct {
	PodName string

	ReplicaCount int32

	ImagesMountName string
	ImagesMountPath string

	Scanner     *Scanner
	Imagefacade *MockImagefacade
}

func NewScannerTester(scanner *Scanner, imagefacade *MockImagefacade) *ScannerTester {
	scannerTester := &ScannerTester{
		PodName: "scanner-tester",

		ReplicaCount: 1,

		ImagesMountName: "var-images",
		ImagesMountPath: "/var/images",

		Scanner:     scanner,
		Imagefacade: imagefacade,
	}

	scanner.ImagesMountName = scannerTester.ImagesMountName
	scanner.ImagesMountPath = scannerTester.ImagesMountPath

	scanner.PodName = scannerTester.PodName

	imagefacade.ImagesMountName = scannerTester.ImagesMountName
	imagefacade.ImagesMountPath = scannerTester.ImagesMountPath

	imagefacade.PodName = scannerTester.PodName

	return scannerTester
}

func (sc *ScannerTester) ReplicationController() *v1.ReplicationController {
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
					},
					Containers: []v1.Container{*sc.Scanner.Container(), *sc.Imagefacade.Container()},
				}}}}
}
