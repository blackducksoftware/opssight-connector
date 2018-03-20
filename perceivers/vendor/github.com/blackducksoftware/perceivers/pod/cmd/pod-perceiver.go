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

package main

import (
	"fmt"

	"github.com/blackducksoftware/perceivers/pkg/annotations"
	"github.com/blackducksoftware/perceivers/pod/cmd/app"

	log "github.com/sirupsen/logrus"
)

// TODO metrics
// number of namespaces found
// number of pods per namespace
// number of images per pod
// number of occurrences of each pod
// number of successes, failures, of each perceptor endpoint
// ??? number of scan results fetched from perceptor

func main() {
	log.Info("starting pod-perceiver")

	handler := annotations.PodAnnotatorHandlerFuncs{
		PodLabelCreationFunc:      annotations.CreatePodLabels,
		PodAnnotationCreationFunc: annotations.CreatePodAnnotations,
		ImageAnnotatorHandlerFuncs: annotations.ImageAnnotatorHandlerFuncs{
			ImageLabelCreationFunc:      annotations.CreateImageLabels,
			ImageAnnotationCreationFunc: annotations.CreateImageAnnotations,
			MapCompareHandlerFuncs: annotations.MapCompareHandlerFuncs{
				MapCompareFunc: annotations.StringMapContains,
			},
		},
	}

	// Create the Pod Perceiver
	perceiver, err := app.NewPodPerceiver(handler)
	if err != nil {
		panic(fmt.Errorf("failed to create pod-perceiver: %v", err))
	}

	// Run the perceiver
	stopCh := make(chan struct{})
	perceiver.Run(stopCh)
}
