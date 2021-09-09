/*
Copyright (C) 2019 Synopsys, Inc.

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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/imdario/mergo"
)

func generateObjectMeta(name string, ns string, cn string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:        name,
		Namespace:   ns,
		ClusterName: cn,
		Annotations: make(map[string]string),
		Labels:      make(map[string]string),
		Finalizers:  make([]string, 0),
	}
}

// MetadataFuncs defines some common metadata handling functions
type MetadataFuncs struct {
	obj metav1.Object
}

// AddAnnotations adds annotations to the object
func (m *MetadataFuncs) AddAnnotations(new map[string]string) {
	a := m.obj.GetAnnotations()
	mergo.Merge(&a, new, mergo.WithOverride)
	m.obj.SetAnnotations(a)
}

// RemoveAnnotations removes annotations from the object
func (m *MetadataFuncs) RemoveAnnotations(remove []string) {
	data := m.obj.GetAnnotations()
	for _, k := range remove {
		delete(data, k)
	}
	m.obj.SetAnnotations(data)
}

// AddLabels adds labels to the object
func (m *MetadataFuncs) AddLabels(new map[string]string) {
	l := m.obj.GetLabels()
	mergo.Merge(&l, new, mergo.WithOverride)
	m.obj.SetLabels(l)
}

// RemoveLabels removes labels from the object
func (m *MetadataFuncs) RemoveLabels(remove []string) {
	data := m.obj.GetLabels()
	for _, k := range remove {
		delete(data, k)
	}
	m.obj.SetLabels(data)
}

// AddFinalizers will add finalizers to the object
func (m *MetadataFuncs) AddFinalizers(new []string) {
	finalizers := m.obj.GetFinalizers()
	for _, f := range new {
		finalizers = appendStringIfMissing(f, finalizers)
	}
	m.obj.SetFinalizers(finalizers)
}

// RemoveFinalizer will remove a finalizer from the namespace
func (m *MetadataFuncs) RemoveFinalizer(remove string) {
	finalizers := m.obj.GetFinalizers()

	for l, f := range finalizers {
		if strings.Compare(string(f), remove) == 0 {
			finalizers = append(finalizers[:l], finalizers[l+1:]...)
			break
		}
	}

	m.obj.SetFinalizers(finalizers)
}
