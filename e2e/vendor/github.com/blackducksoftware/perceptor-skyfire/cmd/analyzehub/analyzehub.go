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

	hub "github.com/blackducksoftware/perceptor-skyfire/pkg/hub"
	report "github.com/blackducksoftware/perceptor-skyfire/pkg/report"
	log "github.com/sirupsen/logrus"
)

func main() {
	url := os.Args[1]
	username := os.Args[2]
	password := os.Args[3]

	log.SetLevel(log.DebugLevel)

	hubDumper, err := hub.NewHubDumper(url, username, password)
	if err != nil {
		panic(err)
	}

	hubDump, err := hubDumper.Dump()
	if err != nil {
		panic(err)
	}
	hubReport := report.NewHubReport(hubDump)

	dict := map[string]interface{}{
		"Dump":   hubDump,
		"Report": hubReport,
	}

	bytes, err := json.MarshalIndent(dict, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(bytes))

	// TODO
}
