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

package main

import (
	"encoding/json"
	"fmt"
	"os"

	kube "github.com/blackducksoftware/perceptor-skyfire/pkg/kube"
	report "github.com/blackducksoftware/perceptor-skyfire/pkg/report"
)

func main() {
	kubeConfigPath := os.Args[1]
	masterURL := os.Args[2]

	config := &kube.KubeClientConfig{KubeConfigPath: kubeConfigPath, MasterURL: masterURL}
	kubeDumper, err := kube.NewKubeClient(config)
	if err != nil {
		panic(err)
	}

	kubeDump, err := kubeDumper.Dump()
	if err != nil {
		panic(err)
	}
	kubeReport := report.NewKubeReport(kubeDump)

	dict := map[string]interface{}{
		"Dump":   kubeDump,
		"Report": kubeReport,
	}

	bytes, err := json.MarshalIndent(dict, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(bytes))

	// TODO
}
