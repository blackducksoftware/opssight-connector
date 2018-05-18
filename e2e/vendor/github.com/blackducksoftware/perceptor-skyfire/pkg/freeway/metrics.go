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

package freeway

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var linkTypeDurationHistogram *prometheus.HistogramVec
var errorCounter *prometheus.CounterVec

func recordLinkTypeDuration(linkType LinkType, duration time.Duration) {
	milliseconds := float64(duration / time.Millisecond)
	linkTypeDurationHistogram.With(prometheus.Labels{"linkType": linkType.String()}).Observe(milliseconds)
}

func recordError(name string) {
	errorCounter.With(prometheus.Labels{"name": name}).Inc()
}

func init() {
	linkTypeDurationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "perceptor",
		Subsystem: "skyfire",
		Name:      "hub_api_link_duration",
		Help:      "durations for hub API calls in milliseconds, grouped by link type",
		Buckets:   prometheus.ExponentialBuckets(1, 2, 20),
	}, []string{"linkType"})
	prometheus.MustRegister(linkTypeDurationHistogram)

	errorCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "perceptor",
		Subsystem: "skyfire",
		Name:      "hub_api_link_duration_errors",
		Help:      "errors encountered when scraping the Hub API",
	}, []string{"name"})
	prometheus.MustRegister(errorCounter)
}
