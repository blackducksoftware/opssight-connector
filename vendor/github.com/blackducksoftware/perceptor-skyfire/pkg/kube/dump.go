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

// Dump .....
type Dump struct {
	Meta              *Meta
	Pods              []*Pod
	PodsByName        map[string]*Pod
	DuplicatePodNames map[string]bool
	// Images     []*Image
	ImagesBySha        map[string]*Image
	DuplicateImageShas map[string]bool
	ImagesMissingSha   []*Image
}

// ServiceDump .....
type ServiceDump struct {
	Meta     *Meta
	Services []*Service
}

// NewServiceDump .....
func NewServiceDump(meta *Meta, services []*Service) *ServiceDump {
	dump := &ServiceDump{
		Meta:     meta,
		Services: services,
	}
	return dump
}

// NewDump .....
func NewDump(meta *Meta, pods []*Pod) *Dump {
	dump := &Dump{
		Meta:               meta,
		Pods:               pods,
		PodsByName:         map[string]*Pod{},
		DuplicatePodNames:  map[string]bool{},
		ImagesBySha:        map[string]*Image{},
		DuplicateImageShas: map[string]bool{},
		ImagesMissingSha:   []*Image{}}
	dump.computeDerivedData()
	return dump
}

func (kd *Dump) computeDerivedData() {
	for _, pod := range kd.Pods {
		_, ok := kd.PodsByName[pod.QualifiedName()]
		if ok {
			kd.DuplicatePodNames[pod.QualifiedName()] = true
		} else {
			kd.PodsByName[pod.QualifiedName()] = pod
		}
		for _, container := range pod.Containers {
			_, sha, err := container.Image.ParseImageID()
			if err != nil {
				kd.ImagesMissingSha = append(kd.ImagesMissingSha, container.Image)
				// log.Errorf("unable to parse sha for pod %s, container %s, image %s: %s", pod.QualifiedName(), container.Name, container.Image.ImageID, err.Error())
			} else {
				_, ok := kd.ImagesBySha[sha]
				if ok {
					kd.DuplicateImageShas[sha] = true
				} else {
					kd.ImagesBySha[sha] = container.Image
				}
			}
		}
	}
}
