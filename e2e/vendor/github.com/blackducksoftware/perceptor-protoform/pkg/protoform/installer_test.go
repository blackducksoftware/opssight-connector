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

package protoform

import (
	"os"
	"regexp"
	"testing"
)

func TestProto(t *testing.T) {
	os.Setenv("PCP_HUBUSERPASSWORD", "example")

	d := NewDefaultsObj()
	d.DefaultCPU = "300m"
	d.DefaultMem = "1300Mi"

	installer := NewInstaller(d, "../../cmd/protoform.json")
	installer.init()
	rcsArray := installer.replicationControllers

	var imageRegexp = regexp.MustCompile("(.+)/(.+):(.+)")

	args := map[string]string{
		"perceptor":             "/etc/perceptor/perceptor.yaml",
		"pod-perceiver":         "/etc/perceiver/perceiver.yaml",
		"image-perceiver":       "/etc/perceiver/perceiver.yaml",
		"perceptor-scanner":     "/etc/perceptor_scanner/perceptor_scanner.yaml",
		"perceptor-imagefacade": "/etc/perceptor_imagefacade/perceptor_imagefacade.yaml",
		"skyfire":               "/etc/skyfire/skyfire.yaml",
	}

	for _, rcs := range rcsArray {
		for _, container := range rcs.Containers {

			// verify the image expressions
			match := imageRegexp.FindStringSubmatch(container.Image)
			if len(match) != 4 {
				t.Errorf("%s is not matching to the regex %s", container.Image, imageRegexp.String())
			}

			// verify the args parameter in Replication Controller
			arg := container.Args

			if arg[0].String() != args[container.Name] {
				t.Errorf("Arguments not matched for %s, Expected: %s, Actual: %s", container.Name, args[container.Name], arg[0].String())
			}

			// verify the default cpu parameters
			if d.DefaultCPU != container.CPU.Min {
				t.Errorf("Default CPU is not configured for %s, Expected: %s, Actual: %s", container.Name, d.DefaultCPU, container.CPU.Min)
			}

			// verify the default memory parameters
			if d.DefaultMem != container.Mem.Min {
				t.Errorf("Default memory is not configured for %s, Expected: %s, Actual: %s", container.Name, d.DefaultMem, container.Mem.Min)
			}

		}
	}

	// Image facade needs to be privileged !
	if *rcsArray[2].PodTemplate.Containers[1].Privileged == false {
		t.Errorf("%v %v", rcsArray[2].PodTemplate.Containers[1].Name, *rcsArray[2].PodTemplate.Containers[1].Privileged)
	}

	// The scanner needs to be UNPRIVILEGED
	if *rcsArray[2].PodTemplate.Containers[0].Privileged == true {
		t.Errorf("%v %v", rcsArray[2].PodTemplate.Containers[0].Name, *rcsArray[2].PodTemplate.Containers[0].Privileged)
	}

	t.Logf("template: %v ", rcsArray[2].PodTemplate)
	scannerSvc := rcsArray[2].PodTemplate.Account
	if scannerSvc == "" {
		t.Errorf("scanner svc ==> ( %v ) EMPTY !", scannerSvc)
	}

	s0 := rcsArray[2].PodTemplate.Containers[0].Name
	s := rcsArray[2].PodTemplate.Containers[0].VolumeMounts[1].Store
	if s != "var-images" {
		t.Errorf("%v %v", s0, s)
	}
}
