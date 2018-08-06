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
	"fmt"

	common "github.com/blackducksoftware/perceptor-scanner/pkg/common"
	log "github.com/sirupsen/logrus"
)

type Model struct {
	State            ModelState
	Images           map[string]common.ImageStatus
	pullImageChannel chan *common.Image
}

func NewModel() *Model {
	return &Model{
		State:            ModelStateReady,
		Images:           map[string]common.ImageStatus{},
		pullImageChannel: make(chan *common.Image),
	}
}

func (model *Model) PullImageChannel() <-chan *common.Image {
	return model.pullImageChannel
}

func (model *Model) pullImage(image *common.Image) error {
	if model.State != ModelStateReady {
		err := fmt.Errorf("unable to pull image %s, image pull in progress (%s)", image.PullSpec, model.State.String())
		log.Errorf(err.Error())
		return err
	}

	log.Infof("about to start pulling image %s -- model state %s", image.PullSpec, model.State.String())
	model.Images[image.PullSpec] = common.ImageStatusInProgress
	model.State = ModelStatePulling
	model.pullImageChannel <- image
	return nil
}

func (model *Model) finishImagePull(image *common.Image, err error) {
	if err == nil {
		log.Infof("successfully finished image pull for %s", image.PullSpec)
		model.Images[image.PullSpec] = common.ImageStatusDone
	} else {
		log.Errorf("finished image pull for %s with error %s", image.PullSpec, err.Error())
		model.Images[image.PullSpec] = common.ImageStatusError
	}
	model.State = ModelStateReady
}

func (model *Model) imageStatus(image *common.Image) common.ImageStatus {
	imageStatus, ok := model.Images[image.PullSpec]
	if !ok {
		return common.ImageStatusUnknown
	}
	return imageStatus
}
