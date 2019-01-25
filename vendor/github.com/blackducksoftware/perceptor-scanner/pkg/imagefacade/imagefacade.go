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
	"time"

	"github.com/blackducksoftware/perceptor-scanner/pkg/common"
	pdocker "github.com/blackducksoftware/perceptor-scanner/pkg/docker"
	imagepullerinterface "github.com/blackducksoftware/perceptor-scanner/pkg/interfaces"
	"github.com/blackducksoftware/perceptor-scanner/pkg/skopeo"
	log "github.com/sirupsen/logrus"
)

const (
	diskMetricsPause = 15 * time.Second
)

// ImageFacade ...
type ImageFacade struct {
	model            *Model
	imagePuller      imagepullerinterface.ImagePuller
	createImagesOnly bool
}

// NewImageFacade ...
func NewImageFacade(dockerRegistries []common.RegistryAuth, createImagesOnly bool, imagePullerType string, stop <-chan struct{}) *ImageFacade {
	model := NewModel(stop)
	var imagePuller imagepullerinterface.ImagePuller

	switch imagePullerType {
	case "skopeo":
		imagePuller = skopeo.NewImagePuller(dockerRegistries)
	default:
		imagePuller = pdocker.NewImagePuller(dockerRegistries)
	}

	imageFacade := &ImageFacade{
		model:            model,
		imagePuller:      imagePuller,
		createImagesOnly: createImagesOnly}

	SetupHTTPServer(imageFacade)

	go func() {
		for {
			select {
			case <-stop:
				return
			case <-time.After(diskMetricsPause):
				imageFacade.pullDiskMetrics()
			}
		}
	}()

	return imageFacade
}

func (imf *ImageFacade) pullImage(image *common.Image) error {
	var err error
	if imf.createImagesOnly {
		err = imf.imagePuller.CreateImageInLocalDocker(image)
	} else {
		err = imf.imagePuller.PullImage(image)
	}
	recordImagePullResult(err == nil)
	return err
}

func (imf *ImageFacade) pullDiskMetrics() {
	log.Debugf("getting disk metrics")
	diskMetrics, err := getDiskMetrics()
	if err == nil {
		log.Debugf("got disk metrics: %+v", diskMetrics)
		recordDiskMetrics(diskMetrics)
	} else {
		log.Errorf("unable to get disk metrics: %s", err.Error())
	}
}

// HTTPResponder implementation

// PullImage ...
func (imf *ImageFacade) PullImage(image *common.Image) error {
	err := imf.model.StartImagePull(image)
	if err != nil {
		return err
	}
	go func() {
		pullErr := imf.pullImage(image)
		if pullErr != nil {
			log.Errorf("unable to pull image: %s", pullErr.Error())
		}
		finishErr := imf.model.finishImagePull(image, pullErr)
		if finishErr != nil {
			log.Errorf("unable to finish image pull: %s", finishErr.Error())
		}
	}()
	return nil
}

// GetImage ...
func (imf *ImageFacade) GetImage(image *common.Image) common.ImageStatus {
	return imf.model.GetImageStatus(image)
}

// GetModel ...
func (imf *ImageFacade) GetModel() map[string]interface{} {
	return imf.model.GetAPIModel()
}
