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

package util

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
)

// Container defines the configuration for a container
type Container struct {
	ContainerConfig       *horizonapi.ContainerConfig
	EnvConfigs            []*horizonapi.EnvConfig
	VolumeMounts          []*horizonapi.VolumeMountConfig
	PortConfig            []*horizonapi.PortConfig
	ActionConfig          *horizonapi.ActionConfig
	ReadinessProbeConfigs []*horizonapi.ProbeConfig
	LivenessProbeConfigs  []*horizonapi.ProbeConfig
	PreStopConfig         *horizonapi.ActionConfig
}

// PodConfig used for configuring the pod
type PodConfig struct {
	Name                   string
	Labels                 map[string]string
	ServiceAccount         string
	Containers             []*Container
	Volumes                []*components.Volume
	InitContainers         []*Container
	PodAffinityConfigs     map[horizonapi.AffinityType][]*horizonapi.PodAffinityConfig
	PodAntiAffinityConfigs map[horizonapi.AffinityType][]*horizonapi.PodAffinityConfig
	NodeAffinityConfigs    map[horizonapi.AffinityType][]*horizonapi.NodeAffinityConfig
	ImagePullSecrets       []string
	FSGID                  *int64
}

// MergeEnvMaps will merge the source and destination environs. If the same value exist in both, source environ will given more preference
func MergeEnvMaps(source, destination map[string]string) map[string]string {
	// if the source key present in the destination map, it will overrides the destination value
	// if the source value is empty, then delete it from the destination
	for key, value := range source {
		if len(value) == 0 {
			delete(destination, key)
		} else {
			destination[key] = value
		}
	}
	return destination
}

// MergeEnvSlices will merge the source and destination environs. If the same value exist in both, source environ will given more preference
func MergeEnvSlices(source, destination []string) []string {
	// create a destination map
	destinationMap := make(map[string]string)
	for _, value := range destination {
		values := strings.SplitN(value, ":", 2)
		if len(values) == 2 {
			mapKey := strings.TrimSpace(values[0])
			mapValue := strings.TrimSpace(values[1])
			if len(mapKey) > 0 && len(mapValue) > 0 {
				destinationMap[mapKey] = mapValue
			}
		}
	}

	// if the source key present in the destination map, it will overrides the destination value
	// if the source value is empty, then delete it from the destination
	for _, value := range source {
		values := strings.SplitN(value, ":", 2)
		if len(values) == 2 {
			mapKey := strings.TrimSpace(values[0])
			mapValue := strings.TrimSpace(values[1])
			if len(mapValue) == 0 {
				delete(destinationMap, mapKey)
			} else {
				destinationMap[mapKey] = mapValue
			}
		}
	}

	// convert destination map to string array
	mergedValues := []string{}
	for key, value := range destinationMap {
		mergedValues = append(mergedValues, fmt.Sprintf("%s:%s", key, value))
	}
	return mergedValues
}

// UniqueStringSlice returns a unique subset of the string slice provided.
func UniqueStringSlice(input []string) []string {
	output := []string{}
	m := make(map[string]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			output = append(output, val)
		}
	}

	return output
}

// GetResourceName returns the name of the resource
func GetResourceName(name string, appName string, defaultName string) string {
	if len(appName) == 0 {
		return fmt.Sprintf("%s-%s", name, defaultName)
	}

	if len(defaultName) == 0 {
		return fmt.Sprintf("%s-%s", name, appName)
	}

	return fmt.Sprintf("%s-%s-%s", name, appName, defaultName)
}

// RemoveFromStringSlice will remove the string from the slice and it will maintain the order
func RemoveFromStringSlice(slice []string, str string) []string {
	for index, value := range slice {
		if value == str {
			slice = append(slice[:index], slice[index+1:]...)
		}
	}
	return slice
}

// IsExistInStringSlice will check for the input string in the given slice
func IsExistInStringSlice(slice []string, str string) bool {
	for _, value := range slice {
		if value == str {
			return true
		}
	}
	return false
}

// IsExposeServiceValid validates the expose service type
func IsExposeServiceValid(serviceType string) bool {
	switch strings.ToUpper(serviceType) {
	case NONE, NODEPORT, LOADBALANCER, OPENSHIFT:
		return true
	}
	return false
}

// IsBlackDuckVersionSupportMultipleInstance returns whether it supports multiple instance in a single namespace
func IsBlackDuckVersionSupportMultipleInstance(version string) (bool, error) {
	return isYearAndMonthGreaterThanOrEqualTo(version, 2019, time.August)
}

// isYearAndMonthGreaterThanOrEqualTo returns whether the given version is greater than or equal to the given year and month
func isYearAndMonthGreaterThanOrEqualTo(version string, year int, month time.Month) (bool, error) {
	versionArr := strings.Split(version, ".")
	if len(versionArr) >= 3 {
		t, err := time.Parse("2006.1", fmt.Sprintf("%s.%s", versionArr[0], versionArr[1]))
		if err != nil {
			return false, err
		}
		if t.Year() >= year && t.Month() >= month {
			return true, nil
		}
	}
	return false, nil
}

// IsVersionGreaterThanOrEqualTo returns whether the given version is greater than or equal to the given year, month and dot release
func IsVersionGreaterThanOrEqualTo(version string, year int, month time.Month, dotRelease int) (bool, error) {
	versionArr := strings.Split(version, ".")
	if len(versionArr) >= 3 {
		t, err := time.Parse("2006.1", fmt.Sprintf("%s.%s", versionArr[0], versionArr[1]))
		if err != nil {
			return false, err
		}

		minorDotVersion, err := strconv.Atoi(versionArr[2])
		if err != nil {
			return false, err
		}

		if t.Year() >= year && t.Month() >= month && minorDotVersion >= dotRelease {
			return true, nil
		}
	}
	return false, nil
}

// IsNotDefaultVersionGreaterThanOrEqualTo returns whether the given version is greater than or equal to the given inputs
func IsNotDefaultVersionGreaterThanOrEqualTo(version string, majorRelease int, minorRelease int, dotRelease int) (bool, error) {
	versionArr := strings.Split(version, ".")
	if len(versionArr) >= 3 {
		majorReleaseVersion, err := strconv.Atoi(versionArr[0])
		if err != nil {
			return false, err
		}
		minorReleaseVersion, err := strconv.Atoi(versionArr[1])
		if err != nil {
			return false, err
		}
		dotReleaseVersion, err := strconv.Atoi(versionArr[2])
		if err != nil {
			return false, err
		}
		if (majorReleaseVersion > majorRelease) ||
			(majorReleaseVersion == majorRelease && minorReleaseVersion > minorRelease) ||
			(majorReleaseVersion == majorRelease && minorReleaseVersion == minorRelease && dotReleaseVersion >= dotRelease) {
			return true, nil
		}
	}
	return false, nil
}

// StringArrayToMapSplitBySeparator converts the string array to map based on separator for each string in the array
func StringArrayToMapSplitBySeparator(strs []string, separator string) map[string]string {
	maps := make(map[string]string, 0)
	for _, str := range strs {
		strArr := strings.SplitN(str, separator, 2)
		if len(strArr) == 2 {
			maps[strArr[0]] = strArr[1]
		}
	}
	return maps
}

// MapToStringArrayJoinBySeparator converts the map to string array and each key and value of the map will be joined by separator
func MapToStringArrayJoinBySeparator(maps map[string]string, separator string) []string {
	strs := make([]string, 0)
	for key, value := range maps {
		strs = append(strs, fmt.Sprintf("%s%s%s", key, separator, value))
	}
	return strs
}
