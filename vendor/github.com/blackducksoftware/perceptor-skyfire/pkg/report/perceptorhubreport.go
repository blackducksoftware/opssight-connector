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

package report

import (
	"fmt"
)

// PerceptorHubReport .....
type PerceptorHubReport struct {
	JustPerceptorImages []string
	JustHubImages       []string
	allHubImages        []string
	allHubImagesSet     map[string]bool
}

// NewPerceptorHubReport .....
func NewPerceptorHubReport(dump *Dump) *PerceptorHubReport {
	images := []string{}
	imagesSet := map[string]bool{}
	for _, hubDump := range dump.Hubs {
		for _, scan := range hubDump.Scans {
			images = append(images, scan.Name)
			imagesSet[scan.Name] = true
		}
	}
	return &PerceptorHubReport{
		JustPerceptorImages: PerceptorNotHubImages(dump, imagesSet),
		JustHubImages:       HubNotPerceptorImages(dump),
		allHubImages:        images,
		allHubImagesSet:     imagesSet,
	}
}

// HumanReadableString .....
func (p *PerceptorHubReport) HumanReadableString() string {
	return fmt.Sprintf(`
Perceptor<->Hub:
 - %d image(s) in Perceptor that were not in the Hub
 - %d image(s) in the Hub that were not in Perceptor
	`,
		len(p.JustPerceptorImages),
		len(p.JustHubImages))
}

// PerceptorNotHubImages .....
func PerceptorNotHubImages(dump *Dump, allHubImagesSet map[string]bool) []string {
	images := []string{}
	for sha := range dump.Perceptor.ImagesBySha {
		_, ok := allHubImagesSet[sha]
		if !ok {
			images = append(images, sha)
		}
	}
	return images
}

// HubNotPerceptorImages .....
func HubNotPerceptorImages(dump *Dump) []string {
	images := []string{}
	for _, hubDump := range dump.Hubs {
		for _, scan := range hubDump.Scans {
			if _, ok := dump.Perceptor.ImagesBySha[scan.Name]; !ok {
				images = append(images, scan.Name)
			}
		}
	}
	return images
}
