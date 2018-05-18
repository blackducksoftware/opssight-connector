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

package docker

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var tarballSize *prometheus.HistogramVec
var dockerCreateDurationHistogram prometheus.Histogram
var dockerGetDurationHistogram prometheus.Histogram
var dockerTotalDurationHistogram prometheus.Histogram
var errorsCounter *prometheus.CounterVec
var eventsCounter *prometheus.CounterVec

// durations

func recordDockerCreateDuration(duration time.Duration) {
	dockerCreateDurationHistogram.Observe(duration.Seconds())
}

func recordDockerGetDuration(duration time.Duration) {
	dockerGetDurationHistogram.Observe(duration.Seconds())
}

func recordDockerTotalDuration(duration time.Duration) {
	dockerTotalDurationHistogram.Observe(duration.Seconds())
}

func recordEvent(event string) {
	eventsCounter.With(prometheus.Labels{"event": event}).Inc()
}

// tar file size and docker errors

func recordTarFileSize(fileSizeMBs int) {
	tarballSize.WithLabelValues("tarballSize").Observe(float64(fileSizeMBs))
}

func recordDockerError(errorStage string, errorName string, image Image, err error) {
	// TODO what use can be made of `image` and `err`?
	// we might want to group the errors by image sha or something
	errorsCounter.With(prometheus.Labels{"stage": errorStage, "errorName": errorName}).Inc()
}

// init

func init() {
	tarballSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "perceptor",
			Subsystem: "imagefacade",
			Name:      "tarballsize",
			Help:      "tarball file size in MBs",
			Buckets:   prometheus.ExponentialBuckets(1, 2, 15),
		},
		[]string{"tarballSize"})

	dockerCreateDurationHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "perceptor",
		Subsystem: "imagefacade",
		Name:      "docker_create_duration",
		Help:      "time durations of docker create operations",
		Buckets:   prometheus.ExponentialBuckets(0.25, 2, 20),
	})

	dockerGetDurationHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "perceptor",
		Subsystem: "imagefacade",
		Name:      "docker_get_duration",
		Help:      "time durations of docker get operations",
		Buckets:   prometheus.ExponentialBuckets(0.25, 2, 20),
	})

	dockerTotalDurationHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "perceptor",
		Subsystem: "imagefacade",
		Name:      "docker_total_duration",
		Help:      "time durations of docker total operations",
		Buckets:   prometheus.ExponentialBuckets(0.25, 2, 20),
	})

	errorsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "perceptor",
		Subsystem: "imagefacade",
		Name:      "dockerErrors",
		Help:      "error codes from image pulling from docker",
	}, []string{"stage", "errorName"})

	eventsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "perceptor",
		Subsystem: "imagefacade",
		Name:      "events",
		Help:      "miscellaneous events from imagefacade",
	}, []string{"event"})

	prometheus.MustRegister(errorsCounter)
	prometheus.MustRegister(dockerGetDurationHistogram)
	prometheus.MustRegister(dockerCreateDurationHistogram)
	prometheus.MustRegister(dockerTotalDurationHistogram)
	prometheus.MustRegister(tarballSize)
	prometheus.MustRegister(eventsCounter)
}
