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

type PerceptorHubReport struct {
	JustPerceptorImages []string
	JustHubImages       []string
}

func NewPerceptorHubReport(dump *Dump) *PerceptorHubReport {
	return &PerceptorHubReport{
		JustPerceptorImages: PerceptorNotHubImages(dump),
		JustHubImages:       HubNotPerceptorImages(dump),
	}
}

func (p *PerceptorHubReport) HumanReadableString() string {
	return fmt.Sprintf(`
Perceptor<->Hub:
 - %d image(s) in Perceptor that were not in the Hub
 - %d image(s) in the Hub that were not in Perceptor
	`,
		len(p.JustPerceptorImages),
		len(p.JustHubImages))
}

func PerceptorNotHubImages(dump *Dump) []string {
	images := []string{}
	for sha := range dump.Perceptor.ImagesBySha {
		sha20 := sha[:20]
		_, ok := dump.Hub.ProjectsBySha[sha20]
		if !ok {
			images = append(images, sha)
		}
	}
	return images
}

func HubNotPerceptorImages(dump *Dump) []string {
	images := []string{}
	for sha := range dump.Hub.ProjectsBySha {
		foundMatch := false
		// can't do a dictionary lookup, because hub sha only has first 20 characters
		for _, perceptorImage := range dump.Perceptor.ScanResults.Images {
			if perceptorImage.Sha[:20] == sha {
				foundMatch = true
				break
			}
		}
		if !foundMatch {
			images = append(images, sha)
		}
	}
	return images
}
