/*
Copyright (C) 2019 Synopsys, Inc.

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

package controller

import (
	"fmt"
	"time"

	"github.com/blackducksoftware/perceivers/pkg/communicator"
	utils "github.com/blackducksoftware/perceivers/pkg/utils"
	perceptorapi "github.com/blackducksoftware/perceptor/pkg/api"

	log "github.com/sirupsen/logrus"
)

// ArtifactoryController handles watching images and sending them to perceptor
type ArtifactoryController struct {
	perceptorURL  string
	registryAuths []*utils.RegistryAuth
}

// NewArtifactoryController creates a new ArtifactoryController object
func NewArtifactoryController(perceptorURL string, credentials []*utils.RegistryAuth) *ArtifactoryController {
	return &ArtifactoryController{
		perceptorURL:  perceptorURL,
		registryAuths: credentials,
	}
}

// Run starts a controller that watches images and sends them to perceptor
func (ic *ArtifactoryController) Run(interval time.Duration, stopCh <-chan struct{}) {
	log.Infof("Controller: starting artifactory controller")
	for {
		select {
		case <-stopCh:
			return
		default:
		}

		err := ic.imageLookup()
		if err != nil {
			log.Errorf("Controller: failed to add artifactory images to scan queue: %v", err)
		}

		time.Sleep(interval)
	}
}

func (ic *ArtifactoryController) imageLookup() error {
	log.Infof("Controller: Total %d private registries credentials found!", len(ic.registryAuths))
	for _, registry := range ic.registryAuths {

		cred, err := utils.PingArtifactoryServer("https://"+registry.URL, registry.User, registry.Password)
		if err != nil {
			log.Debugf("Controller: URL %s either not a valid Artifactory repository or incorrect credentials: %e", registry.URL, err)
			continue
		}

		dockerRepos := &utils.ArtDockerRepo{}
		images := &utils.ArtImages{}
		imageTags := &utils.ArtImageTags{}
		imageSHAs := &utils.ArtImageSHAs{}

		url := fmt.Sprintf("%s/api/repositories?packageType=docker", cred.URL)
		err = utils.GetResourceOfType(url, cred, "", dockerRepos)
		if err != nil {
			log.Errorf("Controller: Error in getting docker repo: %e", err)
			continue
		}

		for _, repo := range *dockerRepos {
			url = fmt.Sprintf("%s/api/docker/%s/v2/_catalog", cred.URL, repo.Key)
			err = utils.GetResourceOfType(url, cred, "", images)
			if err != nil {
				log.Errorf("Controller: Error in getting catalog in repo: %e", err)
				continue
			}

			for _, image := range images.Repositories {
				url = fmt.Sprintf("%s/api/docker/%s/v2/%s/tags/list", cred.URL, repo.Key, image)
				err = utils.GetResourceOfType(url, cred, "", imageTags)
				if err != nil {
					log.Errorf("Controller: Error in getting image: %e", err)
					continue
				}

				for _, tag := range imageTags.Tags {
					url = fmt.Sprintf("%s/api/storage/%s/%s/%s/manifest.json?properties=sha256", cred.URL, repo.Key, image, tag)
					err = utils.GetResourceOfType(url, cred, "", imageSHAs)
					if err != nil {
						log.Errorf("Controller: Error in getting SHAs of the artifactory image: %e", err)
						continue
					}

					for _, sha := range imageSHAs.Properties.Sha256 {

						// Remove Tag & HTTPS because image model doesn't require it
						url = fmt.Sprintf("%s/%s/%s", registry.URL, repo.Key, image)
						priority := 1
						artImage := perceptorapi.NewImage(url, tag, sha, &priority, url, tag)
						imageURL := fmt.Sprintf("%s/%s", ic.perceptorURL, perceptorapi.ImagePath)
						err = communicator.SendPerceptorAddEvent(imageURL, artImage)
						if err != nil {
							log.Errorf("Controller: Error putting artifactory image %v in perceptor queue %e", artImage, err)
						} else {
							log.Infof("Controller: Successfully put image %s with tag %s in perceptor queue", url, tag)
						}

					}
				}
			}
		}

		log.Infof("Controller: There were total %d docker repositories found in artifactory instance %s.", len(images.Repositories), registry.URL)

	}

	return nil
}
