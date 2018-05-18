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
	"k8s.io/api/core/v1"
)

func mapKubePod(kubePod *v1.Pod) *Pod {
	containers := []*Container{}
	for _, newCont := range kubePod.Status.ContainerStatuses {
		newImage := NewImage(newCont.Image, newCont.ImageID)
		addedCont := NewContainer(newImage, newCont.Name)
		containers = append(containers, addedCont)
	}
	return &Pod{
		kubePod.Name,
		string(kubePod.UID),
		kubePod.Namespace,
		containers,
		kubePod.Annotations,
		kubePod.Labels}
}
