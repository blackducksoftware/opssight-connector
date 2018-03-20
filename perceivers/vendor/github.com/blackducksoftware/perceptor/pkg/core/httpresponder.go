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

package core

import (
	"net/http"
	"sync"

	api "github.com/blackducksoftware/perceptor/pkg/api"
	log "github.com/sirupsen/logrus"
)

// HTTPResponder ...
type HTTPResponder struct {
	AddPodChannel                 chan Pod
	UpdatePodChannel              chan Pod
	DeletePodChannel              chan string
	AddImageChannel               chan Image
	AllPodsChannel                chan []Pod
	AllImagesChannel              chan []Image
	PostNextImageChannel          chan func(*Image)
	PostFinishScanJobChannel      chan api.FinishedScanClientJob
	SetConcurrentScanLimitChannel chan int
	GetModelChannel               chan func(json string)
	GetScanResultsChannel         chan func(scanResults api.ScanResults)
}

func NewHTTPResponder() *HTTPResponder {
	return &HTTPResponder{
		AddPodChannel:                 make(chan Pod),
		UpdatePodChannel:              make(chan Pod),
		DeletePodChannel:              make(chan string),
		AddImageChannel:               make(chan Image),
		AllPodsChannel:                make(chan []Pod),
		AllImagesChannel:              make(chan []Image),
		PostNextImageChannel:          make(chan func(*Image)),
		PostFinishScanJobChannel:      make(chan api.FinishedScanClientJob),
		SetConcurrentScanLimitChannel: make(chan int),
		GetModelChannel:               make(chan func(json string)),
		GetScanResultsChannel:         make(chan func(api.ScanResults))}
}

func (hr *HTTPResponder) GetModel() string {
	var wg sync.WaitGroup
	wg.Add(1)
	var modelString string
	hr.GetModelChannel <- func(json string) {
		modelString = json
		wg.Done()
	}
	wg.Wait()
	return modelString
}

func (hr *HTTPResponder) AddPod(apiPod api.Pod) {
	recordAddPod()
	pod := *newPod(apiPod)
	hr.AddPodChannel <- pod
	log.Infof("handled add pod %s -- %s", pod.UID, pod.QualifiedName())
}

func (hr *HTTPResponder) DeletePod(qualifiedName string) {
	recordDeletePod()
	hr.DeletePodChannel <- qualifiedName
	log.Infof("handled delete pod %s", qualifiedName)
}

func (hr *HTTPResponder) UpdatePod(apiPod api.Pod) {
	recordUpdatePod()
	pod := *newPod(apiPod)
	hr.UpdatePodChannel <- pod
	log.Infof("handled update pod %s -- %s", pod.UID, pod.QualifiedName())
}

func (hr *HTTPResponder) AddImage(apiImage api.Image) {
	recordAddImage()
	image := *newImage(apiImage)
	hr.AddImageChannel <- image
	log.Infof("handled add image %s", image.HumanReadableName())
}

func (hr *HTTPResponder) UpdateAllPods(allPods api.AllPods) {
	recordAllPods()
	pods := []Pod{}
	for _, apiPod := range allPods.Pods {
		pods = append(pods, *newPod(apiPod))
	}
	hr.AllPodsChannel <- pods
	log.Infof("handled update all pods -- %d pods", len(allPods.Pods))
}

func (hr *HTTPResponder) UpdateAllImages(allImages api.AllImages) {
	recordAllImages()
	images := []Image{}
	for _, apiImage := range allImages.Images {
		images = append(images, *newImage(apiImage))
	}
	hr.AllImagesChannel <- images
	log.Infof("handled update all images -- %d images", len(allImages.Images))
}

// GetScanResults returns results for:
//  - all images that have a scan status of complete
//  - all pods for which all their images have a scan status of complete
func (hr *HTTPResponder) GetScanResults() api.ScanResults {
	recordGetScanResults()
	var wg sync.WaitGroup
	wg.Add(1)
	var scanResults api.ScanResults
	hr.GetScanResultsChannel <- func(results api.ScanResults) {
		wg.Done()
		scanResults = results
	}
	wg.Wait()
	return scanResults
}

func (hr *HTTPResponder) GetNextImage() api.NextImage {
	recordGetNextImage()
	var wg sync.WaitGroup
	var nextImage api.NextImage
	wg.Add(1)
	hr.PostNextImageChannel <- func(image *Image) {
		imageString := "null"
		var imageSpec *api.ImageSpec
		if image != nil {
			imageString = image.HumanReadableName()
			imageSpec = api.NewImageSpec(
				image.PullSpec(),
				string(image.Sha),
				image.HubProjectName(),
				image.HubProjectVersionName(),
				image.HubScanName())
		}
		nextImage = *api.NewNextImage(imageSpec)
		log.Infof("handled GET next image -- %s", imageString)
		wg.Done()
	}
	wg.Wait()
	return nextImage
}

func (hr *HTTPResponder) PostFinishScan(job api.FinishedScanClientJob) {
	recordPostFinishedScan()
	hr.PostFinishScanJobChannel <- job
	log.Infof("handled finished scan job -- %v", job)
}

// internal use

func (hr *HTTPResponder) SetConcurrentScanLimit(limit api.SetConcurrentScanLimit) {
	hr.SetConcurrentScanLimitChannel <- limit.Limit
	log.Infof("handled set concurrent scan limit -- %d", limit)
}

// errors

func (hr *HTTPResponder) NotFound(w http.ResponseWriter, r *http.Request) {
	log.Errorf("HTTPResponder not found from request %+v", r)
	recordHTTPNotFound(r)
	http.NotFound(w, r)
}

func (hr *HTTPResponder) Error(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	log.Errorf("HTTPResponder error %s with code %d from request %+v", err.Error(), statusCode, r)
	recordHTTPError(r, err, statusCode)
	http.Error(w, err.Error(), statusCode)
}
