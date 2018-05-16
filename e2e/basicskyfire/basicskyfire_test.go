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
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	skyfire "github.com/blackducksoftware/perceptor-skyfire/pkg/report"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var skyfireHost string
var skyfirePort string

func init() {
	flag.StringVar(&skyfireHost, "skyfireHost", "", "skyfireHost is where to find skyfire")
	flag.StringVar(&skyfirePort, "skyfirePort", "", "skyfirePort is where to find skyfire port")
}

func TestBasicSkyfire(t *testing.T) {
	skyfireURL := fmt.Sprintf("http://%s:%s/latestreport", skyfireHost, skyfirePort)
	BasicSkyfireTests(skyfireURL)
	RegisterFailHandler(Fail)
	RunSpecs(t, "basic-skyfire")
}

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

func BasicSkyfireTests(skyfireURL string) {
	report, err := FetchSkyfireReport(skyfireURL)
	if err != nil {
		Fail(fmt.Sprintf("unable to fetch skyfire report from %s: %s", skyfireURL, err.Error()))
		return
	}

	Describe("All report data should be self-consistent", func() {
		It("All Kube data should be in order", func() {
			Expect(len(report.Kube.PartiallyAnnotatedPods)).Should(Equal(len(report.Kube.PartiallyAnnotatedPods)))
			Expect(len(report.Kube.PartiallyLabeledPods)).Should(Equal(len(report.Kube.PartiallyLabeledPods)))
			Expect(len(report.Kube.UnanalyzeablePods)).Should(Equal(len(report.Kube.UnanalyzeablePods)))
			Expect(len(report.Kube.UnparseableImages)).Should(Equal(len(report.Kube.UnparseableImages)))
		})

		It("All Kube<->Perceptor data should be in order", func() {
			Expect(len(report.KubePerceptor.ConflictingAnnotationsPods)).Should(Equal(len(report.KubePerceptor.ConflictingAnnotationsPods)))
			Expect(len(report.KubePerceptor.ConflictingLabelsPods)).Should(Equal(len(report.KubePerceptor.ConflictingLabelsPods)))
			Expect(len(report.KubePerceptor.FinishedJustKubePods)).Should(Equal(len(report.KubePerceptor.FinishedJustKubePods)))
			Expect(len(report.KubePerceptor.FinishedJustPerceptorPods)).Should(Equal(len(report.KubePerceptor.FinishedJustPerceptorPods)))
			Expect(len(report.KubePerceptor.JustKubeImages)).Should(Equal(len(report.KubePerceptor.JustKubeImages)))
			Expect(len(report.KubePerceptor.JustKubePods)).Should(Equal(len(report.KubePerceptor.JustKubePods)))
			Expect(len(report.KubePerceptor.JustPerceptorImages)).Should(Equal(len(report.KubePerceptor.JustPerceptorImages)))
			Expect(len(report.KubePerceptor.JustPerceptorPods)).Should(Equal(len(report.KubePerceptor.JustPerceptorPods)))
			Expect(len(report.KubePerceptor.UnanalyzeablePods)).Should(Equal(len(report.KubePerceptor.UnanalyzeablePods)))
		})

		It("All Perceptor<->Hub data should be in order", func() {
			Expect(len(report.PerceptorHub.JustHubImages)).Should(Equal(len(report.PerceptorHub.JustHubImages)))
			Expect(len(report.PerceptorHub.JustPerceptorImages)).Should(Equal(len(report.PerceptorHub.JustPerceptorImages)))
		})

		It("All Hub data should be in order", func() {
			Expect(len(report.Hub.ProjectsMultipleVersions)).Should(Equal(len(report.Hub.ProjectsMultipleVersions)))
			Expect(len(report.Hub.VersionsMultipleCodeLocations)).Should(Equal(len(report.Hub.VersionsMultipleCodeLocations)))
			Expect(len(report.Hub.CodeLocationsMultipleScanSummaries)).Should(Equal(len(report.Hub.CodeLocationsMultipleScanSummaries)))
		})
	})
}
