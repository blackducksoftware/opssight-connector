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

package skopeo

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/blackducksoftware/perceptor-scanner/pkg/common"
	imageInterface "github.com/blackducksoftware/perceptor-scanner/pkg/interfaces"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
)

const (
	dockerSocketPath = "/var/run/docker.sock"

	copyStage = "copy docker image"
	getStage  = "get docker image"
)

// ImagePuller ...
type ImagePuller struct {
	registries []common.RegistryAuth
}

// NewImagePuller ...
func NewImagePuller(registries []common.RegistryAuth) *ImagePuller {
	log.Infof("creating Skopeo image puller")
	return &ImagePuller{registries: registries}
}

// PullImage gives us access to a docker image by:
//   1. hitting a docker create endpoint (?)
//   2. pulling down the newly created image and saving as a tarball
// It does this by accessing the host's docker daemon, locally, over the docker
// socket.  This gives us a window into any images that are local.
func (ip *ImagePuller) PullImage(image imageInterface.Image) error {
	start := time.Now()
	log.Infof("Processing image: %s in %s", image.DockerPullSpec(), image.DockerTarFilePath())

	err := ip.SaveImageToTar(image)
	if err != nil {
		return errors.Annotatef(err, "unable to save image %s to tar file", image.DockerPullSpec())
	}

	common.RecordDockerTotalDuration(time.Now().Sub(start))

	log.Infof("Ready to scan image %s at path %s", image.DockerPullSpec(), image.DockerTarFilePath())
	return nil
}

// CreateImageInLocalDocker could also be implemented using curl:
// this example hits ... ? the default registry?  docker hub?
//   curl --unix-socket /var/run/docker.sock -X POST http://localhost/images/create?fromImage=alpine
// this example hits the kipp registry:
//   curl --unix-socket /var/run/docker.sock -X POST http://localhost/images/create\?fromImage\=registry.kipp.blackducksoftware.com%2Fblackducksoftware%2Fhub-jobrunner%3A4.5.0
//
func (ip *ImagePuller) CreateImageInLocalDocker(image imageInterface.Image) error {
	start := time.Now()
	dockerPullSpec := image.DockerPullSpec()
	log.Infof("Attempting to create %s ......", dockerPullSpec)

	authHeader := ip.needAuthHeader(image)
	var headerValue string
	if strings.Compare(authHeader, "") != 0 {
		headerValue = fmt.Sprintf("--src-creds=%s", authHeader)
	}

	var cmd *exec.Cmd
	if len(headerValue) > 0 {
		cmd = exec.Command("skopeo",
			"--insecure-policy",
			"copy",
			headerValue,
			fmt.Sprintf("docker://%s", dockerPullSpec),
			fmt.Sprintf("docker-daemon:%s", dockerPullSpec))
	} else {
		cmd = exec.Command("skopeo",
			"--insecure-policy",
			"copy",
			fmt.Sprintf("docker://%s", dockerPullSpec),
			fmt.Sprintf("docker-daemon:%s", dockerPullSpec))
	}

	log.Infof("running skopeo copy command %+v", cmd)
	stdoutStderr, err := cmd.CombinedOutput()

	if err != nil {
		common.RecordDockerError(copyStage, "skopeo copy failed", image, err)
		log.Errorf("skopeo copy command failed for %s with error %s and output:\n%s\n", dockerPullSpec, err.Error(), string(stdoutStderr))
		return errors.Annotatef(err, "Create failed for image %s", dockerPullSpec)
	}

	common.RecordDockerCreateDuration(time.Now().Sub(start))

	err = ip.recordTarFileSize(image)

	return err
}

// SaveImageToTar -- part of what it does is to issue an http request similar to the following:
//   curl --unix-socket /var/run/docker.sock -X GET http://localhost/images/openshift%2Forigin-docker-registry%3Av3.6.1/get
func (ip *ImagePuller) SaveImageToTar(image imageInterface.Image) error {
	start := time.Now()
	dockerPullSpec := image.DockerPullSpec()
	log.Infof("Attempting to create %s ......", dockerPullSpec)

	authHeader := ip.needAuthHeader(image)
	var headerValue string
	if strings.Compare(authHeader, "") != 0 {
		headerValue = fmt.Sprintf("--src-creds=%s", authHeader)
	}

	tarFilePath := image.DockerTarFilePath()

	var cmd *exec.Cmd
	if len(headerValue) > 0 {
		cmd = exec.Command("skopeo",
			"--insecure-policy",
			"copy",
			headerValue,
			fmt.Sprintf("docker://%s", dockerPullSpec),
			fmt.Sprintf("docker-archive:%s", tarFilePath))
	} else {
		cmd = exec.Command("skopeo",
			"--insecure-policy",
			"copy",
			fmt.Sprintf("docker://%s", dockerPullSpec),
			fmt.Sprintf("docker-archive:%s", tarFilePath))
	}

	log.Infof("running skopeo copy command %+v", cmd)

	stdoutStderr, err := cmd.CombinedOutput()

	if err != nil {
		common.RecordDockerError(copyStage, "skopeo copy failed", image, err)
		log.Errorf("skopeo copy command failed for %s with error: %s, stdouterr: %s", dockerPullSpec, err.Error(), string(stdoutStderr))
		return errors.Annotatef(err, "Create failed for image %s", dockerPullSpec)
	}

	common.RecordDockerGetDuration(time.Now().Sub(start))

	err = ip.recordTarFileSize(image)

	return err
}

func (ip *ImagePuller) needAuthHeader(image imageInterface.Image) string {
	var headerValue string
	dockerPullSpec := image.DockerPullSpec()
	if registryAuth := common.NeedsAuthHeader(image, ip.registries); registryAuth != nil {
		headerValue = fmt.Sprintf("%s:%s", registryAuth.User, registryAuth.Password)

		common.RecordEvent("add auth header")
		log.Debugf("adding auth header for %s", dockerPullSpec)

		// // the -n prevents echo from appending a newline
		// fmt.Printf("XRA=`echo -n \"{ \\\"username\\\": \\\"%s\\\", \\\"password\\\": \\\"%s\\\" }\" | base64 --wrap=0`\n", ip.dockerUser, ip.dockerPassword)
		// fmt.Printf("curl -i --unix-socket /var/run/docker.sock -X POST -d \"\" -H \"X-Registry-Auth: %s\" %s\n", headerValue, imageURL)
	} else {
		common.RecordEvent("omit auth header")
		log.Debugf("omitting auth header for %s", dockerPullSpec)
	}
	return headerValue
}

func (ip *ImagePuller) recordTarFileSize(image imageInterface.Image) error {
	// What's the right way to get the size of the file?
	//  1. resp.ContentLength
	//  2. check the size of the file after it's written
	// fileSizeInMBs := int(resp.ContentLength / (1024 * 1024))
	stats, err := os.Stat(image.DockerTarFilePath())

	if err != nil {
		common.RecordDockerError(getStage, "unable to get tar file stats", image, err)
		return err
	}

	fileSizeInMBs := int(stats.Size() / (1024 * 1024))
	common.RecordTarFileSize(fileSizeInMBs)
	return nil
}
