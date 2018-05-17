/*
Copyright (C) 2018 Black Duck Software, Inc.

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

package perceptor

import (
	"fmt"
	"reflect"
	"time"

	api "github.com/blackducksoftware/perceptor/pkg/api"
	resty "github.com/go-resty/resty"
)

type PerceptorDumper struct {
	Resty *resty.Client
	Host  string
	Port  int
}

func NewPerceptorDumper(host string, port int) *PerceptorDumper {
	restyClient := resty.New()
	restyClient.SetTimeout(time.Duration(5 * time.Second))
	return &PerceptorDumper{
		Resty: restyClient,
		Host:  host,
		Port:  port,
	}
}

func (pd *PerceptorDumper) Dump() (*Dump, error) {
	scanResults, err := pd.DumpScanResults()
	if err != nil {
		return nil, err
	}
	model, err := pd.DumpModel()
	if err != nil {
		return nil, err
	}
	return NewDump(scanResults, model), nil
}

func (pd *PerceptorDumper) DumpModel() (*api.Model, error) {
	url := fmt.Sprintf("http://%s:%d/model", pd.Host, pd.Port)
	resp, err := pd.Resty.R().SetResult(&api.Model{}).Get(url)
	if err != nil {
		return nil, err
	}
	switch result := resp.Result().(type) {
	case *api.Model:
		return result, nil
	default:
		return nil, fmt.Errorf("invalid response type: expected *api.Model, got %s (%+v)", reflect.TypeOf(result), result)
	}
}

func (pd *PerceptorDumper) DumpScanResults() (*api.ScanResults, error) {
	url := fmt.Sprintf("http://%s:%d/scanresults", pd.Host, pd.Port)
	resp, err := pd.Resty.R().SetResult(&api.ScanResults{}).Get(url)
	if err != nil {
		return nil, err
	}
	switch result := resp.Result().(type) {
	case *api.ScanResults:
		return result, nil
	default:
		return nil, fmt.Errorf("invalid response type: expected *api.ScanResults, got %s (%+v)", reflect.TypeOf(result), result)
	}
}
