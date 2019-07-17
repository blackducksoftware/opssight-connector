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

package components

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/util"

	"github.com/koki/short/converter/converters"
	"github.com/koki/short/types"

	"k8s.io/apimachinery/pkg/runtime"
)

// Pod defines the pod component
type Pod struct {
	obj                 *types.Pod
	defaultSELinux      *api.SELinuxType
	defaultRunUID       *int64
	defaultRunGID       *int64
	defaultForceNonRoot *bool
}

// NewPod create a Pod object
func NewPod(config api.PodConfig) *Pod {
	p := &types.Pod{
		Version: config.APIVersion,
		PodTemplateMeta: types.PodTemplateMeta{
			Cluster:   config.ClusterName,
			Name:      config.Name,
			Namespace: config.Namespace,
		},
		PodTemplate: types.PodTemplate{
			Account:                config.ServiceAccount,
			TerminationGracePeriod: config.TerminationGracePeriod,
			ActiveDeadline:         config.ActiveDeadline,
			Node:                   config.Node,
			FSGID:                  config.FSGID,
			Hostname:               config.Hostname,
			SchedulerName:          config.SchedulerName,
		},
	}

	switch config.RestartPolicy {
	case api.RestartPolicyAlways:
		p.PodTemplate.RestartPolicy = types.RestartPolicyAlways
	case api.RestartPolicyOnFailure:
		p.PodTemplate.RestartPolicy = types.RestartPolicyOnFailure
	case api.RestartPolicyNever:
		p.PodTemplate.RestartPolicy = types.RestartPolicyNever
	}

	switch config.DNSPolicy {
	case api.DNSClusterFirstWithHostNet:
		p.PodTemplate.DNSPolicy = types.DNSClusterFirstWithHostNet
	case api.DNSClusterFirst:
		p.PodTemplate.DNSPolicy = types.DNSClusterFirst
	case api.DNSDefault:
		p.PodTemplate.DNSPolicy = types.DNSDefault
	default:
		p.PodTemplate.DNSPolicy = types.DNSDefault
	}

	if config.PriorityValue != nil || len(config.PriorityClass) > 0 {
		p.Priority = &types.Priority{
			Value: config.PriorityValue,
			Class: config.PriorityClass,
		}
	}

	return &Pod{
		obj:                 p,
		defaultForceNonRoot: config.ForceNonRoot,
		defaultRunGID:       config.RunAsGroup,
		defaultRunUID:       config.RunAsUser,
		defaultSELinux:      config.SELinux,
	}
}

// GetObj returns the pod object in a format the deployer can use
func (p *Pod) GetObj() *types.Pod {
	return p.obj
}

// GetName returns the name of the pod
func (p *Pod) GetName() string {
	return p.obj.PodTemplateMeta.Name
}

// AddAnnotations adds annotations to the pod
func (p *Pod) AddAnnotations(new map[string]string) {
	p.obj.Annotations = util.MapMerge(p.obj.Annotations, new)
}

// RemoveAnnotations removes annotations from the pod
func (p *Pod) RemoveAnnotations(remove []string) {
	for _, k := range remove {
		p.obj.Annotations = util.RemoveElement(p.obj.Annotations, k)
	}
}

// AddLabels adds labels to the pod
func (p *Pod) AddLabels(new map[string]string) {
	p.obj.Labels = util.MapMerge(p.obj.Labels, new)
}

// RemoveLabels removes labels from the pod
func (p *Pod) RemoveLabels(remove []string) {
	for _, k := range remove {
		p.obj.Labels = util.RemoveElement(p.obj.Labels, k)
	}
}

// AddContainer adds a container to the pod
func (p *Pod) AddContainer(obj *Container) error {
	if p.findContainerPos(obj.GetName(), p.obj.Containers) >= 0 {
		return fmt.Errorf("container with name %s already exists", obj.GetName())
	}

	o := obj.GetObj()
	p.addContainerDefaults(o)
	p.obj.Containers = append(p.obj.Containers, *o)

	return nil
}

