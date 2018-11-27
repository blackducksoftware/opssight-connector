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

	"github.com/blackducksoftware/perceptor-skyfire/pkg/kube"
)

// KubeReport .....
type KubeReport struct {
	UnanalyzeablePods      []string
	UnparseableImages      []string
	PartiallyAnnotatedPods []string
	PartiallyLabeledPods   []string
}

// NewKubeReport .....
func NewKubeReport(dump *kube.Dump) *KubeReport {
	partiallyAnnotatedKubePods, partiallyLabeledKubePods := PartiallyHandledKubePods(dump)
	return &KubeReport{
		UnanalyzeablePods:      UnanalyzeablePods(dump),
		UnparseableImages:      UnparseableKubeImages(dump),
		PartiallyAnnotatedPods: partiallyAnnotatedKubePods,
		PartiallyLabeledPods:   partiallyLabeledKubePods,
	}
}

// HumanReadableString .....
func (k *KubeReport) HumanReadableString() string {
	return fmt.Sprintf(`
Kubernetes:
 - %d unanalyzeable pod(s)
 - %d unparseable ImageID(s)
 - %d partially annotated pod(s)
 - %d partially labeled pod(s)
`,
		len(k.UnanalyzeablePods),
		len(k.UnparseableImages),
		len(k.PartiallyAnnotatedPods),
		len(k.PartiallyLabeledPods))
}

// PartiallyHandledKubePods .....
func PartiallyHandledKubePods(dump *kube.Dump) (partiallyAnnotatedKubePods []string, partiallyLabeledKubePods []string) {
	partiallyAnnotatedKubePods = []string{}
	partiallyLabeledKubePods = []string{}
	for podName, pod := range dump.PodsByName {
		if pod.HasAnyBDAnnotations() && !pod.HasAllBDAnnotations() {
			partiallyAnnotatedKubePods = append(partiallyAnnotatedKubePods, podName)
		}

		if pod.HasAnyBDLabels() && !pod.HasAllBDLabels() {
			partiallyLabeledKubePods = append(partiallyLabeledKubePods, podName)
		}
	}
	return
}

// UnparseableKubeImages .....
func UnparseableKubeImages(dump *kube.Dump) []string {
	images := []string{}
	for _, image := range dump.ImagesMissingSha {
		images = append(images, image.ImageID)
	}
	return images
}

// UnanalyzeablePods .....
func UnanalyzeablePods(dump *kube.Dump) []string {
	unanalyzeablePods := []string{}
	for podName, pod := range dump.PodsByName {
		_, err := PodShas(pod)
		if err != nil {
			unanalyzeablePods = append(unanalyzeablePods, podName)
		}
	}
	return unanalyzeablePods
}
