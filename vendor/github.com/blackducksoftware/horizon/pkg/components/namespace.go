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

package components

import (
	"github.com/blackducksoftware/horizon/pkg/api"

	"k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Namespace defines the namespace component
type Namespace struct {
	*v1.Namespace
	MetadataFuncs
}

// NewNamespace creates a Namespace object
func NewNamespace(config api.NamespaceConfig) *Namespace {
	version := "v1"
	if len(config.APIVersion) > 0 {
		version = config.APIVersion
	}

	n := v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: version,
		},
		ObjectMeta: generateObjectMeta(config.Name, config.Namespace, config.ClusterName),
	}

	return &Namespace{&n, MetadataFuncs{&n}}
}

// Deploy will deploy the namespace to the cluster
func (n *Namespace) Deploy(res api.DeployerResources) error {
	_, err := res.KubeClient.CoreV1().Namespaces().Create(n.Namespace)
	return err
}

// Undeploy will remove the namespace from the cluster
func (n *Namespace) Undeploy(res api.DeployerResources) error {
	return res.KubeClient.CoreV1().Namespaces().Delete(n.Name, &metav1.DeleteOptions{})
}
