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
	"time"

	"github.com/blackducksoftware/perceivers/pkg/annotations"
	"github.com/blackducksoftware/perceivers/pkg/communicator"
	"github.com/blackducksoftware/perceivers/pkg/docker"
	"github.com/blackducksoftware/perceivers/pkg/metrics"
	"github.com/blackducksoftware/perceivers/pkg/utils"

	perceptorapi "github.com/blackducksoftware/perceptor/pkg/api"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/api/core/v1"

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	log "github.com/sirupsen/logrus"
)

// PodAnnotator handles annotating pods with vulnerability and policy issues
type PodAnnotator struct {
	coreV1         corev1.CoreV1Interface
	scanResultsURL string
	h              annotations.PodAnnotatorHandler
}

// NewPodAnnotator creates a new PodAnnotator object
func NewPodAnnotator(pl corev1.CoreV1Interface, perceptorURL string, handler annotations.PodAnnotatorHandler) *PodAnnotator {
	return &PodAnnotator{
		coreV1:         pl,
		scanResultsURL: fmt.Sprintf("%s/%s", perceptorURL, perceptorapi.ScanResultsPath),
		h:              handler,
	}
}

// Run starts a controller that will annotate pods
func (pa *PodAnnotator) Run(interval time.Duration, stopCh <-chan struct{}) {
	log.Infof("starting pod pod_annotator controller")

	for {
		select {
		case <-stopCh:
			return
		default:
		}

		time.Sleep(interval)

		err := pa.annotate()
		if err != nil {
			log.Errorf("failed to annotate pods: %v", err)
		}
	}
}

func (pa *PodAnnotator) annotate() error {
	// Get all the scan results from the Perceptor
	log.Infof("attempting to get scan results with GET %s for pod annotation", pa.scanResultsURL)
	scanResults, err := pa.getScanResults()
	if err != nil {
		metrics.RecordError("pod_annotator", "error getting scan results")
		return fmt.Errorf("error getting scan results: %v", err)
	}

	// Process the scan results and apply annotations/labels to pods
	log.Infof("GET to %s succeeded, about to update annotations on all pods", pa.scanResultsURL)
	pa.addAnnotationsToPods(*scanResults)
	return nil
}

func (pa *PodAnnotator) getScanResults() (*perceptorapi.ScanResults, error) {
	var results perceptorapi.ScanResults

	bytes, err := communicator.GetPerceptorScanResults(pa.scanResultsURL)
	if err != nil {
		metrics.RecordError("pod_annotator", "unable to get scan results")
		return nil, fmt.Errorf("unable to get scan results: %v", err)
	}

	err = json.Unmarshal(bytes, &results)
	if err != nil {
		metrics.RecordError("pod_annotator", "unable to Unmarshal ScanResults")
		return nil, fmt.Errorf("unable to Unmarshal ScanResults from url %s: %v", pa.scanResultsURL, err)
	}

	return &results, nil
}

func (pa *PodAnnotator) addAnnotationsToPods(results perceptorapi.ScanResults) {
	for _, pod := range results.Pods {
		podName := fmt.Sprintf("%s:%s", pod.Namespace, pod.Name)
		getPodStart := time.Now()
		kubePod, err := pa.coreV1.Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{})
		metrics.RecordDuration("get pod", time.Now().Sub(getPodStart))
		if err != nil {
			metrics.RecordError("pod_annotator", "unable to get pod")
			log.Errorf("unable to get pod %s: %v", podName, err)
			continue
		}

		podAnnotations := annotations.NewPodAnnotationData(pod.PolicyViolations, pod.Vulnerabilities, pod.OverallStatus, "", "")

		// Update the pod if any label or annotation isn't correct
		if pa.addPodAnnotations(kubePod, podAnnotations, results.Images) ||
			pa.addPodLabels(kubePod, podAnnotations, results.Images) {
			updatePodStart := time.Now()
			_, err = pa.coreV1.Pods(pod.Namespace).Update(kubePod)
			metrics.RecordDuration("update pod", time.Now().Sub(updatePodStart))
			if err != nil {
				metrics.RecordError("pod_annotator", "unable to update annotations/labels for pod")
				log.Errorf("unable to update annotations/labels for pod %s: %v", podName, err)
			} else {
				metrics.RecordPodAnnotation("pod_annotator", podName)
				log.Infof("successfully annotated pod %s", podName)
			}
		}
	}
}

