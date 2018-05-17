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
	log "github.com/sirupsen/logrus"
)

const (
	diskMetricsPause = 15 * time.Second
)

type ImageFacade struct {
	server            *HTTPServer
	reducer           *reducer
	finishedImagePull chan *finishedImagePull
	imagePuller       *pdocker.ImagePuller
	createImagesOnly  bool
}

func NewImageFacade(dockerUser string, dockerPassword string, internalDockerRegistries []string, createImagesOnly bool) *ImageFacade {
	actions := make(chan Action)
	finishedImagePull := make(chan *finishedImagePull)

	server := NewHTTPServer()
	model := NewModel()
	reducer := newReducer(model, actions)

	go func() {
		for {
			select {
			case pullImage := <-server.PullImageChannel():
				actions <- pullImage
			case getImage := <-server.GetImageChannel():
				actions <- getImage
			case finished := <-finishedImagePull:
				actions <- finished
			}
		}
	}()

	imageFacade := &ImageFacade{
		server:            server,
		reducer:           reducer,
		finishedImagePull: finishedImagePull,
		imagePuller:       pdocker.NewImagePuller(dockerUser, dockerPassword, internalDockerRegistries),
		createImagesOnly:  createImagesOnly}

	go func() {
		for {
			select {
			case nextImage := <-model.PullImageChannel():
				imageFacade.pullImage(nextImage)
			}
		}
	}()

	go imageFacade.pullDiskMetrics()

	return imageFacade
}

func (imf *ImageFacade) pullImage(image *common.Image) {
	var err error
	if imf.createImagesOnly {
		err = imf.imagePuller.CreateImageInLocalDocker(image)
	} else {
		err = imf.imagePuller.PullImage(image)
	}
	recordImagePullResult(err == nil)
	imf.finishedImagePull <- &finishedImagePull{image: image, err: err}
}

func (imf *ImageFacade) pullDiskMetrics() {
	for {
		log.Debugf("getting disk metrics")
		diskMetrics, err := getDiskMetrics()
		if err == nil {
			log.Debugf("got disk metrics: %+v", diskMetrics)
			recordDiskMetrics(diskMetrics)
		} else {
			log.Errorf("unable to get disk metrics: %s", err.Error())
		}
		time.Sleep(diskMetricsPause)
	}
}