// RemoveContainer removes a container from the pod
func (p *Pod) RemoveContainer(name string) error {
	loc := p.findContainerPos(name, p.obj.Containers)
	if loc < 0 {
		return fmt.Errorf("container with name %s doesn't exist", name)
	}
	p.obj.Containers = append(p.obj.Containers[:loc], p.obj.Containers[loc+1:]...)
	return nil
}

// AddInitContainer adds an init container to the pod
func (p *Pod) AddInitContainer(obj *Container) error {
	if p.findContainerPos(obj.GetName(), p.obj.InitContainers) >= 0 {
		return fmt.Errorf("init container with name %s already exists", obj.GetName())
	}

	o := obj.GetObj()
	p.addContainerDefaults(o)
	p.obj.InitContainers = append(p.obj.InitContainers, *o)

	return nil
}

// RemoveInitContainer removes an init container from the pod
func (p *Pod) RemoveInitContainer(name string) error {
	loc := p.findContainerPos(name, p.obj.InitContainers)
	if loc < 0 {
		return fmt.Errorf("init container with name %s doesn't exist", name)
	}
	p.obj.Containers = append(p.obj.InitContainers[:loc], p.obj.InitContainers[loc+1:]...)
	return nil
}

func (p *Pod) findContainerPos(name string, containers []types.Container) int {
	for i, c := range containers {
		if strings.Compare(c.Name, name) == 0 {
			return i
		}
	}

	return -1
}

func (p *Pod) addContainerDefaults(c *types.Container) {
	if c.ForceNonRoot == nil && p.defaultForceNonRoot != nil {
		c.ForceNonRoot = p.defaultForceNonRoot
	}

	if c.UID == nil && p.defaultRunUID != nil {
		c.UID = p.defaultRunUID
	}

	if c.GID == nil && p.defaultRunGID != nil {
		c.GID = p.defaultRunGID
	}

	if c.SELinux == nil && p.defaultSELinux != nil {
		c.SELinux = createSELinuxObj(*p.defaultSELinux)
	}
}

// AddVolume adds a volume to the pod
func (p *Pod) AddVolume(obj *Volume) error {
	if _, exists := p.obj.Volumes[obj.Name]; exists {
		return fmt.Errorf("volume %s already exists", obj.Name)
	}
	if p.obj.Volumes == nil {
		p.obj.Volumes = make(map[string]types.Volume)
	}
	p.obj.Volumes[obj.Name] = *obj.GetObj()
	return nil
}

// RemoveVolume removes a volume from the pod
func (p *Pod) RemoveVolume(name string) {
	delete(p.obj.Volumes, name)
}

// AddAffinity adds an affinity configuration to the pod
func (p *Pod) AddAffinity(config api.AffinityConfig) {
	a := types.Affinity{
		NodeAffinity:    config.NodeAffinity,
		PodAffinity:     config.PodAffinity,
		PodAntiAffinity: config.PodAntiAffinity,
		Topology:        config.Topology,
		Namespaces:      config.Namespaces,
	}
	p.obj.Affinity = append(p.obj.Affinity, a)
}

// RemoveAffinity removes an affinity configuration from the pod
func (p *Pod) RemoveAffinity(config api.AffinityConfig) {
	for l, a := range p.obj.Affinity {
		if strings.Compare(a.NodeAffinity, config.NodeAffinity) == 0 ||
			strings.Compare(a.PodAffinity, config.PodAffinity) == 0 ||
			strings.Compare(a.PodAntiAffinity, config.PodAntiAffinity) == 0 ||
			strings.Compare(a.Topology, config.Topology) == 0 {
			p.obj.Affinity = append(p.obj.Affinity[:l], p.obj.Affinity[l+1:]...)
			break
		}
	}
}

