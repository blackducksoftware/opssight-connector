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

package perceptor

import (
	"fmt"

	"github.com/blackducksoftware/perceptor/pkg/api"
)

// Dump .....
type Dump struct {
	ScanResults        *api.ScanResults
	Model              *api.Model
	PodsByName         map[string]*api.ScannedPod
	DuplicatePodNames  map[string]bool
	ImagesBySha        map[string]*api.ScannedImage
	DuplicateImageShas map[string]bool
}

// NewDump .....
func NewDump(scanResults *api.ScanResults, model *api.Model) *Dump {
	dump := &Dump{
		ScanResults:        scanResults,
		Model:              model,
		PodsByName:         map[string]*api.ScannedPod{},
		DuplicatePodNames:  map[string]bool{},
		ImagesBySha:        map[string]*api.ScannedImage{},
		DuplicateImageShas: map[string]bool{}}
	dump.computeDerivedData()
	return dump
}

func (pd *Dump) computeDerivedData() {
	// TODO figure out what, if anything, to do with Model
	for _, pod := range pd.ScanResults.Pods {
		podName := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
		_, ok := pd.PodsByName[podName]
		if ok {
			pd.DuplicatePodNames[podName] = true
		} else {
			// wow, what a hack.  avoids accidentally assigning the same pod every time ...
			// since `pod` is actually the same variable for every iteration through the loop
			// the actual solution would be to have perceptor pass by reference or something
			podRef := pod
			pd.PodsByName[podName] = &podRef
		}
	}
	for _, image := range pd.ScanResults.Images {
		_, ok := pd.ImagesBySha[image.Sha]
		if ok {
			pd.DuplicateImageShas[image.Sha] = true
		} else {
			// again: wow, what a hack
			imageRef := image
			pd.ImagesBySha[image.Sha] = &imageRef
		}
	}
}
