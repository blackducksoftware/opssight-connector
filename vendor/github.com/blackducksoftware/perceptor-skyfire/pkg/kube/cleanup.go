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

package kube

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (client *KubeClient) CleanupAllPods() error {
	pods := client.clientset.CoreV1().Pods(v1.NamespaceAll)
	podList, err := pods.List(meta_v1.ListOptions{})
	if err != nil {
		log.Errorf("unable to list pods: %s", err.Error())
		return err
	}
	for _, pod := range podList.Items {
		log.Debugf("annotations before:\n%+v", pod.Annotations)
		log.Debugf("labels before:\n%+v\n", pod.Labels)
		updatedAnnotations := RemoveBDPodAnnotationKeys(len(pod.Status.ContainerStatuses), pod.Annotations)
		updatedLabels := RemoveBDPodLabelKeys(len(pod.Status.ContainerStatuses), pod.Labels)
		log.Debugf("annotations after:\n%+v", updatedAnnotations)
		log.Debugf("labels after:\n%+v\n\n", updatedLabels)
		pod.SetAnnotations(updatedAnnotations)
		pod.SetLabels(updatedLabels)
		nsPods := client.clientset.CoreV1().Pods(pod.Namespace)
		_, err := nsPods.Update(&pod)
		if err != nil {
			log.Errorf("unable to update pod %+v: %s", pod, err.Error())
			return err
		}
	}
	return nil
}
