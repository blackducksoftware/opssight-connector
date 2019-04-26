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
	"fmt"
	"os"

	"github.com/blackducksoftware/perceptor-scanner/pkg/common"
	"github.com/blackducksoftware/perceptor/pkg/api"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
)

// Scanner stores the scanner configurations
type Scanner struct {
	ifClient       ImageFacadeClientInterface
	scanClient     ScanClientInterface
	imageDirectory string
	stop           <-chan struct{}
}

// NewScanner return the Scanner configurations
func NewScanner(ifClient ImageFacadeClientInterface, scanClient ScanClientInterface, imageDirectory string, stop <-chan struct{}) *Scanner {
	return &Scanner{
		ifClient:       ifClient,
		scanClient:     scanClient,
		imageDirectory: imageDirectory,
		stop:           stop}
}

// ScanFullDockerImage runs the scan client on a full tar from 'docker export'
func (scanner *Scanner) ScanFullDockerImage(apiImage *api.ImageSpec) error {
	pullSpec := fmt.Sprintf("%s@sha256:%s", apiImage.Repository, apiImage.Sha)
	image := common.NewImage(scanner.imageDirectory, pullSpec)
	err := scanner.ifClient.PullImage(image)
	if err != nil {
		cleanUpFile(image.DockerTarFilePath())
		return errors.Trace(err)
	}
	defer cleanUpFile(image.DockerTarFilePath())
	return scanner.ScanFile(apiImage.Scheme, apiImage.Domain, apiImage.Port, apiImage.User, apiImage.Password, image.DockerTarFilePath(), apiImage.BlackDuckProjectName, apiImage.BlackDuckProjectVersionName, apiImage.BlackDuckScanName)
}

// ScanFile runs the scan client against a single file
func (scanner *Scanner) ScanFile(scheme string, host string, port int, username string, password string, path string, blackDuckProjectName string, blackDuckVersionName string, blackDuckScanName string) error {
	return scanner.scanClient.Scan(scheme, host, port, username, password, path, blackDuckProjectName, blackDuckVersionName, blackDuckScanName)
}

// cleanUpFile cleans up the file that is locally pulled for scanning
func cleanUpFile(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Debugf("unable to find the file path %s due to %s", path, err.Error())
		return
	}
	err := os.Remove(path)
	recordCleanUpFile(err == nil)
	if err != nil {
		log.Errorf("unable to remove file %s: %s", path, err.Error())
	} else {
		log.Infof("successfully cleaned up file %s", path)
	}
}
