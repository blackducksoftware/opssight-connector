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

package basicskyfire

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/blackducksoftware/perceptor-protoform/pkg/api"
	"github.com/blackducksoftware/perceptor-protoform/pkg/protoform"
	skyfire "github.com/blackducksoftware/perceptor-skyfire/pkg/report"
)

type Pod struct {
	Name      string
	ImageName string
	Port      int32
}

var pods []*Pod

func FetchSkyfireReport(skyfireURL string) (*skyfire.Report, error) {
	httpClient := http.Client{Timeout: 5 * time.Second}
	resp, err := httpClient.Get(skyfireURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("invalid status code %d, expected 200", resp.StatusCode)
	}

	var report *skyfire.Report
	err = json.Unmarshal(bodyBytes, &report)
	if err != nil {
		return nil, err
	}

	return report, nil
}

func createDefaults() *api.ProtoformDefaults {
	d := protoform.NewDefaultsObj()
	d.Namespace = "perceptor"
	d.DefaultCPU = "100m"
	d.DefaultMem = "100Mi"
	return d
}

func addPods(name string, imageName string, port int32) {
	pod := &Pod{Name: name, ImageName: imageName, Port: port}
	pods = append(pods, pod)
}

func createPods() {
	os.Setenv("PCP_HUBUSERPASSWORD", "example")
	defaults := createDefaults()
	i := protoform.NewInstaller(defaults, "protoform.json")

	fmt.Printf("Default CPU is %s \n", defaults.DefaultCPU)
	fmt.Printf("Default Memory is %s \n", defaults.DefaultMem)

	defaultCPU, err := i.GenerateDefaultCPU(defaults.DefaultCPU)
	if err != nil {
		fmt.Errorf("Generating default CPU failed for %s due to %+v", defaults.DefaultCPU, err.Error())
	}

	defaultMemory, err := i.GenerateDefaultMemory(defaults.DefaultMem)
	if err != nil {
		fmt.Errorf("Generating default memory failed for %s due to %+v", defaults.DefaultMem, err.Error())
	}

	for _, pod := range pods {
		i.AddPod([]*protoform.ReplicationController{
			{
				Name:   pod.Name,
				Image:  pod.ImageName,
				CPU:    defaultCPU,
				Memory: defaultMemory,
				Port:   pod.Port,
			},
		})
	}

	i.RunUpstream()
}
