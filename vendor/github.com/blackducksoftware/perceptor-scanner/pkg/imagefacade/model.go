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
	"time"

	common "github.com/blackducksoftware/perceptor-scanner/pkg/common"
	log "github.com/sirupsen/logrus"
)

type action struct {
	name  string
	apply func() error
}

// Model ...
type Model struct {
	actions chan *action
	State   ModelState
	Images  map[string]common.ImageStatus
}

// NewModel ...
func NewModel(stop <-chan struct{}) *Model {
	model := &Model{
		actions: make(chan *action),
		State:   ModelStateReady,
		Images:  map[string]common.ImageStatus{},
	}

	go func() {
		stopTime := time.Now()
		for {
			select {
			case <-stop:
				return
			case action := <-model.actions:
				// metrics: log message type
				log.Debugf("processing action of type %s", action.name)
				recordActionType(action.name)

				// metrics: how long idling since the last action finished processing?
				startTime := time.Now()
				recordReducerActivity(false, startTime.Sub(stopTime))

				err := action.apply()

				// metrics: how long did the work take?
				stopTime = time.Now()
				recordReducerActivity(true, stopTime.Sub(startTime))

				if err != nil {
					log.Errorf("error processing action of type %s: %s", action.name, err.Error())
				} else {
					log.Debugf("successfully processed action of type %s", action.name)
				}
			}
		}
	}()

	return model
}

// public interface

// StartImagePull ...
func (model *Model) StartImagePull(image *common.Image) error {
	ch := make(chan error)
	model.actions <- &action{"startImagePull", func() error {
		err := model.pullImage(image)
		ch <- err
		return err
	}}
	return <-ch
}

// GetImageStatus ...
func (model *Model) GetImageStatus(image *common.Image) common.ImageStatus {
	ch := make(chan common.ImageStatus)
	model.actions <- &action{"getImageStatus", func() error {
		status, err := model.imageStatus(image)
		ch <- status
		return err
	}}
	return <-ch
}

// FinishImagePull ...
func (model *Model) FinishImagePull(image *common.Image, imagePullError error) error {
	ch := make(chan error)
	model.actions <- &action{"finishImagePull", func() error {
		err := model.finishImagePull(image, imagePullError)
		ch <- err
		return err
	}}
	return <-ch
}

// GetAPIModel ...
func (model *Model) GetAPIModel() map[string]interface{} {
	ch := make(chan map[string]interface{})
	model.actions <- &action{"getAPIModel", func() error {
		ch <- model.getAPIModel()
		return nil
	}}
	return <-ch
}

// private interface

func (model *Model) pullImage(image *common.Image) error {
	if model.State != ModelStateReady {
		return fmt.Errorf("unable to pull image %s, image pull in progress (%s)", image.PullSpec, model.State.String())
	}

	log.Infof("about to start pulling image %s -- model state %s", image.PullSpec, model.State.String())
	model.Images[image.PullSpec] = common.ImageStatusInProgress
	model.State = ModelStatePulling
	return nil
}

func (model *Model) finishImagePull(image *common.Image, imagePullError error) error {
	if _, ok := model.Images[image.PullSpec]; !ok {
		return fmt.Errorf("finishImagePull %s with error %t: image not found", image.PullSpec, imagePullError == nil)
	}
	if imagePullError == nil {
		log.Infof("successfully finished image pull for %s", image.PullSpec)
		model.Images[image.PullSpec] = common.ImageStatusDone
	} else {
		log.Errorf("finished image pull for %s with error %s", image.PullSpec, imagePullError.Error())
		model.Images[image.PullSpec] = common.ImageStatusError
	}
	model.State = ModelStateReady
	return nil
}

func (model *Model) imageStatus(image *common.Image) (common.ImageStatus, error) {
	imageStatus, ok := model.Images[image.PullSpec]
	if !ok {
		return common.ImageStatusUnknown, fmt.Errorf("image %s not found", image.PullSpec)
	}
	return imageStatus, nil
}

func (model *Model) getAPIModel() map[string]interface{} {
	images := map[string]string{}
	for key, val := range model.Images {
		images[key] = val.String()
	}
	return map[string]interface{}{
		"State":  model.State.String(),
		"Images": images,
	}
}