// AddHostModes will add a networking host mode to the pod
func (p *Pod) AddHostModes(config []api.HostModeType) {
	var mode types.HostMode

	for _, h := range config {
		switch h {
		case api.HostModeNet:
			mode = types.HostModeNet
		case api.HostModePID:
			mode = types.HostModePID
		case api.HostModeIPC:
			mode = types.HostModeIPC
		}

		found := false
		for _, e := range p.obj.HostMode {
			if strings.Compare(string(mode), string(e)) == 0 {
				found = true
			}
		}
		if !found {
			p.obj.HostMode = append(p.obj.HostMode, mode)
		}
	}

}

// AddSupplementalGIDs will add supplemental GIDs to the pod
func (p *Pod) AddSupplementalGIDs(new []int64) {
	p.obj.GIDs = append(p.obj.GIDs, new...)
}

// RemoveSupplementalGID will remove the provided GID from
// the list of supplemental GIDs on the pod
func (p *Pod) RemoveSupplementalGID(remove int64) {
	for l, g := range p.obj.GIDs {
		if g == remove {
			p.obj.GIDs = append(p.obj.GIDs[:l], p.obj.GIDs[l+1:]...)
			break
		}
	}
}

// AddImagePullSecrets will add image pull secrets to the pod
func (p *Pod) AddImagePullSecrets(new []string) {
	p.obj.Registries = append(p.obj.Registries, new...)
}

// RemoveImagePullSecret will remove an image pull secret from the pod
func (p *Pod) RemoveImagePullSecret(remove string) {
	for l, s := range p.obj.Registries {
		if strings.Compare(s, remove) == 0 {
			p.obj.Registries = append(p.obj.Registries[:l], p.obj.Registries[l+1:]...)
			break
		}
	}
}

// AddHostAliases will add host aliases to the pod
func (p *Pod) AddHostAliases(new []string) {
	p.obj.HostAliases = append(p.obj.HostAliases, new...)
}

// RemoveHostAlias will remove a host alias from the pod
func (p *Pod) RemoveHostAlias(remove string) {
	for l, h := range p.obj.HostAliases {
		if strings.Compare(h, remove) == 0 {
			p.obj.Registries = append(p.obj.HostAliases[:l], p.obj.HostAliases[l+1:]...)
			break
		}
	}
}

// AddTolerations will add tolerations to the pod
func (p *Pod) AddTolerations(config []api.TolerationConfig) {
	for _, c := range config {
		t := types.Toleration{
			ExpiryAfter: c.Expires,
			Selector:    p.createToleration(c),
		}

		p.obj.Tolerations = append(p.obj.Tolerations, t)
	}
}

// RemoveToleration will remove a toleration from the pod
func (p *Pod) RemoveToleration(remove api.TolerationConfig) {
	rt := types.Toleration{
		ExpiryAfter: remove.Expires,
		Selector:    p.createToleration(remove),
	}
	for l, t := range p.obj.Tolerations {
		if reflect.DeepEqual(t, rt) {
			p.obj.Tolerations = append(p.obj.Tolerations[:l], p.obj.Tolerations[l+1:]...)
			break
		}
	}
}

func (p *Pod) createToleration(config api.TolerationConfig) types.Selector {
	var selector string

	switch config.Op {
	case api.TolerationOpExists:
		selector = config.Key
	case api.TolerationOpEqual:
		selector = fmt.Sprintf("%s=%s", config.Key, config.Value)
	}

	switch config.Effect {
	case api.TolerationEffectNoSchedule:
		selector += ":NoSchedule"
	case api.TolerationEffectPreferNoSchedule:
		selector += ":PreferNoSchedule"
	case api.TolerationEffectNoExecute:
		selector += ":NoExecute"
	}

	return types.Selector(selector)
}

// ToKube returns the kubernetes version of the pod
func (p *Pod) ToKube() (runtime.Object, error) {
	wrapper := &types.PodWrapper{Pod: *p.obj}
	return converters.Convert_Koki_Pod_to_Kube_v1_Pod(wrapper)
}
