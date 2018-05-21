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
	"os"
	"time"

	"github.com/blackducksoftware/perceptor/pkg/api"
	log "github.com/sirupsen/logrus"
)

const (
	requestScanJobPause = 20 * time.Second
	imageFacadeBaseURL  = "http://localhost"
)

type Scanner struct {
	scanClient    ScanClientInterface
	httpClient    *http.Client
	perceptorHost string
	perceptorPort int
}

func NewScanner(config *Config) (*Scanner, error) {
	log.Infof("instantiating Scanner with config %+v", config)

	hubPassword, ok := os.LookupEnv(config.HubUserPasswordEnvVar)
	if !ok {
		return nil, fmt.Errorf("unable to get Hub password: environment variable %s not set", config.HubUserPasswordEnvVar)
	}

	err := os.Setenv("BD_HUB_PASSWORD", hubPassword)
	if err != nil {
		log.Errorf("unable to set BD_HUB_PASSWORD environment variable: %s", err.Error())
		return nil, err
	}

	scanClientInfo, err := downloadScanClient(
		config.HubHost,
		config.HubUser,
		hubPassword,
		config.HubPort,
		time.Duration(config.HubClientTimeoutSeconds)*time.Second)
	if err != nil {
		log.Errorf("unable to download scan client: %s", err.Error())
		return nil, err
	}

	log.Infof("instantiating scanner with hub %s, user %s", config.HubHost, config.HubUser)

	imagePuller := NewImageFacadePuller(imageFacadeBaseURL, config.ImageFacadePort)
	scanClient, err := NewHubScanClient(
		config.HubHost,
		config.HubUser,
		config.HubPort,
		scanClientInfo,
		imagePuller)
	if err != nil {
		log.Errorf("unable to instantiate hub scan client: %s", err.Error())
		return nil, err
	}

	httpClient := &http.Client{Timeout: 5 * time.Second}

	scanner := Scanner{
		scanClient:    scanClient,
		httpClient:    httpClient,
		perceptorHost: config.PerceptorHost,
		perceptorPort: config.PerceptorPort}

	return &scanner, nil
}

// StartRequestingScanJobs will start asking for work
func (scanner *Scanner) StartRequestingScanJobs() {
	log.Infof("starting to request scan jobs")
	go func() {
		for {
			time.Sleep(requestScanJobPause)
			scanner.requestAndRunScanJob()
		}
	}()
}

func (scanner *Scanner) requestAndRunScanJob() {
	log.Debug("requesting scan job")
	image, err := scanner.requestScanJob()
	if err != nil {
		log.Errorf("unable to request scan job: %s", err.Error())
		return
	}
	if image == nil {
		log.Debug("requested scan job, got nil")
		return
	}

	log.Infof("processing scan job %+v", image)

	job := NewScanJob(image.PullSpec, image.Sha, image.HubProjectName, image.HubProjectVersionName, image.HubScanName)
	err = scanner.scanClient.Scan(*job)
	errorString := ""
	if err != nil {
		errorString = err.Error()
	}

	finishedJob := api.FinishedScanClientJob{Err: errorString, ImageSpec: *image}
	log.Infof("about to finish job, going to send over %+v", finishedJob)
	err = scanner.finishScan(finishedJob)
	if err != nil {
		log.Errorf("unable to finish scan job: %s", err.Error())
	}
}

func (scanner *Scanner) requestScanJob() (*api.ImageSpec, error) {
	nextImageURL := scanner.buildURL(api.NextImagePath)
	resp, err := scanner.httpClient.Post(nextImageURL, "", bytes.NewBuffer([]byte{}))

	if err != nil {
		recordScannerError("unable to POST get next image")
		log.Errorf("unable to POST to %s: %s", nextImageURL, err.Error())
		return nil, err
	}

	recordHTTPStats(api.NextImagePath, resp.StatusCode)

	if resp.StatusCode != 200 {
		err = fmt.Errorf("http POST request to %s failed with status code %d", nextImageURL, resp.StatusCode)
		log.Error(err.Error())
		return nil, err
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		recordScannerError("unable to read response body")
		log.Errorf("unable to read response body from %s: %s", nextImageURL, err.Error())
		return nil, err
	}

	var nextImage api.NextImage
	err = json.Unmarshal(bodyBytes, &nextImage)
	if err != nil {
		recordScannerError("unmarshaling JSON body failed")
		log.Errorf("unmarshaling JSON body bytes %s failed for URL %s: %s", string(bodyBytes), nextImageURL, err.Error())
		return nil, err
	}

	imageSha := "null"
	if nextImage.ImageSpec != nil {
		imageSha = nextImage.ImageSpec.Sha
	}
	log.Debugf("http POST request to %s succeeded, got image %s", nextImageURL, imageSha)
	return nextImage.ImageSpec, nil
}

func (scanner *Scanner) finishScan(results api.FinishedScanClientJob) error {
	finishedScanURL := scanner.buildURL(api.FinishedScanPath)
	jsonBytes, err := json.Marshal(results)
	if err != nil {
		recordScannerError("unable to marshal json for finished job")
		log.Errorf("unable to marshal json for finished job: %s", err.Error())
		return err
	}

	log.Debugf("about to send over json text for finishing a job: %s", string(jsonBytes))
	// TODO change to exponential backoff or something ... but don't loop indefinitely in production
	for {
		resp, err := scanner.httpClient.Post(finishedScanURL, "application/json", bytes.NewBuffer(jsonBytes))
		if err != nil {
			recordScannerError("unable to POST finished job")
			log.Errorf("unable to POST to %s: %s", finishedScanURL, err.Error())
			continue
		}

		recordHTTPStats(api.FinishedScanPath, resp.StatusCode)

		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Errorf("POST to %s failed with status code %d", finishedScanURL, resp.StatusCode)
			continue
		}

		log.Infof("POST to %s succeeded", finishedScanURL)
		return nil
	}
}

func (scanner *Scanner) buildURL(path string) string {
	return fmt.Sprintf("http://%s:%d/%s", scanner.perceptorHost, scanner.perceptorPort, path)
}
