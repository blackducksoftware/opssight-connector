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
	"github.com/juju/errors"
)

// ClientInterface .....
type ClientInterface interface {
	Dump() (*Dump, error)
}

// Client .....
type Client struct {
	Resty *resty.Client
	Host  string
	Port  int
}

// NewClient .....
func NewClient(host string, port int) *Client {
	restyClient := resty.New()
	restyClient.SetTimeout(time.Duration(5 * time.Second))
	return &Client{
		Resty: restyClient,
		Host:  host,
		Port:  port,
	}
}

// Dump .....
func (pd *Client) Dump() (*Dump, error) {
	scanResults, err := pd.DumpScanResults()
	if err != nil {
		return nil, errors.Annotate(err, "unable to dump scan results")
	}
	model, err := pd.DumpModel()
	if err != nil {
		return nil, errors.Annotate(err, "unable to dump model")
	}
	return NewDump(scanResults, model), nil
}

// DumpModel .....
func (pd *Client) DumpModel() (*api.Model, error) {
	url := fmt.Sprintf("http://%s/model", pd.Host)
	//url := fmt.Sprintf("http://%s:%d/model", pd.Host, pd.Port)
	resp, err := pd.Resty.R().SetResult(&api.Model{}).Get(url)
	if err != nil {
		return nil, errors.Trace(err)
	}
	switch result := resp.Result().(type) {
	case *api.Model:
		return result, nil
	default:
		return nil, fmt.Errorf("invalid response type: expected *api.Model, got %s (%+v)", reflect.TypeOf(result), result)
	}
}

// DumpScanResults .....
func (pd *Client) DumpScanResults() (*api.ScanResults, error) {
	url := fmt.Sprintf("http://%s/scanresults", pd.Host)
	//url := fmt.Sprintf("http://%s:%d/scanresults", pd.Host, pd.Port)
	resp, err := pd.Resty.R().SetResult(&api.ScanResults{}).Get(url)
	if err != nil {
		return nil, errors.Trace(err)
	}
	switch result := resp.Result().(type) {
	case *api.ScanResults:
		return result, nil
	default:
		return nil, fmt.Errorf("invalid response type: expected *api.ScanResults, got %s (%+v)", reflect.TypeOf(result), result)
	}
}
