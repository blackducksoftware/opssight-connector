/*
Copyright (C) 2018 Black Duck Software, Inc.

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

package report

import (
	"fmt"
	"strings"

	"github.com/blackducksoftware/perceptor-skyfire/pkg/kube"
	log "github.com/sirupsen/logrus"
)

type KubePerceptorReport struct {
	JustKubePods        []string
	JustPerceptorPods   []string
	JustKubeImages      []string
	JustPerceptorImages []string

	// TODO:
	// In kube/openshift but not in perceptor scan results:
	// - finished pods (with annotations/labels)
	// - finished images (with annotations/labels)
	FinishedJustKubePods []string

	// In perceptor scan results but not in kube/openshift:
	// - scanned pods
	// - scanned images
	FinishedJustPerceptorPods []string

	ConflictingAnnotationsPods []string
	ConflictingLabelsPods      []string

	UnanalyzeablePods []string
}

func NewKubePerceptorReport(dump *Dump) *KubePerceptorReport {
	finishedJustKubePods, conflictingAnnotationsPods, conflictingLabelsPods, unanalyzeablePods := KubeNotPerceptorFinishedPods(dump)
	return &KubePerceptorReport{
		JustKubePods:               KubeNotPerceptorPods(dump),
		JustPerceptorPods:          PerceptorNotKubePods(dump),
		JustKubeImages:             KubeNotPerceptorImages(dump),
		JustPerceptorImages:        PerceptorNotKubeImages(dump),
		FinishedJustKubePods:       finishedJustKubePods,
		FinishedJustPerceptorPods:  PerceptorNotKubeFinishedPods(dump),
		ConflictingAnnotationsPods: conflictingAnnotationsPods,
		ConflictingLabelsPods:      conflictingLabelsPods,
		UnanalyzeablePods:          unanalyzeablePods,
	}
}

func (kr *KubePerceptorReport) HumanReadableString() string {
	return fmt.Sprintf(`
Kubernetes<->Perceptor:
 - %d pod(s) in Kubernetes that were not in Perceptor
 - %d pod(s) in Perceptor that were not in Kubernetes
 - %d image(s) in Kubernetes that were not in Perceptor
 - %d image(s) in Perceptor that were not in Kubernetes
 - %d pod(s) whose Kubernetes annotations did not match their scan results
 - %d pod(s) whose Kubernetes labels did not match their scan results
 - %d pod(s) with Kubernetes annotations but no scan results
 - %d pod(s) with scan results but not Kubernetes annotations
	 `,
		len(kr.JustKubePods),
		len(kr.JustPerceptorPods),
		len(kr.JustKubeImages),
		len(kr.JustPerceptorImages),
		len(kr.ConflictingAnnotationsPods),
		len(kr.ConflictingLabelsPods),
		len(kr.FinishedJustKubePods),
		len(kr.FinishedJustPerceptorPods))
}

func KubeNotPerceptorPods(dump *Dump) []string {
	pods := []string{}
	for podName := range dump.Kube.PodsByName {
		_, ok := dump.Perceptor.Model.Pods[podName]
		if !ok {
			pods = append(pods, podName)
		}
	}
	return pods
}

func PerceptorNotKubePods(dump *Dump) []string {
	pods := []string{}
	for podName := range dump.Perceptor.Model.Pods {
		_, ok := dump.Kube.PodsByName[podName]
		if !ok {
			pods = append(pods, podName)
		}
	}
	return pods
}

func KubeNotPerceptorImages(dump *Dump) []string {
	images := []string{}
	for sha := range dump.Kube.ImagesBySha {
		_, ok := dump.Perceptor.Model.Images[sha]
		if !ok {
			images = append(images, sha)
		}
	}
	return images
}

func PerceptorNotKubeImages(dump *Dump) []string {
	images := []string{}
	for sha := range dump.Perceptor.Model.Images {
		_, ok := dump.Kube.ImagesBySha[sha]
		if !ok {
			images = append(images, sha)
		}
	}
	return images
}

func KubeNotPerceptorFinishedPods(dump *Dump) (finishedKubePods []string, incorrectAnnotationsPods []string, incorrectLabelsPods []string, unanalyzeablePods []string) {
	finishedKubePods = []string{}
	incorrectAnnotationsPods = []string{}
	incorrectLabelsPods = []string{}
	unanalyzeablePods = []string{}

	for podName, pod := range dump.Kube.PodsByName {
		imageShas, err := PodShas(pod)
		if err != nil {
			unanalyzeablePods = append(unanalyzeablePods, podName)
			continue
		}

		if pod.HasAllBDAnnotations() && pod.HasAllBDLabels() {
			_, ok := dump.Perceptor.PodsByName[podName]
			if !ok {
				finishedKubePods = append(finishedKubePods, podName)
			}
		}

		expectedPodAnnotations, err := ExpectedPodAnnotations(podName, imageShas, dump)
		if err == nil {
			missingKeys := []string{} // TODO do we actually need this?
			keysOfWrongValues := []string{}
			for key, expectedVal := range expectedPodAnnotations {
				actualVal, ok := pod.Annotations[key]
				if !ok {
					missingKeys = append(missingKeys, key)
				} else if expectedVal != actualVal {
					keysOfWrongValues = append(keysOfWrongValues, key)
				}
			}

			if len(keysOfWrongValues) > 0 {
				incorrectAnnotationsPods = append(incorrectAnnotationsPods, podName)
			}
		} else {
			unanalyzeablePods = append(unanalyzeablePods, podName)
		}

		expectedPodLabels, err := ExpectedPodLabels(podName, imageShas, dump)
		if err == nil {
			missingKeys := []string{} // TODO do we actually need this?
			keysOfWrongValues := []string{}
			for key, expectedVal := range expectedPodLabels {
				actualVal, ok := pod.Labels[key]
				if !ok {
					missingKeys = append(missingKeys, key)
				} else if expectedVal != actualVal {
					log.Warnf("conflicting values for key %s: expected %s, actual %s", key, expectedVal, actualVal)
					keysOfWrongValues = append(keysOfWrongValues, key)
				}
			}

			if len(keysOfWrongValues) > 0 {
				incorrectLabelsPods = append(incorrectLabelsPods, podName)
			}
		} else {
			unanalyzeablePods = append(unanalyzeablePods, podName)
		}
	}
	return
}

func PerceptorNotKubeFinishedPods(dump *Dump) []string {
	pods := []string{}
	for podName, _ := range dump.Perceptor.PodsByName {
		kubePod, ok := dump.Kube.PodsByName[podName]
		if !ok {
			// this should be handled elsewhere, right?
			continue
		}
		if !(kubePod.HasAllBDAnnotations() && kubePod.HasAllBDLabels()) {
			pods = append(pods, podName)
		}
	}
	return pods
}

func PodShas(pod *kube.Pod) ([]string, error) {
	imageShas := []string{}
	for _, cont := range pod.Containers {
		_, sha, err := cont.Image.ParseImageID()
		if err != nil {
			return []string{}, err
		}
		imageShas = append(imageShas, sha)
	}
	return imageShas, nil
}

func ExpectedPodAnnotations(podName string, imageShas []string, dump *Dump) (map[string]string, error) {
	perceptor := dump.Perceptor
	annotations := map[string]string{}
	pod, ok := perceptor.PodsByName[podName]
	if !ok {
		// didn't find this pod in the scan results?  then there shouldn't be any BD annotations
		return annotations, nil
	}

	for i, sha := range imageShas {
		image, ok := perceptor.ImagesBySha[sha]
		if !ok {
			return nil, fmt.Errorf("unable to find image %s", sha)
		}
		annotations[kube.PodImageAnnotationKeyOverallStatus.String(i)] = image.OverallStatus
		annotations[kube.PodImageAnnotationKeyVulnerabilities.String(i)] = fmt.Sprintf("%d", image.Vulnerabilities)
		annotations[kube.PodImageAnnotationKeyPolicyViolations.String(i)] = fmt.Sprintf("%d", image.PolicyViolations)
		annotations[kube.PodImageAnnotationKeyProjectEndpoint.String(i)] = image.ComponentsURL
		annotations[kube.PodImageAnnotationKeyScannerVersion.String(i)] = dump.Hub.Version
		annotations[kube.PodImageAnnotationKeyServerVersion.String(i)] = dump.Hub.Version
		name, _, _ := dump.Kube.ImagesBySha[sha].ParseImageID() // just ignore errors and missing values!  maybe not a good idea TODO
		name = strings.Replace(name, "/", ".", -1)
		name = strings.Replace(name, ":", ".", -1)
		annotations[kube.PodImageAnnotationKeyImage.String(i)] = name
	}

	annotations[kube.PodAnnotationKeyOverallStatus.String()] = pod.OverallStatus
	annotations[kube.PodAnnotationKeyVulnerabilities.String()] = fmt.Sprintf("%d", pod.Vulnerabilities)
	annotations[kube.PodAnnotationKeyPolicyViolations.String()] = fmt.Sprintf("%d", pod.PolicyViolations)
	annotations[kube.PodAnnotationKeyScannerVersion.String()] = dump.Hub.Version
	annotations[kube.PodAnnotationKeyServerVersion.String()] = dump.Hub.Version

	return annotations, nil
}

// ShortenLabelContent will ensure the data is less than the 63 character limit and doesn't contain
// any characters that are not allowed
func ShortenLabelContent(data string) string {
	newData := RemoveRegistryInfo(data)

	// Label values can not be longer than 63 characters
	if len(newData) > 63 {
		newData = newData[0:63]
	}

	return newData
}

// RemoveRegistryInfo will take a string and return a string that removes any registry name information
// and replaces all / with .
func RemoveRegistryInfo(d string) string {
	s := strings.Split(d, "/")

	// If the data includes a . or : before the first / then that string is most likely
	// a registry name.  Remove it because it could make the data too long and
	// truncate useful information
	if strings.Contains(s[0], ".") || strings.Contains(s[0], ":") {
		s = s[1:]
	}
	return strings.Join(s, ".")
}

func ExpectedPodLabels(podName string, imageShas []string, dump *Dump) (map[string]string, error) {
	perceptor := dump.Perceptor
	labels := map[string]string{}
	pod, ok := perceptor.PodsByName[podName]
	if !ok {
		return labels, nil
	}

	for i, sha := range imageShas {
		image, ok := perceptor.ImagesBySha[sha]
		if !ok {
			return nil, fmt.Errorf("unable to find image %s", sha)
		}
		labels[kube.PodImageLabelKeyOverallStatus.String(i)] = image.OverallStatus
		labels[kube.PodImageLabelKeyVulnerabilities.String(i)] = fmt.Sprintf("%d", image.Vulnerabilities)
		labels[kube.PodImageLabelKeyPolicyViolations.String(i)] = fmt.Sprintf("%d", image.PolicyViolations)
		name, _, err := dump.Kube.ImagesBySha[sha].ParseImageID()
		// TODO ignoring errors ... not a great idea
		if err != nil {
			log.Errorf("unable to parse image id %s: %s", dump.Kube.ImagesBySha[sha].ImageID, err.Error())
		}
		labels[kube.PodImageLabelKeyImage.String(i)] = ShortenLabelContent(name)
	}

	labels[kube.PodLabelKeyOverallStatus.String()] = pod.OverallStatus
	labels[kube.PodLabelKeyVulnerabilities.String()] = fmt.Sprintf("%d", pod.Vulnerabilities)
	labels[kube.PodLabelKeyPolicyViolations.String()] = fmt.Sprintf("%d", pod.PolicyViolations)

	return labels, nil
}
