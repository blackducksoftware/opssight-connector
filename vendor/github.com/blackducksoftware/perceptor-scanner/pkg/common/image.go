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

package common

import (
	"fmt"
	"strings"
)

// Image ...
type Image struct {
	Directory string
	PullSpec  string
}

// NewImage ...
func NewImage(directory string, pullSpec string) *Image {
	return &Image{Directory: directory, PullSpec: pullSpec}
}

// DockerPullSpec ...
func (image *Image) DockerPullSpec() string {
	return image.PullSpec
}

// DockerTarFilePath ...
func (image *Image) DockerTarFilePath() string {
	return fmt.Sprintf("%s/%s.tar", image.Directory, strings.Replace(image.PullSpec, "/", "_", -1))
}
