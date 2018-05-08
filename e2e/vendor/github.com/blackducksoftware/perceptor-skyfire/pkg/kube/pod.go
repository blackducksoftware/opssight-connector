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
	"fmt"
)

type Pod struct {
	Name        string
	UID         string
	Namespace   string
	Containers  []*Container
	Annotations map[string]string
	Labels      map[string]string
}

func (pod *Pod) QualifiedName() string {
	return fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
}

func (pod *Pod) hasImage(image *Image) bool {
	for _, cont := range pod.Containers {
		if cont.Image == image {
			return true
		}
	}
	return false
}

func NewPod(name string, uid string, namespace string, containers []*Container) *Pod {
	return &Pod{
		Name:       name,
		UID:        uid,
		Namespace:  namespace,
		Containers: containers,
	}
}

// BD annotations

func (pod *Pod) BDAnnotations() map[string]string {
	bdKeys := BDPodAnnotationKeys(len(pod.Containers))
	dict := map[string]string{}
	for _, key := range bdKeys {
		val, ok := pod.Annotations[key]
		if ok {
			dict[key] = val
		}
	}
	return dict
}

func (pod *Pod) HasAllBDAnnotations() bool {
	return len(pod.BDAnnotations()) == len(BDPodAnnotationKeys(len(pod.Containers)))
}

func (pod *Pod) HasAnyBDAnnotations() bool {
	return len(pod.BDAnnotations()) > 0
}

func BDPodAnnotationKeys(containerCount int) []string {
	keys := append([]string{}, podAnnotationKeyStrings...)
	for i := 0; i < containerCount; i++ {
		keys = append(keys, podImageAnnotationKeyStrings(i)...)
	}
	return keys
}

func RemoveBDPodAnnotationKeys(containerCount int, annotations map[string]string) map[string]string {
	dict := CopyMap(annotations)
	return RemoveKeys(dict, BDPodAnnotationKeys(containerCount))
}

// BD labels

func (pod *Pod) BDLabels() map[string]string {
	bdKeys := BDPodLabelKeys(len(pod.Containers))
	dict := map[string]string{}
	for _, key := range bdKeys {
		val, ok := pod.Labels[key]
		if ok {
			dict[key] = val
		}
	}
	return dict
}

func (pod *Pod) HasAllBDLabels() bool {
	return len(pod.BDLabels()) == len(BDPodLabelKeys(len(pod.Containers)))
}

func (pod *Pod) HasAnyBDLabels() bool {
	return len(pod.BDLabels()) > 0
}

func BDPodLabelKeys(containerCount int) []string {
	keys := append([]string{}, podLabelKeyStrings...)
	for i := 0; i < containerCount; i++ {
		keys = append(keys, podImageLabelKeyStrings(i)...)
	}
	return keys
}

func RemoveBDPodLabelKeys(containerCount int, labels map[string]string) map[string]string {
	dict := CopyMap(labels)
	return RemoveKeys(dict, BDPodLabelKeys(containerCount))
}
