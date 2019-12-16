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

package annotator

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/blackducksoftware/perceivers/pkg/communicator"
	"github.com/blackducksoftware/perceivers/pkg/metrics"
	"github.com/blackducksoftware/perceivers/pkg/utils"

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
	HasAdditional bool      `json:"has_additional"`
	Page          int       `json:"page"`
	Tags          []QuayTag `json:"tags"`
}

// QuayTag contains individual tag info for an image version
type QuayTag struct {
	Name           string `json:"name"`
	Reversion      bool   `json:"reversion"`
	StartTs        int    `json:"start_ts"`
	ImageID        string `json:"image_id"`
	LastModified   string `json:"last_modified"`
	ManifestDigest string `json:"manifest_digest"`
	DockerImageID  string `json:"docker_image_id"`
	IsManifestList bool   `json:"is_manifest_list"`
	Size           int    `json:"size"`
}

// QuayLabels contains a list of returned Labels on an image
type QuayLabels struct {
	Labels []QuayLabel `json:"labels"`
}

// QuayLabel contains info on a single label attached to an image
type QuayLabel struct {
	Value      string `json:"value"`
	MediaType  string `json:"media_type"`
	ID         string `json:"id"`
	Key        string `json:"key"`
	SourceType string `json:"source_type"`
}

// QuayNewLabel is used for Posting a new label,
// doesn't need to have json metadatas but couldn't hurt
type QuayNewLabel struct {
	MediaType string `json:"media_type"`
	Value     string `json:"value"`
	Key       string `json:"key"`
}

// BlackDuck Annotation names
const (
	quayBDPolicy = "blackduck.policyviolations"
	quayBDVuln   = "blackduck.vulnerabilities"
	quayBDSt     = "blackduck.overallstatus"
	quayBDComURL = "blackduck.componentsurl"
)

// QuayAnnotator handles annotating quay images with vulnerability and policy issues
type QuayAnnotator struct {
	client         *http.Client
	scanResultsURL string
	registryAuths  []*utils.RegistryAuth
}

// NewQuayAnnotator creates a new QuayAnnotator object
func NewQuayAnnotator(perceptorURL string, registryAuths []*utils.RegistryAuth) *QuayAnnotator {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr}
	return &QuayAnnotator{
		client:         client,
		scanResultsURL: fmt.Sprintf("%s/%s", perceptorURL, perceptorapi.ScanResultsPath),
		registryAuths:  registryAuths,
	}
}

// Run starts a controller that will annotate images
func (qa *QuayAnnotator) Run(interval time.Duration, stopCh <-chan struct{}) {
	log.Infof("starting quay annotation controller")

	for {
		select {
		case <-stopCh:
			return
		default:
		}

		time.Sleep(interval)

		err := qa.annotate()
		if err != nil {
			log.Errorf("failed to annotate quay images: %v", err)
		}
	}
}

// This method tries to annotate all the images
func (qa *QuayAnnotator) annotate() error {
	// Get all the scan results from the Perceptor
	log.Infof("attempting to GET %s for quay image annotation", qa.scanResultsURL)
	scanResults, err := qa.getScanResults()
	if err != nil {
		metrics.RecordError("quay_annotator", "error getting scan results")
		return fmt.Errorf("error getting scan results: %v", err)
	}

	// Process the scan results and apply annotations/labels to images
	log.Infof("GET to %s succeeded, about to update annotations on all quay images", qa.scanResultsURL)
	qa.addAnnotationsToImages(*scanResults)
	return nil
}

// This method gets the scan results from perceptor and tries to unmarshal it
func (qa *QuayAnnotator) getScanResults() (*perceptorapi.ScanResults, error) {
	var results perceptorapi.ScanResults

	bytes, err := communicator.GetPerceptorScanResults(qa.scanResultsURL)
	if err != nil {
		metrics.RecordError("quay_annotator", "unable to get scan results")
		return nil, fmt.Errorf("unable to get scan results: %v", err)
	}

	err = json.Unmarshal(bytes, &results)
	if err != nil {
		metrics.RecordError("quay_annotator", "unable to Unmarshal ScanResults")
		return nil, fmt.Errorf("unable to Unmarshal ScanResults from url %s: %v", qa.scanResultsURL, err)
	}

	return &results, nil
}

