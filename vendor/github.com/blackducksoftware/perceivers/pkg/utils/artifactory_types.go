/*
Copyright (C) 2019 Synopsys, Inc.

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

package utils

// ArtDockerRepo contains list of docker repos in artifactory
type ArtDockerRepo []struct {
	Key         string `json:"key"`
	Type        string `json:"type"`
	URL         string `json:"url"`
	PackageType string `json:"packageType"`
}

// ArtImages contain list of images inside the docker repo
type ArtImages struct {
	Repositories []string `json:"repositories"`
}

// ArtImageTags lists out all the tags for the image
type ArtImageTags struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// ArtImageSHAs gets all the sha256 of an image
type ArtImageSHAs struct {
	Properties struct {
		Sha256 []string `json:"sha256"`
	} `json:"properties"`
	URI string `json:"uri"`
}

// ArtReposBySha collects URIs for given SHA256
type ArtReposBySha struct {
	Results []struct {
		URI string `json:"uri"`
	} `json:"results"`
}

// ArtHookStruct is the structure returned by Artifactory webhook
type ArtHookStruct struct {
	Artifacts []struct {
		Type      string `json:"type"`
		Name      string `json:"name"`
		Version   string `json:"version"`
		Reference string `json:"reference"`
	} `json:"artifacts"`
}
