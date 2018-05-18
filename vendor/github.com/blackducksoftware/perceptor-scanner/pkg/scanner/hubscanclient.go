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
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	hubScheme = "https"
)

// HubScanClient implements ScanClientInterface using
// the Black Duck hub and scan client programs.
type HubScanClient struct {
	host           string
	username       string
	port           int
	scanClientInfo *scanClientInfo
	imagePuller    ImagePullerInterface
}

// NewHubScanClient requires hub login credentials
func NewHubScanClient(host string, username string, port int, scanClientInfo *scanClientInfo, imagePuller ImagePullerInterface) (*HubScanClient, error) {
	hsc := HubScanClient{
		host:           host,
		username:       username,
		port:           port,
		scanClientInfo: scanClientInfo,
		imagePuller:    imagePuller}
	return &hsc, nil
}

func (hsc *HubScanClient) Scan(job ScanJob) error {
	startTotal := time.Now()
	image := job.image()
	err := hsc.imagePuller.PullImage(image)

	defer cleanUpTarFile(image.DockerTarFilePath())

	if err != nil {
		recordScannerError("docker image pull and tar file creation")
		log.Errorf("unable to pull docker image %s: %s", job.PullSpec, err.Error())
		return err
	}

	scanCliImplJarPath := hsc.scanClientInfo.scanCliImplJarPath()
	scanCliJarPath := hsc.scanClientInfo.scanCliJarPath()
	scanCliJavaPath := hsc.scanClientInfo.scanCliJavaPath()
	path := image.DockerTarFilePath()
	cmd := exec.Command(scanCliJavaPath+"java",
		"-Xms512m",
		"-Xmx4096m",
		"-Dblackduck.scan.cli.benice=true",
		"-Dblackduck.scan.skipUpdate=true",
		"-Done-jar.silent=true",
		"-Done-jar.jar.path="+scanCliImplJarPath,
		"-jar", scanCliJarPath,
		"--host", hsc.host,
		"--port", fmt.Sprintf("%d", hsc.port),
		"--scheme", hubScheme,
		"--project", job.HubProjectName,
		"--release", job.HubProjectVersionName,
		"--username", hsc.username,
		"--name", job.HubScanName,
		"--insecure",
		"-v",
		path)

	log.Infof("running command %+v for image %s\n", cmd, job.Sha)
	startScanClient := time.Now()
	stdoutStderr, err := cmd.CombinedOutput()

	recordScanClientDuration(time.Now().Sub(startScanClient), err == nil)
	recordTotalScannerDuration(time.Now().Sub(startTotal), err == nil)

	if err != nil {
		recordScannerError("scan client failed")
		log.Errorf("java scanner failed for image %s with error %s and output:\n%s\n", job.Sha, err.Error(), string(stdoutStderr))
		return err
	}
	log.Infof("successfully completed java scanner for image %s", job.Sha)
	log.Debugf("output from image %s: %s", job.Sha, stdoutStderr)
	return err
}

// func (hsc *HubScanClient) ScanCliSh(job ScanJob) error {
// 	pathToScanner := "./dependencies/scan.cli-4.3.0/bin/scan.cli.sh"
// 	cmd := exec.Command(pathToScanner,
// 		"--project", job.Image.HubProjectName(),
// 		"--host", hsc.host,
// 		"--port", hubPort,
// 		"--insecure",
// 		"--username", hsc.username,
// 		job.Image.HumanReadableName())
// 	log.Infof("running command %v for image %s\n", cmd, job.Image.HumanReadableName())
// 	stdoutStderr, err := cmd.CombinedOutput()
// 	if err != nil {
// 		message := fmt.Sprintf("failed to run scan.cli.sh: %s", err.Error())
// 		log.Error(message)
// 		log.Errorf("output from scan.cli.sh:\n%v\n", string(stdoutStderr))
// 		return err
// 	}
// 	log.Infof("successfully completed scan.cli.sh: %s", stdoutStderr)
// 	return nil
// }

func cleanUpTarFile(path string) {
	err := os.Remove(path)
	recordCleanUpTarFile(err == nil)
	if err != nil {
		log.Errorf("unable to remove file %s: %s", path, err.Error())
	} else {
		log.Infof("successfully cleaned up file %s", path)
	}
}
