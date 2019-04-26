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
	"strings"

	"github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/util"

	"github.com/koki/short/converter/converters"
	"github.com/koki/short/types"

	"k8s.io/apimachinery/pkg/runtime"
)

// Namespace defines the namespace component
type Namespace struct {
	obj *types.Namespace
}

// NewNamespace creates a Namespace object
func NewNamespace(config api.NamespaceConfig) *Namespace {
	version := "rbac.authorization.k8s.io/v1"
	if len(config.APIVersion) > 0 {
		version = config.APIVersion
	}
	n := &types.Namespace{
		Version:   version,
		Name:      config.Name,
		Cluster:   config.ClusterName,
		Namespace: config.Namespace,
	}

	return &Namespace{obj: n}
}

// GetObj will return the namespace object in a format the deployer can use
func (n *Namespace) GetObj() *types.Namespace {
	return n.obj
}

// GetName will return the name of namespace
func (n *Namespace) GetName() string {
	return n.obj.Name
}

// AddAnnotations adds annotations to the namespace
func (n *Namespace) AddAnnotations(new map[string]string) {
	n.obj.Annotations = util.MapMerge(n.obj.Annotations, new)
}

// RemoveAnnotations removes annotations from the namespace
func (n *Namespace) RemoveAnnotations(remove []string) {
	for _, k := range remove {
		n.obj.Annotations = util.RemoveElement(n.obj.Annotations, k)
	}
}

// AddLabels adds labels to the namespace
func (n *Namespace) AddLabels(new map[string]string) {
	n.obj.Labels = util.MapMerge(n.obj.Labels, new)
}

// RemoveLabels removes labels from the namespace
func (n *Namespace) RemoveLabels(remove []string) {
	for _, k := range remove {
		n.obj.Labels = util.RemoveElement(n.obj.Labels, k)
	}
}

// AddFinalizers will add finalizers to the namespace
func (n *Namespace) AddFinalizers(new []string) {
	for _, f := range new {
		n.obj.Finalizers = n.appendFinalizerIfMissing(f, n.obj.Finalizers)
	}
}

// RemoveFinalizer will remove a finalizer from the namespace
func (n *Namespace) RemoveFinalizer(remove string) {
	for l, f := range n.obj.Finalizers {
		if strings.Compare(string(f), remove) == 0 {
			n.obj.Finalizers = append(n.obj.Finalizers[:l], n.obj.Finalizers[l+1:]...)
		}
	}
}

func (n *Namespace) appendFinalizerIfMissing(new string, list []types.FinalizerName) []types.FinalizerName {
	for _, f := range list {
		if strings.Compare(new, string(f)) == 0 {
			return list
		}
	}
	return append(list, types.FinalizerName(new))
}

// ToKube returns the kubernetes version of the namespace
func (n *Namespace) ToKube() (runtime.Object, error) {
	wrapper := &types.NamespaceWrapper{Namespace: *n.obj}
	return converters.Convert_Koki_Namespace_to_Kube_Namespace(wrapper)
}
