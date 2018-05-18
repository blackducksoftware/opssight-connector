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

package kubebuilder

import (
	"encoding/json"

	// v1beta1 "k8s.io/api/extensions/v1beta1"

	log "github.com/sirupsen/logrus"

	"k8s.io/client-go/kubernetes"
)

func PrettyPrint(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "  ")
	println(string(b))
}

type Builder struct {
	Namespace string
	Resources Resources
	Clientset *kubernetes.Clientset
}

func NewBuilder(namespace string, resources Resources, clientset *kubernetes.Clientset) *Builder {
	return &Builder{
		Namespace: namespace,
		Resources: resources,
		Clientset: clientset,
	}
}

func (b *Builder) CreateResources() {
	clientset := b.Clientset
	resources := b.Resources
	namespace := b.Namespace

	for _, configMap := range resources.GetConfigMaps() {
		PrettyPrint(configMap)
		_, err := clientset.Core().ConfigMaps(namespace).Create(configMap)
		if err != nil {
			log.Errorf("unable to create configmap %+v", configMap)
			panic(err)
		}
	}
	for _, secret := range resources.GetSecrets() {
		PrettyPrint(secret)
		_, err := clientset.Core().Secrets(namespace).Create(secret)
		if err != nil {
			panic(err)
		}
	}
	for _, service := range resources.GetServices() {
		PrettyPrint(service)
		_, err := clientset.Core().Services(namespace).Create(service)
		if err != nil {
			panic(err)
		}
	}
	for _, rc := range resources.GetReplicationControllers() {
		PrettyPrint(rc)
		_, err := clientset.Core().ReplicationControllers(namespace).Create(rc)
		if err != nil {
			panic(err)
		}
	}

	// for _, dep := range deployments {
	// 	PrettyPrint(dep)
	// 	_, err := clientset.Extensions().Deployments(namespace).Create(dep)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }
}
