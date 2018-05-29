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
	"fmt"
	"time"

	skyfire "github.com/blackducksoftware/perceptor-skyfire/pkg/report"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

// func TestLoad(t *testing.T) {
// 	skyfireURL := fmt.Sprintf("http://%s:%s/latestreport", config.SkyfireHost, config.SkyfirePort)
// 	LoadTests(skyfireURL)
// 	RegisterFailHandler(Fail)
// 	RunSpecs(t, "load-test")
// }

func LoadTests(skyfireURL string) {
	var report *skyfire.Report
	var err error
	for {
		report, err = fetchSkyfireReport(skyfireURL)
		if err != nil {
			Fail(fmt.Sprintf("unable to fetch skyfire report from %s: %s", skyfireURL, err.Error()))
			return
		}

		log.Debugf("report: %v", report)
		if report != nil {
			break
		} else {
			time.Sleep(10 * time.Second)
		}
	}

	log.Debugln("Outside the skyfire for loop")

	images := loadPodsIntoCluster()

	log.Infof("Number of images created %d", len(images))

	Describe("Load tons of pods into a cluster", func() {
		It("Should actually have created all the pods in the cluster", func() {
			for _, image := range images {
				for {
					if len(report.Dump.Kube.PodsByName[image.PodName].Name) == 0 {
						time.Sleep(5 * time.Second)
					}
				}
				log.Infof("pod name: %s, image: %s:%s", image.PodName, image.ImageName, image.Tag)
				Expect(image.PodName).Should(Equal(report.Dump.Kube.PodsByName[image.PodName].Name))
			}
		})

		It("Should have all pods in Perceptor", func() {
			for _, image := range images {
				log.Infof("pod name: %s, image: %s:%s", image.PodName, image.ImageName, image.Tag)
				for {
					if len(report.Dump.Perceptor.PodsByName[image.PodName].Name) == 0 {
						time.Sleep(10 * time.Second)
					}
				}
				Expect(image.PodName).Should(Equal(report.Dump.Perceptor.PodsByName[image.PodName].Name))
			}
		})

		It("Should have correct annotations and labels for pods for which all images have been scanned", func() {
			// Check the status of the pod, if the pod status is completed, get the project version for the pod and validate it against the annotation", func() {

		})

	})

}

func loadPodsIntoCluster() []Image {
	dockerClient, err := NewDocker()
	if err != nil {
		log.Errorf("Unable to instantiate Docker client due to %+v", err)
	}
	images := dockerClient.GetDockerImages(config.NoOfPods)

	for _, image := range images {
		log.Infof("pod name: %s, image: %s:%s", image.PodName, image.ImageName, image.Tag)
		addPods(image.PodName, fmt.Sprintf("%s:%s", image.ImageName, image.Tag), int32(3007))
	}

	createPods(configPath)

	return images
}
