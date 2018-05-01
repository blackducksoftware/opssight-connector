package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	skyfire "github.com/blackducksoftware/perceptor-skyfire/pkg/report"
	ginkgo "github.com/onsi/ginkgo"
	gomega "github.com/onsi/gomega"
)

var skyfireBaseURL string

func init() {
	flag.StringVar(&skyfireBaseURL, "skyfireBaseURL", "", "skyfireBaseURL is where to find skyfire")
}

func TestBasicSkyfire(t *testing.T) {
	fmt.Println(os.Args)
	fmt.Printf("skyfire base URL: %s\n", skyfireBaseURL)
	skyfireURL := fmt.Sprintf("%s/latestreport", skyfireBaseURL)
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

		ginkgo.It("All Hub data should be in order", func() {
			gomega.Expect(len(report.Hub.ProjectsMultipleVersions)).Should(gomega.Equal(0))
			gomega.Expect(len(report.Hub.VersionsMultipleCodeLocations)).Should(gomega.Equal(0))
			gomega.Expect(len(report.Hub.CodeLocationsMultipleScanSummaries)).Should(gomega.Equal(0))
		})
	})
}
