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

package docker

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	dockerSocketPath = "/var/run/docker.sock"

	createStage = "create docker image"
	getStage    = "get docker image"
)

type ImagePuller struct {
	client                   *http.Client
	dockerUser               string
	dockerPassword           string
	internalDockerRegistries []string
}

func NewImagePuller(dockerUser string, dockerPassword string, internalDockerRegistries []string) *ImagePuller {
	fd := func(proto, addr string) (conn net.Conn, err error) {
		return net.Dial("unix", dockerSocketPath)
	}
	tr := &http.Transport{Dial: fd}
	client := &http.Client{Transport: tr}
	return &ImagePuller{
		client:                   client,
		dockerUser:               dockerUser,
		dockerPassword:           dockerPassword,
		internalDockerRegistries: internalDockerRegistries}
}

// PullImage gives us access to a docker image by:
//   1. hitting a docker create endpoint (?)
//   2. pulling down the newly created image and saving as a tarball
// It does this by accessing the host's docker daemon, locally, over the docker
// socket.  This gives us a window into any images that are local.
func (ip *ImagePuller) PullImage(image Image) error {
	start := time.Now()

	err := ip.CreateImageInLocalDocker(image)
	if err != nil {
		log.Errorf("unable to continue processing image %s: %s", image.DockerPullSpec(), err.Error())
		return err
	}
	log.Infof("Processing image: %s", image.DockerPullSpec())

	err = ip.SaveImageToTar(image)
	if err != nil {
		log.Errorf("unable to continue processing image %s: %s", image.DockerPullSpec(), err.Error())
		return err
	}

	recordDockerTotalDuration(time.Now().Sub(start))

	log.Infof("Ready to scan image %s at path %s", image.DockerPullSpec(), image.DockerTarFilePath())
	return nil
}

// CreateImageInLocalDocker could also be implemented using curl:
// this example hits ... ? the default registry?  docker hub?
//   curl --unix-socket /var/run/docker.sock -X POST http://localhost/images/create?fromImage=alpine
// this example hits the kipp registry:
//   curl --unix-socket /var/run/docker.sock -X POST http://localhost/images/create\?fromImage\=registry.kipp.blackducksoftware.com%2Fblackducksoftware%2Fhub-jobrunner%3A4.5.0
//
func (ip *ImagePuller) CreateImageInLocalDocker(image Image) error {
	start := time.Now()
	imageURL := createURL(image)
	log.Infof("Attempting to create %s ......", imageURL)
	req, err := http.NewRequest("POST", imageURL, nil)
	if err != nil {
		log.Errorf("unable to create POST request for image %s: %s", imageURL, err.Error())
		recordDockerError(createStage, "unable to create POST request", image, err)
		return err
	}

	if needsAuthHeader(image, ip.internalDockerRegistries) {
		headerValue := encodeAuthHeader(ip.dockerUser, ip.dockerPassword)
		// log.Infof("X-Registry-Auth value:\n%s\n", headerValue)
		req.Header.Add("X-Registry-Auth", headerValue)

		recordEvent("add auth header")
		log.Debugf("adding auth header for %s", image.DockerPullSpec())

		// // the -n prevents echo from appending a newline
		// fmt.Printf("XRA=`echo -n \"{ \\\"username\\\": \\\"%s\\\", \\\"password\\\": \\\"%s\\\" }\" | base64 --wrap=0`\n", ip.dockerUser, ip.dockerPassword)
		// fmt.Printf("curl -i --unix-socket /var/run/docker.sock -X POST -d \"\" -H \"X-Registry-Auth: %s\" %s\n", headerValue, imageURL)
	} else {
		recordEvent("omit auth header")
		log.Debugf("omitting auth header for %s", image.DockerPullSpec())
	}

	resp, err := ip.client.Do(req)
	if err != nil {
		log.Errorf("Create failed for image %s: %s", imageURL, err.Error())
		recordDockerError(createStage, "POST request failed", image, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		message := fmt.Sprintf("Create may have failed for %s: status code %d, response %+v", imageURL, resp.StatusCode, resp)
		log.Errorf(message)
		recordDockerError(createStage, "POST request failed", image, err)
		return errors.New(message)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		recordDockerError(createStage, "unable to read POST response body", image, err)
		log.Errorf("unable to read response body for %s: %s", imageURL, err.Error())
	}
	log.Debugf("body of POST response from %s: %s", imageURL, string(bodyBytes))

	recordDockerCreateDuration(time.Now().Sub(start))

	return err
}

// SaveImageToTar: part of what it does is to issue an http request similar to the following:
//   curl --unix-socket /var/run/docker.sock -X GET http://localhost/images/openshift%2Forigin-docker-registry%3Av3.6.1/get
func (ip *ImagePuller) SaveImageToTar(image Image) error {
	start := time.Now()
	url := getURL(image)
	log.Infof("Making docker GET image request: %s", url)
	resp, err := ip.client.Get(url)
	if err != nil {
		recordDockerError(getStage, "GET request failed", image, err)
		return err
	} else if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("docker GET failed: received status != 200 from %s: %s", url, resp.Status)
		recordDockerError(getStage, "GET request failed", image, err)
		return err
	}

	log.Infof("docker GET request for image %s successful", url)

	body := resp.Body
	defer func() {
		body.Close()
	}()
	tarFilePath := image.DockerTarFilePath()
	log.Infof("Starting to write file contents to tar file %s", tarFilePath)

	f, err := os.OpenFile(tarFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		recordDockerError(getStage, "unable to create tar file", image, err)
		return err
	}
	if _, err = io.Copy(f, body); err != nil {
		recordDockerError(getStage, "unable to copy tar file", image, err)
		return err
	}

	recordDockerGetDuration(time.Now().Sub(start))

	// What's the right way to get the size of the file?
	//  1. resp.ContentLength
	//  2. check the size of the file after it's written
	// fileSizeInMBs := int(resp.ContentLength / (1024 * 1024))
	stats, err := os.Stat(tarFilePath)

	if err != nil {
		recordDockerError(getStage, "unable to get tar file stats", image, err)
		return err
	}

	fileSizeInMBs := int(stats.Size() / (1024 * 1024))
	recordTarFileSize(fileSizeInMBs)

	return nil
}
