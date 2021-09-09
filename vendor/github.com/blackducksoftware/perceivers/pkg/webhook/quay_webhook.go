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

// QuayRepo contains a quay image with list of tags
type QuayRepo struct {
	Name        string   `json:"name"`
	Repository  string   `json:"repository"`
	Namespace   string   `json:"namespace"`
	DockerURL   string   `json:"docker_url"`
	Homepage    string   `json:"homepage"`
	UpdatedTags []string `json:"updated_tags"`
}

// QuayTagDigest contains Digest for a particular Quay image
type QuayTagDigest struct {
	HasAdditional bool `json:"has_additional"`
	Page          int  `json:"page"`
	Tags          []struct {
		Name           string `json:"name"`
		Reversion      bool   `json:"reversion"`
		StartTs        int    `json:"start_ts"`
		ImageID        string `json:"image_id"`
		LastModified   string `json:"last_modified"`
		ManifestDigest string `json:"manifest_digest"`
		DockerImageID  string `json:"docker_image_id"`
		IsManifestList bool   `json:"is_manifest_list"`
		Size           int    `json:"size"`
	} `json:"tags"`
}

// QuayWebhook handles watching images and sending them to perceptor
type QuayWebhook struct {
	certificate    string
	certificateKey string
	perceptorURL   string
	registryAuths  []*utils.RegistryAuth
}

// NewQuayWebhook creates a new QuayWebhook object
func NewQuayWebhook(perceptorURL string, credentials []*utils.RegistryAuth, certificate string, certificateKey string) *QuayWebhook {
	return &QuayWebhook{
		perceptorURL:   perceptorURL,
		registryAuths:  credentials,
		certificate:    certificate,
		certificateKey: certificateKey,
	}
}

// Run starts a controller that watches images and sends them to perceptor
func (qw *QuayWebhook) Run() {

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			log.Info("Quay webhook incoming!")
			qr := &QuayRepo{}
			json.NewDecoder(r.Body).Decode(qr)
			for _, registry := range qw.registryAuths {
				if strings.Contains(qr.DockerURL, registry.URL) && len(registry.Token) > 0 {
					qw.webhook(registry.Token, qr)
				}
			}
		}
	})

	if len(qw.certificate) > 0 && len(qw.certificateKey) > 0 {
		errC := ioutil.WriteFile("cert", []byte(qw.certificate), 0644)
		errK := ioutil.WriteFile("key", []byte(qw.certificateKey), 0644)
		if errC != nil || errK != nil {
			log.Errorf("Webhook: Writing to a certificate file failed %e %e", errC, errK)
		} else {
			log.Infof("Webhook: Starting HTTPs webhook with TLS enabled for quay on :3002 at /webhook")
			err := http.ListenAndServeTLS(":3002", "cert", "key", nil)
			if err != nil {
				log.Errorf("Webhook: HTTPs listener on port 3002 failed: %e", err)
			}
		}
	} else {
		log.Infof("Webhook: starting HTTP webhook for quay on :3002 at /webhook")
		err := http.ListenAndServe(":3002", nil)
		if err != nil {
			log.Errorf("Webhook: HTTP listener on port 3002 failed: %e", err)
		}
	}
}

func (qw *QuayWebhook) webhook(bearerToken string, qr *QuayRepo) {

	rt := &QuayTagDigest{}
	url := strings.Replace(qr.Homepage, "repository", "api/v1/repository", -1)
	url = fmt.Sprintf("%s/tag", url)
	err := utils.GetResourceOfType(url, nil, bearerToken, rt)
	if err != nil {
		log.Errorf("Webhook: Error in getting docker repo: %+v", err)
	}

	for _, tagDigest := range rt.Tags {
		sha := strings.Replace(tagDigest.ManifestDigest, "sha256:", "", -1)
		priority := 1
		quayImage := perceptorapi.NewImage(qr.DockerURL, tagDigest.Name, sha, &priority, qr.DockerURL, tagDigest.Name)
		imageURL := fmt.Sprintf("%s/%s", qw.perceptorURL, perceptorapi.ImagePath)
		err = communicator.SendPerceptorAddEvent(imageURL, quayImage)
		if err != nil {
			log.Errorf("Webhook: Error putting image %v in perceptor queue %e", quayImage, err)
		} else {
			log.Infof("Webhook: Successfully put image %s with tag %s in perceptor queue", url, tagDigest.Name)
		}
	}

}