func (pa *PodAnnotator) addPodAnnotations(pod *v1.Pod, podAnnotations *annotations.PodAnnotationData, images []perceptorapi.ScannedImage) bool {
	podName := fmt.Sprintf("%s/%s", pod.GetNamespace(), pod.GetName())

	// Get the list of annotations currently on the pod
	getAnnotationsStart := time.Now()
	currentAnnotations := pod.GetAnnotations()
	metrics.RecordDuration("get pod annotations", time.Now().Sub(getAnnotationsStart))
	if currentAnnotations == nil {
		currentAnnotations = map[string]string{}
	}

	// Get the new annotations
	newAnnotations := pa.createNewAnnotations(pod, podAnnotations, images)

	// Apply updated annotations to the pod if the existing annotations don't
	// contain the expected entries
	if !pa.h.CompareMaps(currentAnnotations, newAnnotations) {
		log.Infof("annotations are missing or incorrect on pod %s.  Expected %v to contain %v", podName, currentAnnotations, newAnnotations)
		setAnnotationsStart := time.Now()
		pod.SetAnnotations(utils.MapMerge(currentAnnotations, newAnnotations))
		metrics.RecordDuration("set pod annotations", time.Now().Sub(setAnnotationsStart))
		return true
	}
	return false
}

func (pa *PodAnnotator) createNewAnnotations(pod *v1.Pod, podData *annotations.PodAnnotationData, images []perceptorapi.ScannedImage) map[string]string {
	// Generate the pod level annotations that should be on the pod
	podAnnotations := pa.h.CreatePodAnnotations(podData)

	// Generate the image level annotations that should be on the pod
	imageAnnotations := pa.getPodContainerMap(pod, images, podData.GetHubVersion(), podData.GetScanClientVersion(), pa.h.CreateImageAnnotations)

	// Merge the pod and image level annotations
	return utils.MapMerge(podAnnotations, imageAnnotations)
}

func (pa *PodAnnotator) addPodLabels(pod *v1.Pod, podAnnotations *annotations.PodAnnotationData, images []perceptorapi.ScannedImage) bool {
	podName := fmt.Sprintf("%s/%s", pod.GetNamespace(), pod.GetName())

	// Get the list of labels currently on the pod
	getLabelsStart := time.Now()
	currentLabels := pod.GetLabels()
	metrics.RecordDuration("get pod labels", time.Now().Sub(getLabelsStart))
	if currentLabels == nil {
		currentLabels = map[string]string{}
	}

	// Get the new labels
	newLabels := pa.createNewLabels(pod, podAnnotations, images)

	// Apply updated labels to the pod if the existing labels don't
	// contain the expected entries
	if !pa.h.CompareMaps(currentLabels, newLabels) {
		log.Infof("labels are missing or incorrect on pod %s.  Expected %v to contain %v", podName, currentLabels, newLabels)
		setLabelsStart := time.Now()
		pod.SetLabels(utils.MapMerge(currentLabels, newLabels))
		metrics.RecordDuration("set pod labels", time.Now().Sub(setLabelsStart))
		return true
	}
	return false
}

func (pa *PodAnnotator) createNewLabels(pod *v1.Pod, podAnnotations *annotations.PodAnnotationData, images []perceptorapi.ScannedImage) map[string]string {
	// Generate the pod level labels that should be on the pod
	labels := pa.h.CreatePodLabels(podAnnotations)

	// Generate the image level labels that should be on the pod
	imageLabels := pa.getPodContainerMap(pod, images, podAnnotations.GetHubVersion(), podAnnotations.GetScanClientVersion(), pa.h.CreateImageLabels)

	// Merge the pod and image level annotations
	return utils.MapMerge(labels, imageLabels)
}

func (pa *PodAnnotator) getPodContainerMap(pod *v1.Pod, scannedImages []perceptorapi.ScannedImage, hubVersion string, scVersion string, mapGenerator func(interface{}, string, int) map[string]string) map[string]string {
	containerMap := make(map[string]string)

	for cnt, container := range pod.Status.ContainerStatuses {
		name, sha, err := docker.ParseImageIDString(container.ImageID)
		if err != nil {
			metrics.RecordError("pod_annotator", "unable to parse kubernetes imageID")
			log.Errorf("unable to parse kubernetes imageID string %s from pod %s/%s: %v", container.ImageID, pod.Namespace, pod.Name, err)
			continue
		}
		imageScanResults := pa.findImageAnnotations(name, sha, scannedImages)
		if imageScanResults != nil {
			imageAnnotations := pa.createImageAnnotationsFromImageScanResults(imageScanResults, hubVersion, scVersion)
			containerMap = utils.MapMerge(containerMap, mapGenerator(imageAnnotations, name, cnt))
		}
	}
	return containerMap
}

func (pa *PodAnnotator) findImageAnnotations(imageName string, imageSha string, imageList []perceptorapi.ScannedImage) *perceptorapi.ScannedImage {
	for _, image := range imageList {
		if image.Repository == imageName && image.Sha == imageSha {
			return &image
		}
	}

	return nil
}

func (pa *PodAnnotator) createImageAnnotationsFromImageScanResults(scannedImage *perceptorapi.ScannedImage, hv string, scv string) *annotations.ImageAnnotationData {
	return annotations.NewImageAnnotationData(scannedImage.PolicyViolations,
		scannedImage.Vulnerabilities, scannedImage.OverallStatus, scannedImage.ComponentsURL, hv, scv)
}
