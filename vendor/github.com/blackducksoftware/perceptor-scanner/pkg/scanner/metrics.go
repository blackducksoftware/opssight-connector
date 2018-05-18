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

package scanner

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var httpResults *prometheus.CounterVec
var scanClientDurationHistogram *prometheus.HistogramVec
var totalScannerDurationHistogram *prometheus.HistogramVec
var errorsCounter *prometheus.CounterVec

// helpers

func recordError(errorStage string, errorName string) {
	errorsCounter.With(prometheus.Labels{"stage": errorStage, "errorName": errorName}).Inc()
}

// recorders

func recordHTTPStats(path string, statusCode int) {
	httpResults.With(prometheus.Labels{"path": path, "code": fmt.Sprintf("%d", statusCode)}).Inc()
}

func recordScanClientDuration(duration time.Duration, isSuccess bool) {
	result := "success"
	if !isSuccess {
		result = "failure"
	}
	scanClientDurationHistogram.With(prometheus.Labels{"result": result}).Observe(duration.Seconds())
}

func recordTotalScannerDuration(duration time.Duration, isSuccess bool) {
	result := "success"
	if !isSuccess {
		result = "failure"
	}
	totalScannerDurationHistogram.With(prometheus.Labels{"result": result}).Observe(duration.Seconds())
}

func recordScannerError(errorName string) {
	recordError("scan client", errorName)
}

func recordCleanUpTarFile(isSuccess bool) {
	if !isSuccess {
		recordScannerError("clean up tar file")
	}
	// TODO should we have a metric for success?
}

// init

func init() {
	httpResults = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "perceptor",
		Subsystem: "scanner",
		Name:      "http_response_status_codes",
		Help:      "status codes for responses from HTTP requests issued by scanner",
	}, []string{"path", "code"})

	scanClientDurationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "perceptor",
		Subsystem: "scanner",
		Name:      "scan_client_duration",
		Help:      "time duration of running the java scan client",
		Buckets:   prometheus.ExponentialBuckets(0.25, 2, 20),
	}, []string{"result"})

	totalScannerDurationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "perceptor",
		Subsystem: "scanner",
		Name:      "scanner_total_duration",
		Help:      "total time duration of running the scanner",
		Buckets:   prometheus.ExponentialBuckets(0.25, 2, 20),
	}, []string{"result"})

	errorsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "perceptor",
		Subsystem: "scanner",
		Name:      "scannerErrors",
		Help:      "error codes from image scanning",
	}, []string{"stage", "errorName"})

	prometheus.MustRegister(errorsCounter)
	prometheus.MustRegister(scanClientDurationHistogram)
	prometheus.MustRegister(totalScannerDurationHistogram)
	prometheus.MustRegister(httpResults)
}
