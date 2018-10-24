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

package metrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var httpResults *prometheus.CounterVec
var durationsHistogram *prometheus.HistogramVec
var errorsCounter *prometheus.CounterVec
var imagesAnnotated *prometheus.CounterVec
var totalImagesAnnotated *prometheus.CounterVec
var podsAnnotated *prometheus.CounterVec
var totalPodsAnnotated *prometheus.CounterVec

// RecordError records metric information related to errors
func RecordError(errorStage string, errorName string) {
	errorsCounter.With(prometheus.Labels{"stage": errorStage, "errorName": errorName}).Inc()
}

// RecordDuration records the duration of an operation
func RecordDuration(operation string, duration time.Duration) {
	durationsHistogram.With(prometheus.Labels{"operation": operation}).Observe(duration.Seconds())
}

// RecordHTTPStats records metric information related to http requests
func RecordHTTPStats(path string, success bool) {
	httpResults.With(prometheus.Labels{"path": path, "result": fmt.Sprintf("%t", success)}).Inc()
}

// RecordImageAnnotation records metric information related to image annotations
func RecordImageAnnotation(annotator string, imageName string) {
	imagesAnnotated.With(prometheus.Labels{"annotator": annotator, "image_name": imageName}).Inc()
	totalImagesAnnotated.With(prometheus.Labels{"annotator": annotator, "images_annotated": "total"}).Inc()
}

// RecordPodAnnotation records metric information related to pod annotations
func RecordPodAnnotation(annotator string, podName string) {
	podsAnnotated.With(prometheus.Labels{"annotator": annotator, "pod_name": podName}).Inc()
	totalPodsAnnotated.With(prometheus.Labels{"annotator": annotator, "pods_annotated": "total"}).Inc()
}

func init() {
	httpResults = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "perceptor",
			Subsystem: "image_perceiver",
			Name:      "http_response_status_codes",
			Help:      "success/failure responses from HTTP requests issued by image perceiver",
		}, []string{"path", "result"})

	durationsHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "perceptor",
			Subsystem: "image_perceiver",
			Name:      "timings",
			Help:      "time durations of image perceiver operations",
			Buckets:   prometheus.ExponentialBuckets(0.0000001, 2, 30),
		}, []string{"operation"})

	errorsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "perceptor",
			Subsystem: "image_perceiver",
			Name:      "errors",
			Help:      "errors from image perceiver operations",
		}, []string{"stage", "errorName"})

	imagesAnnotated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "perceptor",
			Subsystem: "image_perceiver",
			Name:      "annotations",
			Help:      "individual image annotations",
		}, []string{"annotator", "image_name"})

	totalImagesAnnotated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "perceptor",
			Subsystem: "image_perceiver",
			Name:      "total_annotations",
			Help:      "total images annotated",
		}, []string{"annotator", "images_annotated"})

	podsAnnotated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "perceptor",
			Subsystem: "pod_perceiver",
			Name:      "annotations",
			Help:      "individual pod annotations",
		}, []string{"annotator", "pod_name"})

	totalPodsAnnotated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "perceptor",
			Subsystem: "pod_perceiver",
			Name:      "total_annotations",
			Help:      "total pods annotated",
		}, []string{"annotator", "pods_annotated"})

	prometheus.MustRegister(errorsCounter)
	prometheus.MustRegister(durationsHistogram)
	prometheus.MustRegister(httpResults)
	prometheus.MustRegister(imagesAnnotated)
	prometheus.MustRegister(totalImagesAnnotated)
	prometheus.MustRegister(podsAnnotated)
	prometheus.MustRegister(totalPodsAnnotated)
}
