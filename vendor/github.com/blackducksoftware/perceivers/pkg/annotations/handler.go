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

package annotations

// MapCompareHandler handles comparing 2 maps
type MapCompareHandler interface {
	CompareMaps(map[string]string, map[string]string) bool
}

// MapCompareHandlerFuncs is an adapter to let you easily define
// a function to use to compare 2 maps if desired while still
// implementing MapCompareHandler
type MapCompareHandlerFuncs struct {
	MapCompareFunc func(map[string]string, map[string]string) bool
}

// CompareMaps calls MapCompareFunc if it is not null
func (m MapCompareHandlerFuncs) CompareMaps(bigMap map[string]string, subset map[string]string) bool {
	if m.MapCompareFunc != nil {
		return m.MapCompareFunc(bigMap, subset)
	}
	return true
}

// ImageAnnotatorHandler provides the functions needed to annotate images
type ImageAnnotatorHandler interface {
	MapCompareHandler
	CreateImageLabels(interface{}, string, int) map[string]string
	CreateImageAnnotations(interface{}, string, int) map[string]string
}

// ImageAnnotatorHandlerFuncs is an adapter to let you easily define
// as many of the image annotation functions as desired while still implementing
// ImageAnnotatorHandler
type ImageAnnotatorHandlerFuncs struct {
	MapCompareHandlerFuncs
	ImageLabelCreationFunc      func(interface{}, string, int) map[string]string
	ImageAnnotationCreationFunc func(interface{}, string, int) map[string]string
}

// CreateImageLabels calls LabelCreationFunc if it is not null
func (i ImageAnnotatorHandlerFuncs) CreateImageLabels(data interface{}, name string, count int) map[string]string {
	if i.ImageLabelCreationFunc != nil {
		return i.ImageLabelCreationFunc(data, name, count)
	}
	return make(map[string]string)
}

// CreateImageAnnotations calls AnnotationCreationFunc if it is not null
func (i ImageAnnotatorHandlerFuncs) CreateImageAnnotations(data interface{}, name string, count int) map[string]string {
	if i.ImageLabelCreationFunc != nil {
		return i.ImageAnnotationCreationFunc(data, name, count)
	}
	return make(map[string]string)
}

// PodAnnotatorHandler provides the functions needed to annotate pods
type PodAnnotatorHandler interface {
	ImageAnnotatorHandler
	CreatePodLabels(interface{}) map[string]string
	CreatePodAnnotations(interface{}) map[string]string
}

// PodAnnotatorHandlerFuncs is an adapter to let you easily define
// as many of the pod annotation functions as desired while still implementing
// PodAnnotatorHandler
type PodAnnotatorHandlerFuncs struct {
	ImageAnnotatorHandlerFuncs
	PodLabelCreationFunc      func(interface{}) map[string]string
	PodAnnotationCreationFunc func(interface{}) map[string]string
}

// CreatePodLabels calls LabelCreationFunc if it is not null
func (p PodAnnotatorHandlerFuncs) CreatePodLabels(data interface{}) map[string]string {
	if p.PodLabelCreationFunc != nil {
		return p.PodLabelCreationFunc(data)
	}
	return make(map[string]string)
}

// CreatePodAnnotations calls AnnotationCreationFunc if it is not null
func (p PodAnnotatorHandlerFuncs) CreatePodAnnotations(data interface{}) map[string]string {
	if p.PodLabelCreationFunc != nil {
		return p.PodAnnotationCreationFunc(data)
	}
	return make(map[string]string)
}
