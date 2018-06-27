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

package mapper

import (
	"fmt"

	"github.com/blackducksoftware/perceivers/pkg/docker"

	perceptorapi "github.com/blackducksoftware/perceptor/pkg/api"

	imageapi "github.com/openshift/api/image/v1"

	metrics "github.com/blackducksoftware/perceivers/image/pkg/metrics"
)

// NewPerceptorImageFromOSImage will convert an openshift image object to a
// perceptor image object
func NewPerceptorImageFromOSImage(image *imageapi.Image) (*perceptorapi.Image, error) {
	dockerRef := image.DockerImageReference
	name, sha, err := docker.ParseImageIDString(dockerRef)
	if err != nil {
		metrics.RecordError("image_mapper", "unable to parse openshift imageID")
		return nil, fmt.Errorf("unable to parse openshift imageID %s: %v", dockerRef, err)
	}

	return perceptorapi.NewImage(name, sha, dockerRef), nil
}
