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

package annotator

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/blackducksoftware/perceivers/image/pkg/metrics"
	"github.com/blackducksoftware/perceivers/pkg/annotations"
	"github.com/blackducksoftware/perceivers/pkg/communicator"
	"github.com/blackducksoftware/perceivers/pkg/docker"
	"github.com/blackducksoftware/perceivers/pkg/utils"

	perceptorapi "github.com/blackducksoftware/perceptor/pkg/api"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/api/image/v1"

	imageclient "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"

	log "github.com/sirupsen/logrus"
)

// ImageAnnotator handles annotating images with vulnerability and policy issues
type ImageAnnotator struct {
	client         *imageclient.ImageV1Client
	scanResultsURL string
	h              annotations.ImageAnnotatorHandler
}

// NewImageAnnotator creates a new ImageAnnotator object
func NewImageAnnotator(ic *imageclient.ImageV1Client, perceptorURL string, handler annotations.ImageAnnotatorHandler) *ImageAnnotator {
	return &ImageAnnotator{
		client:         ic,
		scanResultsURL: fmt.Sprintf("%s/%s", perceptorURL, perceptorapi.ScanResultsPath),
		h:              handler,
	}
}

// Run starts a controller that will annotate images
func (ia *ImageAnnotator) Run(interval time.Duration, stopCh <-chan struct{}) {
	log.Infof("starting image annotator controller")

	for {
		select {
		case <-stopCh:
			return
		default:
		}

		time.Sleep(interval)

		err := ia.annotate()
		if err != nil {
			metrics.RecordError("image_controller", "failed to annotate images")
			log.Errorf("failed to annotate images: %v", err)
		}
	}
}

func (ia *ImageAnnotator) annotate() error {
	// Get all the scan results from the Perceptor
	log.Infof("attempting to GET %s for image annotation", ia.scanResultsURL)
	scanResults, err := ia.getScanResults()
	if err != nil {
		metrics.RecordError("image_controller", "error getting scan results")
		return fmt.Errorf("error getting scan results: %v", err)
	}

	// Process the scan results and apply annotations/labels to images
	log.Infof("GET to %s succeeded, about to update annotations on all images", ia.scanResultsURL)
	ia.addAnnotationsToImages(*scanResults)
	return nil
}

func (ia *ImageAnnotator) getScanResults() (*perceptorapi.ScanResults, error) {
	var results perceptorapi.ScanResults

	bytes, err := communicator.GetPerceptorScanResults(ia.scanResultsURL)
	if err != nil {
		metrics.RecordError("image_controller", "failed to annotate images")
		return nil, fmt.Errorf("unable to get scan results: %v", err)
	}

	err = json.Unmarshal(bytes, &results)
	if err != nil {
		metrics.RecordError("image_controller", "unable to Unmarshal ScanResults")
		return nil, fmt.Errorf("unable to Unmarshal ScanResults from url %s: %v", ia.scanResultsURL, err)
	}

	return &results, nil
}

func (ia *ImageAnnotator) addAnnotationsToImages(results perceptorapi.ScanResults) {
	for _, image := range results.Images {
		var imageName string
		getName := fmt.Sprintf("sha256:%s", image.Sha)
		fullImageName := fmt.Sprintf("%s@%s", image.Name, getName)

		nameStart := strings.LastIndex(image.Name, "/") + 1
		if nameStart >= 0 {
			imageName = image.Name[nameStart:]
		} else {
			imageName = image.Name
		}

		// Get the image
		getImageStart := time.Now()
		osImage, err := ia.client.Images().Get(getName, metav1.GetOptions{})
		metrics.RecordDuration("get image", time.Now().Sub(getImageStart))
		if err != nil {
			metrics.RecordError("image_controller", "unable to get image")
		}
		if errors.IsNotFound(err) {
			// This isn't an image in openshift
			continue
		} else if err != nil {
			// Some other kind of error, possibly couldn't communicate, so return
			// an error
			log.Errorf("unexpected error retrieving image %s: %v", fullImageName, err)
			continue
		}

		// Verify the sha of the scanned image matches that of the image we retrieved
		_, imageSha, err := docker.ParseImageIDString(osImage.DockerImageReference)
		if err != nil {
			metrics.RecordError("image_controller", "unable to parse openshift imageID")
			log.Errorf("unable to parse openshift imageID from image %s: %v", imageName, err)
			continue
		}
		if imageSha != image.Sha {
			metrics.RecordError("image_controller", "image sha doesn't match")
			log.Errorf("image sha doesn't match for image %s.  Got %s, expected %s", imageName, image.Sha, imageSha)
			continue
		}

		imageAnnotations := annotations.NewImageAnnotationData(image.PolicyViolations, image.Vulnerabilities, image.OverallStatus, image.ComponentsURL, results.HubVersion, results.HubScanClientVersion)

		// Update the image if any label or annotation isn't correct
		if ia.addImageAnnotations(fullImageName, osImage, imageAnnotations) ||
			ia.addImageLabels(fullImageName, osImage, imageAnnotations) {
			updateImageStart := time.Now()
			_, err = ia.client.Images().Update(osImage)
			metrics.RecordDuration("update image", time.Now().Sub(updateImageStart))
			if err != nil {
				metrics.RecordError("image_controller", "unable to update image")
				log.Errorf("unable to update annotations/labels for image %s: %v", fullImageName, err)
			} else {
				log.Infof("successfully annotated image %s", fullImageName)
			}
		}
	}
}

func (ia *ImageAnnotator) addImageAnnotations(name string, image *v1.Image, imageAnnotations *annotations.ImageAnnotationData) bool {
	// Get existing annotations on the image
	currentAnnotations := image.GetAnnotations()
	if currentAnnotations == nil {
		currentAnnotations = map[string]string{}
	}

	// Generate the annotations that should be on the image
	newAnnotations := ia.h.CreateImageAnnotations(imageAnnotations, "", 0)

	// Apply updated annotations to the image if the existing annotations don't
	// contain the expected entries
	if !ia.h.CompareMaps(currentAnnotations, newAnnotations) {
		metrics.RecordError("image_controller", "annotations are missing or incorrect")
		log.Infof("annotations are missing or incorrect on image %s.  Expected %v to contain %v", name, currentAnnotations, newAnnotations)
		setAnnotationsStart := time.Now()
		image.SetAnnotations(utils.MapMerge(currentAnnotations, newAnnotations))
		metrics.RecordDuration("set image annotations", time.Now().Sub(setAnnotationsStart))
		return true
	}
	return false
}

func (ia *ImageAnnotator) addImageLabels(name string, image *v1.Image, imageAnnotations *annotations.ImageAnnotationData) bool {
	// Get existing labels on the image
	currentLabels := image.GetLabels()
	if currentLabels == nil {
		currentLabels = map[string]string{}
	}

	// Generate the labels that should be on the image
	newLabels := ia.h.CreateImageLabels(imageAnnotations, "", 0)

	// Apply updated labels to the image if the existing annotations don't
	// contain the expected entries
	if !ia.h.CompareMaps(currentLabels, newLabels) {
		metrics.RecordError("image_controller", "labels are missing or incorrect")
		log.Infof("labels are missing or incorrect on image %s.  Expected %v to contain %v", name, currentLabels, newLabels)
		setLabelsStart := time.Now()
		image.SetLabels(utils.MapMerge(currentLabels, newLabels))
		metrics.RecordDuration("set image labels", time.Now().Sub(setLabelsStart))
		return true
	}

	return false
}
