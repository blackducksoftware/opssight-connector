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

package mapper

import (
	"fmt"

	"github.com/blackducksoftware/perceivers/pkg/docker"

	perceptorapi "github.com/blackducksoftware/perceptor/pkg/api"

	"k8s.io/api/core/v1"

	metrics "github.com/blackducksoftware/perceivers/pod/pkg/metrics"
)

// NewPerceptorPodFromKubePod will convert a kubernetes pod object to a
// perceptor pod object
func NewPerceptorPodFromKubePod(kubePod *v1.Pod) (*perceptorapi.Pod, error) {
	containers := []perceptorapi.Container{}
	actual := len(kubePod.Status.ContainerStatuses)
	expected := len(kubePod.Spec.Containers)
	
	// Note that even this is not a permanant solution to race conditions between
	// unprocessed apiserver pod objects https://github.com/blackducksoftware/perceivers/issues/54
	// Revise in the 1.0.1 timeline.
	if actual != expected {
		return nil, fmt.Errorf("unable to instantiate perceptor pod: kube pod %s/%s has %d container statuses, but %d containers in its spec", kubePod.Namespace, kubePod.Name, actual, expected)
	}
	for _, newCont := range kubePod.Status.ContainerStatuses {
		if len(newCont.ImageID) > 0 {
			name, sha, err := docker.ParseImageIDString(newCont.ImageID)
			if err != nil {
				metrics.RecordError("pod_mapper", "unable to parse kubernetes imageID")
				return nil, fmt.Errorf("unable to parse kubernetes imageID string %s from pod %s/%s: %v", newCont.ImageID, kubePod.Namespace, kubePod.Name, err)
			}
			addedCont := perceptorapi.NewContainer(*perceptorapi.NewImage(name, sha, newCont.Image), newCont.Name)
			containers = append(containers, *addedCont)
		} else {
			metrics.RecordError("pod_mapper", "empty kubernetes imageID")
			return nil, fmt.Errorf("empty kubernetes imageID from pod %s/%s, container %s", kubePod.Namespace, kubePod.Name, newCont.Name)
		}
	}
	return perceptorapi.NewPod(kubePod.Name, string(kubePod.UID), kubePod.Namespace, containers), nil
}