// This method tries to annotate all the Images found in BD by matching their SHAs
func (qa *QuayAnnotator) addAnnotationsToImages(results perceptorapi.ScanResults) {
	regs := 0
	imgs := 0

	for _, registry := range qa.registryAuths {
		auth, err := qa.PingQuayServer("https://"+registry.URL, registry.User, registry.Password, registry.Token)

		if err != nil {
			log.Debugf("Annotator: URL %s either not a valid quay repository or incorrect token: %e", registry.URL, err)
			continue
		}

		regs = regs + 1
		for _, image := range results.Images {

			// The base URL may contain something in their instance/registry, splitting has no loss
			if !strings.Contains(image.Repository, strings.Split(registry.URL, "/")[0]) {
				log.Debugf("Annotator: Registry URL %s does not correspond to scan repo %s", registry.URL, image.Repository)
				continue
			}

			log.Infof("Scan %s corresponds to %s", image.Repository, registry.URL)
			repoSlice := strings.Split(image.Repository, "/")[1:]
			repo := strings.Join(repoSlice, "/")
			labelList := &QuayLabels{}
			// Look for SHA
			url := fmt.Sprintf("%s/api/v1/repository/%s/manifest/%s/labels", auth.URL, repo, fmt.Sprintf("sha256:%s", image.Sha))
			log.Infof("Getting labels from: %s", url)
			err = utils.GetResourceOfType(url, nil, auth.Password, labelList)
			if err != nil {
				log.Errorf("Error in getting labels for repo %s: %e", repo, err)
				continue
			}

			imgs = imgs + 1

			// Create a map of BD tags and retrieved values
			nt := make(map[string]string)
			nt[quayBDComURL] = image.ComponentsURL
			nt[quayBDPolicy] = fmt.Sprintf("%d", image.PolicyViolations)
			nt[quayBDSt] = image.OverallStatus
			nt[quayBDVuln] = fmt.Sprintf("%d", image.Vulnerabilities)

			// Create a map of Quay tags and retrieved values
			ot := make(map[string]string)
			for _, label := range labelList.Labels {
				ot[label.Key] = label.Value
			}

			// Merge them with updated BD values
			tags := utils.MapMerge(ot, nt)
			for key, value := range tags {
				// Don't need to touch other tags apart form BD ones
				if _, ok := nt[key]; ok {
					imageInfo := fmt.Sprintf("%s:%s with SHA %s", image.Repository, image.Tag, image.Sha)
					qa.UpdateAnnotation(url, key, value, imageInfo, registry.Token)
				}
			}

		}

		log.Infof("Total scanned images in Quay with URL %s: %d", registry.URL, imgs)
	}

	log.Infof("Total valid Quay Registries: %d", regs)
}

// UpdateAnnotation takes the specific Quay URL and applies the properties/annotations given by BD
func (qa *QuayAnnotator) UpdateAnnotation(url string, labelKey string, newValue string, imageInfo string, quayToken string) {

	filterURL := fmt.Sprintf("%s?filter=%s", url, labelKey)
	labelList := &QuayLabels{}
	err := utils.GetResourceOfType(filterURL, nil, quayToken, labelList)
	if err != nil {
		log.Errorf("Error in getting labels at URL %s for update: %e", url, err)
		return
	}

	for _, label := range labelList.Labels {
		deleteURL := fmt.Sprintf("%s/%s", url, label.ID)
		err = qa.DeleteQuayLabel(deleteURL, quayToken, label.ID)
		if err != nil {
			log.Errorf("Error in deleting label %s at URL %s: %e", label.Key, deleteURL, err)
			log.Errorf("Images may contain duplicate labels!")
		}
	}

	err = qa.AddQuayLabel(url, quayToken, labelKey, newValue)
	if err != nil {
		log.Errorf("Error in adding label %s at URL %s after deleting: %e", labelKey, url, err)
		return
	}

	labelInfo := fmt.Sprintf("%s:%s", labelKey, newValue)
	log.Infof("Successfully annotated %s with %s!", imageInfo, labelInfo)
}

// PingQuayServer takes in the specified URL with access token and checks weather
// it's a valid token for quay by pinging the server
func (qa *QuayAnnotator) PingQuayServer(url string, user string, password string, accessToken string) (*utils.RegistryAuth, error) {
	url = fmt.Sprintf("%s/api/v1/user", url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Error in creating ping request for quay server %e", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := qa.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error in pinging quay server %+v, response: %+v", err, resp)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		// Making sure that http and https both return not OK
		if strings.Contains(url, "https://") {
			url = strings.Replace(url, "https://", "http://", -1)
			// Reset to baseURL
			url = strings.Replace(url, "/api/v1/user", "", -1)
			return qa.PingQuayServer(url, user, password, accessToken)
		}

		return nil, fmt.Errorf("Error in pinging quay server supposed to get %d response code got %d", http.StatusOK, resp.StatusCode)
	}

	// Reset to baseURL
	url = strings.Replace(url, "/api/v1/user", "", -1)
	return &utils.RegistryAuth{URL: url, User: user, Password: password, Token: accessToken}, nil
}

// AddQuayLabel takes the specific Quay URL and adds the properties/annotations given by BD
func (qa *QuayAnnotator) AddQuayLabel(url string, accessToken string, labelKey string, labelValue string) error {
	quayLabel := QuayNewLabel{MediaType: "text/plain", Value: labelValue, Key: labelKey}
	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(quayLabel)
	req, err := http.NewRequest(http.MethodPost, url, buffer)
	if err != nil {
		return fmt.Errorf("Error in adding label request %e", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := qa.client.Do(req)
	if err != nil {
		return fmt.Errorf("Error in adding label %e", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Successful creation not observer from server %s, status code: %d", url, resp.StatusCode)
	}
	return nil
}

// DeleteQuayLabel takes the specific Quay URL and deletes the properties/annotations given by BD
func (qa *QuayAnnotator) DeleteQuayLabel(url string, accessToken string, labelID string) error {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("Error in deleting label request %e", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := qa.client.Do(req)
	if err != nil {
		return fmt.Errorf("Error in deleting label %e", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Successful deletion not observer from server %s, status code: %d", url, resp.StatusCode)
	}
	return nil
}
