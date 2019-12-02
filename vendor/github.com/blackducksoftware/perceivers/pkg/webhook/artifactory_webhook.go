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

package webhook

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/blackducksoftware/perceivers/pkg/communicator"
	utils "github.com/blackducksoftware/perceivers/pkg/utils"
	perceptorapi "github.com/blackducksoftware/perceptor/pkg/api"
	log "github.com/sirupsen/logrus"
)

// ArtifactoryWebhook handles watching images and sending them to perceptor
type ArtifactoryWebhook struct {
	perceptorURL   string
	registryAuths  []*utils.RegistryAuth
	certificate    string
	certificateKey string
}

// NewArtifactoryWebhook creates a new ArtifactoryWebhook object
func NewArtifactoryWebhook(perceptorURL string, credentials []*utils.RegistryAuth, certificate string, certificateKey string) *ArtifactoryWebhook {
	return &ArtifactoryWebhook{
		perceptorURL:   perceptorURL,
		registryAuths:  credentials,
		certificate:    certificate,
		certificateKey: certificateKey,
	}
}

// Run starts a controller that watches images and sends them to perceptor
func (aw *ArtifactoryWebhook) Run() {

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			log.Info("Webhook: Artifactory hook incoming!")
			ahs := &utils.ArtHookStruct{}
			json.NewDecoder(r.Body).Decode(ahs)
			for _, registry := range aw.registryAuths {
				cred, err := utils.PingArtifactoryServer("https://"+registry.URL, registry.User, registry.Password)
				if err != nil {
					log.Debugf("Webhook: URL %s either not a valid Artifactory repository or incorrect credentials: %e", registry.URL, err)
					continue
				}
				aw.webhook(ahs, cred, aw.perceptorURL)
			}
		}
	})

	if len(aw.certificate) > 0 && len(aw.certificateKey) > 0 {
		errC := ioutil.WriteFile("cert", []byte(aw.certificate), 0644)
		errK := ioutil.WriteFile("key", []byte(aw.certificateKey), 0644)
		if errC != nil || errK != nil {
			log.Errorf("Webhook: Writing to a certificate file failed %e %e", errC, errK)
		} else {
			log.Infof("Webhook: Starting HTTPs webhook with TLS enabled for artifactory on :3002 at /webhook")
			err := http.ListenAndServeTLS(":3002", "cert", "key", nil)
			if err != nil {
				log.Errorf("Webhook: HTTPs listener on port 3002 failed: %e", err)
			}
		}
	} else {
		log.Infof("Webhook: starting HTTP webhook for artifactory on :3002 at /webhook")
		err := http.ListenAndServe(":3002", nil)
		if err != nil {
			log.Errorf("Webhook: HTTP listener on port 3002 failed: %e", err)
		}
	}
}

func (aw *ArtifactoryWebhook) webhook(ahs *utils.ArtHookStruct, cred *utils.RegistryAuth, perceptorURL string) {
	for _, a := range ahs.Artifacts {
		// Trying to find the repo key, cannot split because image may contain '/'
		// So stripping down the returned URL by removing everything
		returnedURL := a.Reference
		if strings.Contains(returnedURL, "null") {
			log.Errorf("Webhook: The plugin needs to setup BaseURL in the webhook json")
		}
		woBase := strings.Replace(returnedURL, cred.URL+"/", "", -1)
		woRepo := strings.Replace(woBase, "/"+a.Name, "", -1)
		repoKey := strings.Replace(woRepo, ":"+a.Version, "", -1)

		imageSHAs := &utils.ArtImageSHAs{}
		url := fmt.Sprintf("%s/api/storage/%s/%s/%s/manifest.json?properties=sha256", cred.URL, repoKey, a.Name, a.Version)
		err := utils.GetResourceOfType(url, cred, "", imageSHAs)
		if err != nil {
			log.Errorf("Webhook: Error in getting SHAs of the artifactory image: %e", err)
			continue
		}
		for _, sha := range imageSHAs.Properties.Sha256 {

			// Remove Tag & HTTPS, /artifactory because image model doesn't require it
			url = fmt.Sprintf("%s/%s/%s", cred.URL, repoKey, a.Name)
			url = strings.Replace(url, "http://", "", -1)
			url = strings.Replace(url, "https://", "", -1)
			url = strings.Replace(url, "/artifactory", "", -1)
			priority := 1
			artImage := perceptorapi.NewImage(url, a.Version, sha, &priority, url, a.Version)

			imageURL := fmt.Sprintf("%s/%s", perceptorURL, perceptorapi.ImagePath)
			err = communicator.SendPerceptorAddEvent(imageURL, artImage)
			if err != nil {
				log.Errorf("Webhook: Error putting artifactory image %v in perceptor queue %e", artImage, err)
			} else {
				log.Infof("Webhook: Successfully put image %s with tag %s in perceptor queue", url, a.Version)
			}

		}

	}
}
