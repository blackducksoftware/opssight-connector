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

package scanner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/blackducksoftware/perceptor-scanner/pkg/api"
	"github.com/blackducksoftware/perceptor-scanner/pkg/common"
	"github.com/prometheus/common/log"
)

const (
	pullImagePath  = "pullimage"
	checkImagePath = "checkimage"
)

type ImageFacadePuller struct {
	ImageFacadeHost string
	ImageFacadePort int
	httpClient      *http.Client
}

func NewImageFacadePuller(imageFacadeHost string, imageFacadePort int) *ImageFacadePuller {
	return &ImageFacadePuller{
		ImageFacadeHost: imageFacadeHost,
		ImageFacadePort: imageFacadePort,
		httpClient:      &http.Client{Timeout: 5 * time.Second}}
}

func (ifp *ImageFacadePuller) PullImage(image *common.Image) error {
	log.Infof("attempting to pull image %s", image.PullSpec)

	err := ifp.startImagePull(image)
	if err != nil {
		log.Errorf("unable to pull image %s: %s", image.PullSpec, err.Error())
		return err
	}

	for {
		time.Sleep(5 * time.Second)

		imageStatus, err := ifp.checkImage(image)
		if err != nil {
			log.Errorf("unable to check image %s: %s", image.PullSpec, err.Error())
		}

		switch imageStatus {
		case common.ImageStatusUnknown:
			// job got lost somehow -- maybe the container crashed
			return fmt.Errorf("unable to pull image %s: job was lost", image.PullSpec)
		case common.ImageStatusInProgress:
			// just keep on waiting
			break
		case common.ImageStatusDone:
			log.Infof("finished pulling image %s", image.PullSpec)
			return nil
		case common.ImageStatusError:
			return fmt.Errorf("unable to pull image %s", image.PullSpec)
		default:
			panic(fmt.Errorf("invalid ImageStatus value %d", imageStatus))
		}
	}
}

func (ifp *ImageFacadePuller) startImagePull(image *common.Image) error {
	url := ifp.buildURL(pullImagePath)

	requestBytes, err := json.Marshal(image)
	if err != nil {
		log.Errorf("unable to marshal JSON for %s: %s", image.PullSpec, err.Error())
		return err
	}

	resp, err := ifp.httpClient.Post(url, "application/json", bytes.NewBuffer(requestBytes))
	if err != nil {
		log.Errorf("unable to create request to %s for image %s: %s", url, image.PullSpec, err.Error())
		return err
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("request to start image pull for image %s failed with status code %d", url, resp.StatusCode)
		log.Errorf(err.Error())
		return err
	}

	defer resp.Body.Close()
	_, _ = ioutil.ReadAll(resp.Body)

	log.Infof("request to start image pull for image %s succeeded", image.PullSpec)

	return nil
}

func (ifp *ImageFacadePuller) checkImage(image *common.Image) (common.ImageStatus, error) {
	url := ifp.buildURL(checkImagePath)

	requestBytes, err := json.Marshal(image)
	if err != nil {
		log.Errorf("unable to marshal JSON for %s: %s", image.PullSpec, err.Error())
		return common.ImageStatusUnknown, err
	}

	resp, err := ifp.httpClient.Post(url, "application/json", bytes.NewBuffer(requestBytes))
	if err != nil {
		log.Errorf("unable to create request to %s for image %s: %s", url, image.PullSpec, err.Error())
		return common.ImageStatusUnknown, err
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("GET %s failed with status code %d", url, resp.StatusCode)
		log.Errorf(err.Error())
		return common.ImageStatusUnknown, err
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		recordScannerError("unable to read response body")
		log.Errorf("unable to read response body from %s: %s", url, err.Error())
		return common.ImageStatusUnknown, err
	}

	var getImage api.CheckImageResponse
	err = json.Unmarshal(bodyBytes, &getImage)
	if err != nil {
		recordScannerError("unmarshaling JSON body failed")
		log.Errorf("unmarshaling JSON body bytes %s failed for URL %s: %s", string(bodyBytes), url, err.Error())
		return common.ImageStatusUnknown, err
	}

	log.Debugf("image check for image %s succeeded", image.PullSpec)

	return getImage.ImageStatus, nil
}

func (ifp *ImageFacadePuller) buildURL(path string) string {
	return fmt.Sprintf("http://%s:%d/%s?", ifp.ImageFacadeHost, ifp.ImageFacadePort, path)
}
