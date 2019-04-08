/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreementd. See the NOTICE file
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

package components

import (
	"fmt"
	"strings"

	"github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/util"

	"github.com/koki/short/converter/converters"
	"github.com/koki/short/types"

	"k8s.io/apimachinery/pkg/runtime"
)

// DaemonSet defines the daemon set component
type DaemonSet struct {
	obj *types.DaemonSet
}

// NewDaemonSet creates a DaemonSet object
func NewDaemonSet(config api.DaemonSetConfig) *DaemonSet {
	d := &types.DaemonSet{
		Version:              config.APIVersion,
		Cluster:              config.ClusterName,
		Name:                 config.Name,
		Namespace:            config.Namespace,
		MinReadySeconds:      config.MinReadySeconds,
		RevisionHistoryLimit: config.RevisionHistoryLimit,
	}

	if config.UpdateStrategy == api.DaemonSetUpdateStrategyOnDelete {
		d.OnDelete = true
	} else {
		d.OnDelete = false
	}

	if len(config.MaxUnavailable) > 0 {
		d.MaxUnavailable = createIntOrStr(config.MaxUnavailable)
	}

	return &DaemonSet{obj: d}
}

// GetObj returns the daemon set object in a format the deployer can use
func (d *DaemonSet) GetObj() *types.DaemonSet {
	return d.obj
}

// GetName returns the name of the daemon set
func (d *DaemonSet) GetName() string {
	return d.obj.Name
}

// AddAnnotations adds annotations to the daemon set
func (d *DaemonSet) AddAnnotations(new map[string]string) {
	d.obj.Annotations = util.MapMerge(d.obj.Annotations, new)
}

// RemoveAnnotations removes annotations from the daemon set
func (d *DaemonSet) RemoveAnnotations(remove []string) {
	for _, k := range remove {
		d.obj.Annotations = util.RemoveElement(d.obj.Annotations, k)
	}
}

// AddLabels adds labels to the daemon set
func (d *DaemonSet) AddLabels(new map[string]string) {
	d.obj.Labels = util.MapMerge(d.obj.Labels, new)
}

// RemoveLabels removes labels from the daemon set
func (d *DaemonSet) RemoveLabels(remove []string) {
	for _, k := range remove {
		d.obj.Labels = util.RemoveElement(d.obj.Labels, k)
	}
}

// AddPod adds a pod to the daemon set
func (d *DaemonSet) AddPod(obj *Pod) error {
	o := obj.GetObj()
	d.obj.TemplateMetadata = &o.PodTemplateMeta
	d.obj.PodTemplate = o.PodTemplate

	return nil
}

// RemovePod removes a pod from the daemon set
func (d *DaemonSet) RemovePod(name string) error {
	if strings.Compare(d.obj.TemplateMetadata.Name, name) != 0 {
		return fmt.Errorf("pod with name %s doesn't exist on daemon set", name)
	}
	d.obj.TemplateMetadata = nil
	d.obj.PodTemplate = types.PodTemplate{}
	return nil
}

// AddMatchLabelsSelectors adds the given match label selectors to the daemon set
func (d *DaemonSet) AddMatchLabelsSelectors(new map[string]string) {
	if d.obj.Selector == nil {
		d.obj.Selector = &types.RSSelector{}
	}
	d.obj.Selector.Labels = util.MapMerge(d.obj.Selector.Labels, new)
}

// RemoveMatchLabelsSelectors removes the given match label selectors from the daemon set
func (d *DaemonSet) RemoveMatchLabelsSelectors(remove []string) {
	for _, k := range remove {
		d.obj.Selector.Labels = util.RemoveElement(d.obj.Selector.Labels, k)
	}
}

// AddMatchExpressionsSelector will add match expressions selectors to the daemon set.
// It takes a string in the following form:
// key <op> <value>
// Where op can be:
// = 	Equal to value ot should be one of the comma separated values
// !=	Key should not be one of the comma separated values
// If no op is provided, then the key should (or should not) exist
// <key>	key should exist
// !<key>	key should not exist
func (d *DaemonSet) AddMatchExpressionsSelector(add string) {
	d.obj.Selector.Shorthand = add
}

// RemoveMatchExpressionsSelector removes the match expressions selector from the daemon set
func (d *DaemonSet) RemoveMatchExpressionsSelector() {
	d.obj.Selector.Shorthand = ""
}

// ToKube returns the kubernetes version of the daemon set
func (d *DaemonSet) ToKube() (runtime.Object, error) {
	wrapper := &types.DaemonSetWrapper{DaemonSet: *d.obj}
	return converters.Convert_Koki_DaemonSet_to_Kube_apps_v1beta2_DaemonSet(wrapper)
}
