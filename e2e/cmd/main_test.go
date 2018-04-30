package main

import (
    "testing"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    skyfire "github.com/blackducksoftware/perceptor-skyfire/pkg/report"
)

func TestCart(t *testing.T) {
    t.Log("«»")
}

func GrabLatestSkyfireReport() *skyfire.Report {

}

var _ = Describe("Perceptor should scan images from newly created pods", func() {
    Context("initially", func() {
        It("has 0 items", func() {})
        It("has 0 units", func() {})
        Specify("the total amount is 0.00", func() {})
    })

    Context("that has 2 units of item A", func() {

        Context("removing 1 unit of item A", func() {
            It("should not reduce the number of items", func() {})
            It("should reduce the number of units by 1", func() {})
            It("should reduce the amount by the item price", func() {})
        })

        Context("removing 2 units of item A", func() {
            It("should reduce the number of items by 1", func() {})
            It("should reduce the number of units by 2", func() {})
            It("should reduce the amount by twice the item price", func() {})
        })
    })
})
