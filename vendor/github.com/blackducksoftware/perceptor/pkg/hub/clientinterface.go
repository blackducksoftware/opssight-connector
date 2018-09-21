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

package hub

import (
	"time"

	"github.com/blackducksoftware/perceptor/pkg/api"
)

// ClientInterface .....
type ClientInterface interface {
	// commands coming in
	StartScanClient(scanName string)
	FinishScanClient(scanName string, scanErr error)
	SetTimeout(timeout time.Duration)
	ResetCircuitBreaker()
	// read-only queries
	Host() string
	Version() (string, error)
	// read-only, async queries (the channel produces a single event)
	Model() <-chan *api.ModelHub
	HasFetchedScans() <-chan bool
	ScansCount() <-chan int
	InProgressScans() <-chan []string
	ScanResults() <-chan map[string]*Scan
	Updates() <-chan Update
	// prelude to clean-up
	Stop()
	StopCh() <-chan struct{}
}
