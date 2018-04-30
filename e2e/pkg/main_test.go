package main

import (
	"testing"

	skyfire "github.com/blackducksoftware/perceptor-skyfire/pkg/report"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCart(t *testing.T) {
	t.Log("«»")
}

func GrabLatestSkyfireReport() *skyfire.Report {
	return nil
}

var _ = Describe("Perceptor should scan all images on a cluster and annotate them correctly.", func() {
	BeforeSuite(func() {
		// submit 100 pods
	})
	It("Should annotate 200 different images", func() {

	})

})
