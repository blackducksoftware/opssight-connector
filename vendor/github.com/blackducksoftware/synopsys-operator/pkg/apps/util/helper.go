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

package util

import (
	"fmt"
	"strings"

	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
)

// GenerateImageTag return the final image after evaluating the registry configuration and the list of images
func GenerateImageTag(defaultImage string, imageRegistries []string, registryConfig *api.RegistryConfiguration) string {
	if len(imageRegistries) > 0 {
		imageName, err := util.GetImageName(defaultImage)
		if err != nil {
			return defaultImage
		}
		if image := getFullContainerNameFromImageRegistryConf(imageName, imageRegistries); len(image) > 0 {
			return image
		}
	}

	if registryConfig != nil && len(registryConfig.Registry) > 0 {
		return getRegistryConfiguration(defaultImage, registryConfig)
	}
	return defaultImage
}

func getRegistryConfiguration(image string, registryConfig *api.RegistryConfiguration) string {
	if registryConfig != nil && len(registryConfig.Registry) > 0 {
		imageName, err := util.GetImageName(image)
		if err != nil {
			log.Errorf("unable to get the image name for %s due to %+v", image, err)
			return image
		}
		imageTag, err := util.GetImageTag(image)
		if err != nil {
			log.Errorf("unable to get the image tag for %s due to %+v", image, err)
			return image
		}
		return fmt.Sprintf("%s/%s:%s", registryConfig.Registry, imageName, imageTag)
	}
	return image
}

func getFullContainerNameFromImageRegistryConf(baseContainer string, images []string) string {
	for _, reg := range images {
		// normal case: we expect registries
		if strings.Contains(reg, baseContainer) {
			_, err := util.ValidateImageString(reg)
			if err != nil {
				break
			}
			return reg
		}
	}
	return ""
}
