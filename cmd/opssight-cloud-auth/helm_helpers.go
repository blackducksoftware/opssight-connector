/*
Copyright (C) 2020 Synopsys, Inc.

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

package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

var settings = cli.New()

// UpdateWithHelm3 uses the helm NewUpgrade action to update a resource in the cluster
func UpdateWithHelm3(releaseName, namespace, chartURL string, vals map[string]interface{}, extraFiles ...string) error {
	actionConfig, err := CreateHelmActionConfiguration(namespace)
	if err != nil {
		return err
	}
	if releaseExists := ReleaseExists(releaseName, namespace); !releaseExists {
		return fmt.Errorf("release '%s' does not exist", releaseName)
	}

	chart, err := LoadChart(chartURL, actionConfig)
	if err != nil {
		return fmt.Errorf("failed to load release at '%s' for updating: %s", chartURL, err)
	}
	validInstallableChart, err := isChartInstallable(chart)
	if !validInstallableChart {
		return fmt.Errorf("release at '%s' is not installable: %s", chartURL, err)
	}

	client := action.NewUpgrade(actionConfig)
	if client.Version == "" && client.Devel {
		client.Version = ">0.0.0-0"
	}
	client.Namespace = namespace

	if err := mergeValuesWithExtraFilesFromChart(chart, vals, extraFiles); err != nil {
		return fmt.Errorf("failed to merge extra configuration files during update due to %s", err)
	}

	client.ResetValues = true                     // rememeber the values that have been set previously
	_, err = client.Run(releaseName, chart, vals) // updates the release in the namespace from the actionConfig
	if err != nil {
		return fmt.Errorf("failed to run upgrade: %s", err)
	}
	return nil
}

// GetWithHelm3 uses the helm NewGet action to return a Release with information about
// a resource from the cluster
func GetWithHelm3(releaseName, namespace string) (*release.Release, error) {
	actionConfig, err := CreateHelmActionConfiguration(namespace)
	if err != nil {
		return nil, err
	}
	// Check if the Release Exists
	aList := action.NewList(actionConfig) // NewGet provides bad error message if release doesn't exist
	charts, err := aList.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run get due to %s", err)
	}
	for _, release := range charts {
		if release.Name == releaseName && release.Namespace == namespace {
			return release, nil
		}
	}
	return nil, fmt.Errorf("unable to find release '%s' in namespace '%s'", releaseName, namespace)
}

// CreateHelmActionConfiguration creates an action.Configuration that points to the specified cluster and namespace
func CreateHelmActionConfiguration(namespace string) (*action.Configuration, error) {
	// TODO: look into using GetActionConfigurations()
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(nil, namespace, "secret", func(format string, v ...interface{}) {}); err != nil {
		return nil, err
	}
	return actionConfig, nil
}

type configFlagsWithTransport struct {
	*genericclioptions.ConfigFlags
	Transport *http.RoundTripper
}

// GetActionConfigurations creates an action.Configuration that points to the specified cluster and namespace
// TODO - this function specifies more values than CreateHelmActionConfiguration(), consider using this
func GetActionConfigurations(host, namespace, token string, transport *http.RoundTripper) *action.Configuration {

	confFlags := &configFlagsWithTransport{
		ConfigFlags: &genericclioptions.ConfigFlags{
			APIServer:   &host,
			BearerToken: &token,
			Namespace:   &namespace,
		},
		Transport: transport,
	}
	inClusterCfg, err := rest.InClusterConfig()

	if err != nil {
		fmt.Print("Running outside cluster, CAFile is unset")
	} else {
		confFlags.CAFile = &inClusterCfg.CAFile
	}
	conf := new(action.Configuration)
	conf.Init(confFlags, namespace, "secrets", klog.Infof)

	return conf
}

// LoadChart returns a chart from the specified chartURL
// Modified from https://github.com/openshift/console/blob/master/pkg/helm/actions/template_test.go
func LoadChart(chartURL string, actionConfig *action.Configuration) (*chart.Chart, error) {
	client := action.NewInstall(actionConfig)

	// Get full path - checks local machine and chart repository
	chartFullPath, err := client.ChartPathOptions.LocateChart(chartURL, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to locate resources at '%s' due to %s", chartURL, err)
	}

	chart, err := loader.Load(chartFullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load resources from '%s' due to %s", chartFullPath, err)
	}
	return chart, nil
}

// isChartInstallable validates if a chart can be installed
// Only the "application" chart type is installable
func isChartInstallable(ch *chart.Chart) (bool, error) {
	switch ch.Metadata.Type {
	case "", "application":
		return true, nil
	}
	return false, fmt.Errorf("resources at '%s' are not installable", ch.Metadata.Type)
}

// ReleaseExists verifies that a resources is deployed in the cluster
func ReleaseExists(releaseName, namespace string) bool {
	actionConfig, err := CreateHelmActionConfiguration(namespace)
	if err != nil {
		return false
	}
	client := action.NewGet(actionConfig)
	release, err := client.Run(releaseName) // lists the releases in the namespace from the actionConfig
	if err != nil || release == nil {
		return false
	}
	return true
}

// mergeValuesWithExtraFilesFromChart merges all extra files from chart with the values map
func mergeValuesWithExtraFilesFromChart(ch *chart.Chart, vals map[string]interface{}, extraFiles []string) error {
	for _, fileName := range extraFiles {
		found := false
		for _, chartFile := range ch.Files {
			if fileName == chartFile.Name {
				found = true
				patch := make(map[string]interface{})
				if err := yaml.Unmarshal(chartFile.Data, &patch); err != nil {
					return err
				}
				if err := mergo.Merge(&vals, patch, mergo.WithOverride); err != nil {
					return err
				}
				break
			}
		}
		if !found {
			return fmt.Errorf("couldn't find file '%s' in release resources", fileName)
		}
	}
	return nil
}

// SetHelmValueInMap adds the finalValue into the valueMapPointer at the location specified
// by the keyList
// valueMapPointer - a map for helm values (maps are pointers in Golang)
//  - it is used to track the current map being updated
// keyList - an ordered list of keys that lead to the location in the valueMapPointer to place the finalValue
// finalValue - the value to set in the map
func SetHelmValueInMap(valueMapPointer map[string]interface{}, keyList []string, finalValue interface{}) {
	for i, currKey := range keyList {
		if i == (len(keyList) - 1) { // at the last key -> set the value
			valueMapPointer[currKey] = finalValue
			return
		}
		if nextMap, _ := valueMapPointer[currKey]; nextMap != nil { // key is in map -> go to next map
			valueMapPointer = nextMap.(map[string]interface{})
		} else { // key is not in the map -> add the key and next key; go to next map
			nextKey := keyList[i+1]
			valueMapPointer[currKey] = map[string]interface{}{nextKey: nil}
			nextMap := valueMapPointer[currKey].(map[string]interface{})
			valueMapPointer = nextMap
		}
	}
}

// GetHelmValueFromMap returns an interface{} if the value exists in the map
func GetHelmValueFromMap(valueMapPointer map[string]interface{}, keyList []string) interface{} {
	for i, currKey := range keyList {
		currVal, ok := valueMapPointer[currKey]
		if !ok { // value doesn't exist - invalid path
			return nil
		}
		if i == (len(keyList) - 1) { // at the last key -> return the value
			return currVal
		}
		valueMapPointer, ok = currVal.(map[string]interface{})
		if !ok { // got a finalValue instead of a map - invalid path
			return nil
		}
	}
	return nil
}

// GetValueFromRelease merges the default Chart Values with the user's set values
// to find the value that is current set in the Release
func GetValueFromRelease(release *release.Release, keyList []string) interface{} {
	chartValues := release.Chart.Values
	userConfig := release.Config
	releaseValues := MergeMaps(chartValues, userConfig)
	return GetHelmValueFromMap(releaseValues, keyList)
}

// DeepCopyHelmValuesMap copies the src map to the dest map. Values in new map have different pointers/references
func DeepCopyHelmValuesMap(src map[string]interface{}, dest map[string]interface{}) error {
	jsonStr, err := json.Marshal(src)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonStr, &dest)
	if err != nil {
		return err
	}
	return nil
}

// GetReleaseValues merges the default Chart Values with the user's set values
// and retuns that set of values
func GetReleaseValues(release *release.Release) map[string]interface{} {
	chartValues := release.Chart.Values
	userConfig := release.Config
	return MergeMaps(chartValues, userConfig)
}

// MergeMaps Copied from https://github.com/helm/helm/blob/9b42702a4bced339ff424a78ad68dd6be6e1a80a/pkg/cli/values/options.go#L88
func MergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = MergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}

func setStringPtrInHelmValueInMap(valueMapPointer map[string]interface{}, keyList []string, value *string) {
	if value != nil {
		SetHelmValueInMap(valueMapPointer, keyList, *value)
	}
}
