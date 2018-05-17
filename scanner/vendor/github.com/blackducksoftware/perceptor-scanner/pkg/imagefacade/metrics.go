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

package imagefacade

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var httpRequestsCounter *prometheus.CounterVec
var actionsCounter *prometheus.CounterVec
var reducerActivityCounter *prometheus.CounterVec
var diskMetricsGauge *prometheus.GaugeVec
var imagePullResultCounter *prometheus.CounterVec

func recordHTTPRequest(path string) {
	httpRequestsCounter.With(prometheus.Labels{"path": path}).Inc()
}

func recordActionType(action string) {
	actionsCounter.With(prometheus.Labels{"action": action}).Inc()
}

func recordReducerActivity(isActive bool, duration time.Duration) {
	state := "idle"
	if isActive {
		state = "active"
	}
	reducerActivityCounter.With(prometheus.Labels{"state": state}).Add(duration.Seconds())
}

func megabytes(number uint64) float64 {
	bytesPerMB := uint64(1024 * 1024)
	return float64(number / bytesPerMB)
}

func recordDiskMetrics(diskMetrics *DiskMetrics) {
	diskMetricsGauge.With(prometheus.Labels{"name": "available_MBs"}).Set(megabytes(diskMetrics.AvailableBytes))
	diskMetricsGauge.With(prometheus.Labels{"name": "free_MBs"}).Set(megabytes(diskMetrics.FreeBytes))
	diskMetricsGauge.With(prometheus.Labels{"name": "total_MBs"}).Set(megabytes(diskMetrics.TotalBytes))
	diskMetricsGauge.With(prometheus.Labels{"name": "used_MBs"}).Set(megabytes(diskMetrics.UsedBytes))
}

func recordImagePullResult(success bool) {
	successString := fmt.Sprintf("%t", success)
	imagePullResultCounter.With(prometheus.Labels{"success": successString}).Inc()
}

func init() {
	httpRequestsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "perceptor",
		Subsystem: "imagefacade",
		Name:      "http_requests_received",
		Help:      "HTTP requests received by imagefacade",
	},
		[]string{"path"})
	prometheus.MustRegister(httpRequestsCounter)

	actionsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "perceptor",
		Subsystem: "imagefacade",
		Name:      "actions",
		Help:      "actions processed by imagefacade and applied to the model",
	},
		[]string{"action"})
	prometheus.MustRegister(actionsCounter)

	reducerActivityCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "perceptor",
		Subsystem: "imagefacade",
		Name:      "reducer_activity",
		Help:      "activity of the reducer -- how much time it's been idle and active, in seconds",
	}, []string{"state"})
	prometheus.MustRegister(reducerActivityCounter)

	diskMetricsGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "perceptor",
		Subsystem: "imagefacade",
		Name:      "disk_metrics",
		Help:      "usage statistics for disk",
	}, []string{"name"})
	prometheus.MustRegister(diskMetricsGauge)

	imagePullResultCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "perceptor",
		Subsystem: "imagefacade",
		Name:      "image_pull_result",
		Help:      "whether image pull/get succeeded or failed",
	}, []string{"success"})
	prometheus.MustRegister(imagePullResultCounter)
}
