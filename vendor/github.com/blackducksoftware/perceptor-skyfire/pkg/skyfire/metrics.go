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

package skyfire

import (
	"github.com/prometheus/client_golang/prometheus"
)

var problemsGauge *prometheus.GaugeVec
var errorCounter *prometheus.CounterVec
var eventCounter *prometheus.CounterVec

func recordReportProblem(name string, count int) {
	problemsGauge.With(prometheus.Labels{"name": name}).Set(float64(count))
}

func recordError(name string) {
	errorCounter.With(prometheus.Labels{"name": name}).Inc()
}

func recordEvent(name string) {
	eventCounter.With(prometheus.Labels{"name": name}).Inc()
}

func init() {
	problemsGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "perceptor",
		Subsystem: "skyfire",
		Name:      "test_issues",
		Help:      "names and counts for issues discovered in perceptor testing",
	}, []string{"name"})
	prometheus.MustRegister(problemsGauge)

	errorCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "perceptor",
		Subsystem: "skyfire",
		Name:      "errors",
		Help:      "internal skyfire errors",
	}, []string{"name"})
	prometheus.MustRegister(errorCounter)

	eventCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "perceptor",
		Subsystem: "skyfire",
		Name:      "events",
		Help:      "internal skyfire events",
	}, []string{"name"})
	prometheus.MustRegister(eventCounter)
}
