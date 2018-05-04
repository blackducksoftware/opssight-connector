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

	oca "github.com/blackducksoftware/opssight-connector/processors/pkg/annotations"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("starting pod-processor")

	handler := annotations.PodAnnotatorHandlerFuncs{
		PodLabelCreationFunc:      oca.CreatePodLabels,
		PodAnnotationCreationFunc: oca.CreatePodAnnotations,
		ImageAnnotatorHandlerFuncs: annotations.ImageAnnotatorHandlerFuncs{
			ImageLabelCreationFunc:      oca.CreateImageLabels,
			ImageAnnotationCreationFunc: oca.CreateImageAnnotations,
			MapCompareHandlerFuncs: annotations.MapCompareHandlerFuncs{
				MapCompareFunc: oca.MapContainsBlackDuckEntries,
			},
		},
	}

	// Create the Pod Perceiver
	processor, err := app.NewPodPerceiver(handler)
	if err != nil {
		panic(fmt.Errorf("failed to create pod-processor: %v", err))
	}

	// Run the processor
	stopCh := make(chan struct{})
	processor.Run(stopCh)
}
