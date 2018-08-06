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
	"testing"
	"time"

	skyfire "github.com/blackducksoftware/perceptor-skyfire/pkg/report"
	ginkgo "github.com/onsi/ginkgo"
	gomega "github.com/onsi/gomega"
)

func TestBasicSkyfire(t *testing.T) {
	skyfireURL := "http://skyfire:3187/latestreport"
	BasicSkyfireTests(skyfireURL)
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "basic-skyfire")
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
		ginkgo.Fail(fmt.Sprintf("unable to fetch skyfire report from %s: %s", skyfireURL, err.Error()))
		return
	}

	ginkgo.Describe("All report data should be self-consistent", func() {
		ginkgo.It("All Kube data should be in order", func() {
			gomega.Expect(len(report.Kube.PartiallyAnnotatedPods)).Should(gomega.Equal(0))
			gomega.Expect(len(report.Kube.PartiallyLabeledPods)).Should(gomega.Equal(0))
			gomega.Expect(len(report.Kube.UnanalyzeablePods)).Should(gomega.Equal(0))
			gomega.Expect(len(report.Kube.UnparseableImages)).Should(gomega.Equal(0))
		})

		ginkgo.It("All Kube<->Perceptor data should be in order", func() {
			gomega.Expect(len(report.KubePerceptor.ConflictingAnnotationsPods)).Should(gomega.Equal(0))
			gomega.Expect(len(report.KubePerceptor.ConflictingLabelsPods)).Should(gomega.Equal(0))
			gomega.Expect(len(report.KubePerceptor.FinishedJustKubePods)).Should(gomega.Equal(0))
			gomega.Expect(len(report.KubePerceptor.FinishedJustPerceptorPods)).Should(gomega.Equal(0))
			gomega.Expect(len(report.KubePerceptor.JustKubeImages)).Should(gomega.Equal(0))
			gomega.Expect(len(report.KubePerceptor.JustKubePods)).Should(gomega.Equal(0))
			gomega.Expect(len(report.KubePerceptor.JustPerceptorImages)).Should(gomega.Equal(0))
			gomega.Expect(len(report.KubePerceptor.JustPerceptorPods)).Should(gomega.Equal(0))
			gomega.Expect(len(report.KubePerceptor.UnanalyzeablePods)).Should(gomega.Equal(0))
		})

		ginkgo.It("All Perceptor<->Hub data should be in order", func() {
			gomega.Expect(len(report.PerceptorHub.JustHubImages)).Should(gomega.Equal(0))
			gomega.Expect(len(report.PerceptorHub.JustPerceptorImages)).Should(gomega.Equal(0))
		})

		ginkgo.It("All Hub data should be in order", func() {
			gomega.Expect(len(report.Hub.ProjectsMultipleVersions)).Should(gomega.Equal(0))
			gomega.Expect(len(report.Hub.VersionsMultipleCodeLocations)).Should(gomega.Equal(0))
			gomega.Expect(len(report.Hub.CodeLocationsMultipleScanSummaries)).Should(gomega.Equal(0))
		})
	})
}
