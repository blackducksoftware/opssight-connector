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
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
)

const (
	pullImagePath  = "pullimage"
	checkImagePath = "checkimage"
)

// ImageFacadeClientInterface ...
type ImageFacadeClientInterface interface {
	PullImage(image *common.Image) error
}

// ImageFacadeClient ...
type ImageFacadeClient struct {
	ImageFacadeHost string
	ImageFacadePort int
	httpClient      *http.Client
}

// NewImageFacadeClient ...
func NewImageFacadeClient(imageFacadeHost string, imageFacadePort int) *ImageFacadeClient {
	return &ImageFacadeClient{
		ImageFacadeHost: imageFacadeHost,
		ImageFacadePort: imageFacadePort,
		httpClient:      &http.Client{Timeout: 5 * time.Second}}
}

// PullImage ...
func (ifp *ImageFacadeClient) PullImage(image *common.Image) error {
	log.Infof("attempting to pull image %s", image.PullSpec)

	err := ifp.startImagePull(image)
	if err != nil {
		return errors.Annotatef(err, "unable to pull image %s", image.PullSpec)
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

func (ifp *ImageFacadeClient) startImagePull(image *common.Image) error {
	url := ifp.buildURL(pullImagePath)

	requestBytes, err := json.Marshal(image)
	if err != nil {
		return errors.Annotatef(err, "unable to marshal JSON for %s", image.PullSpec)
	}

	resp, err := ifp.httpClient.Post(url, "application/json", bytes.NewBuffer(requestBytes))
	if err != nil {
		return errors.Annotatef(err, "unable to create request to %s for image %s", url, image.PullSpec)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("request to start image pull for image %s failed with status code %d", url, resp.StatusCode)
	}

	defer resp.Body.Close()
	_, _ = ioutil.ReadAll(resp.Body)

	log.Infof("request to start image pull for image %s succeeded", image.PullSpec)

	return nil
}

func (ifp *ImageFacadeClient) checkImage(image *common.Image) (common.ImageStatus, error) {
	url := ifp.buildURL(checkImagePath)

	requestBytes, err := json.Marshal(image)
	if err != nil {
		return common.ImageStatusUnknown, errors.Annotatef(err, "unable to marshal JSON for %s", image.PullSpec)
	}

	resp, err := ifp.httpClient.Post(url, "application/json", bytes.NewBuffer(requestBytes))
	if err != nil {
		return common.ImageStatusUnknown, errors.Annotatef(err, "unable to create request to %s for image %s", url, image.PullSpec)
	}

	if resp.StatusCode != 200 {
		return common.ImageStatusUnknown, fmt.Errorf("GET %s failed with status code %d", url, resp.StatusCode)
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		recordScannerError("unable to read response body")
		return common.ImageStatusUnknown, errors.Annotatef(err, "unable to read response body from %s", url)
	}

	var getImage api.CheckImageResponse
	err = json.Unmarshal(bodyBytes, &getImage)
	if err != nil {
		recordScannerError("unmarshaling JSON body failed")
		return common.ImageStatusUnknown, errors.Annotatef(err, "unmarshaling JSON body bytes %s failed for URL %s", string(bodyBytes), url)
	}

	log.Debugf("image check for image %s succeeded, status %s", image.PullSpec, getImage.ImageStatus.String())

	return getImage.ImageStatus, nil
}

func (ifp *ImageFacadeClient) buildURL(path string) string {
	return fmt.Sprintf("http://%s:%d/%s?", ifp.ImageFacadeHost, ifp.ImageFacadePort, path)
}
