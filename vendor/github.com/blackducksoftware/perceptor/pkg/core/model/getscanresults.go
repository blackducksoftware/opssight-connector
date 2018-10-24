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

package model

import (
	"fmt"

	"github.com/blackducksoftware/perceptor/pkg/api" // TODO I hate how this package depends on the api package
	"github.com/blackducksoftware/perceptor/pkg/hub"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
)

// GetScanResults .....
type GetScanResults struct {
	Done chan api.ScanResults
}

// NewGetScanResults ...
func NewGetScanResults() *GetScanResults {
	return &GetScanResults{Done: make(chan api.ScanResults)}
}

// Apply .....
func (g *GetScanResults) Apply(model *Model) error {
	scanResults, err := ScanResults(model)
	go func() {
		g.Done <- scanResults
	}()
	return err
}

func scanResultsForPod(model *Model, podName string) (*Scan, error) {
	pod, ok := model.Pods[podName]
	if !ok {
		return nil, fmt.Errorf("could not find pod of name %s in cache", podName)
	}

	overallStatus := hub.PolicyStatusTypeNotInViolation
	policyViolationCount := 0
	vulnerabilityCount := 0
	for _, container := range pod.Containers {
		imageScan, err := scanResultsForImage(model, container.Image.Sha)
		if err != nil {
			return nil, errors.Annotatef(err, "unable to get scan results for image %s", container.Image.Sha)
		}
		if imageScan == nil {
			return nil, nil
		}
		policyViolationCount += imageScan.PolicyViolations
		vulnerabilityCount += imageScan.Vulnerabilities
		imageScanOverallStatus := imageScan.OverallStatus
		if imageScanOverallStatus != hub.PolicyStatusTypeNotInViolation {
			overallStatus = imageScanOverallStatus
		}
	}
	podScan := &Scan{
		OverallStatus:    overallStatus,
		PolicyViolations: policyViolationCount,
		Vulnerabilities:  vulnerabilityCount}
	return podScan, nil
}

func scanResultsForImage(model *Model, sha DockerImageSha) (*Scan, error) {
	imageInfo, ok := model.Images[sha]
	if !ok {
		return nil, fmt.Errorf("could not find image of sha %s in cache", sha)
	}

	if imageInfo.ScanStatus != ScanStatusComplete {
		return nil, nil
	}
	if imageInfo.ScanResults == nil {
		return nil, fmt.Errorf("model inconsistency: could not find scan results for completed image %s", sha)
	}

	imageScan := &Scan{
		OverallStatus:    imageInfo.ScanResults.OverallStatus(),
		PolicyViolations: imageInfo.ScanResults.PolicyViolationCount(),
		Vulnerabilities:  imageInfo.ScanResults.VulnerabilityCount()}
	return imageScan, nil
}

// ScanResults .....
func ScanResults(model *Model) (api.ScanResults, error) {
	errors := []error{}
	// pods
	pods := []api.ScannedPod{}
	for podName, pod := range model.Pods {
		podScan, err := scanResultsForPod(model, podName)
		if err != nil {
			errors = append(errors, fmt.Errorf("unable to retrieve scan results for Pod %s: %s", podName, err.Error()))
			continue
		}
		if podScan == nil {
			log.Debugf("image scans not complete for pod %s, skipping", podName)
			continue
		}
		pods = append(pods, api.ScannedPod{
			Namespace:        pod.Namespace,
			Name:             pod.Name,
			PolicyViolations: podScan.PolicyViolations,
			Vulnerabilities:  podScan.Vulnerabilities,
			OverallStatus:    podScan.OverallStatus.String()})
	}

	// images
	images := []api.ScannedImage{}
	for sha, imageInfo := range model.Images {
		if imageInfo.ScanStatus != ScanStatusComplete {
			continue
		}
		if imageInfo.ScanResults == nil {
			errors = append(errors, fmt.Errorf("model inconsistency: found ScanStatusComplete for image %s, but nil ScanResults (imageInfo %+v)", sha, imageInfo))
			continue
		}
		image := imageInfo.Image()
		apiImage := api.ScannedImage{
			Repository:       image.Repository,
			Tag:              image.Tag,
			Sha:              string(image.Sha),
			PolicyViolations: imageInfo.ScanResults.PolicyViolationCount(),
			Vulnerabilities:  imageInfo.ScanResults.VulnerabilityCount(),
			OverallStatus:    imageInfo.ScanResults.OverallStatus().String(),
			ComponentsURL:    imageInfo.ScanResults.ComponentsHref}
		images = append(images, apiImage)
	}

	return *api.NewScanResults(pods, images), combineErrors("scanResults", errors)
}
