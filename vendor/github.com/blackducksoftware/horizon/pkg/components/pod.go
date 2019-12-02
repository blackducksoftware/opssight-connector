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
	"github.com/imdario/mergo"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Pod defines the pod component
type Pod struct {
	*v1.Pod
	MetadataFuncs
}

// NewPod create a Pod object
func NewPod(config api.PodConfig) *Pod {
	version := "v1"
	if len(config.APIVersion) > 0 {
		version = config.APIVersion
	}

	p := v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: version,
		},
		ObjectMeta: generateObjectMeta(config.Name, config.Namespace, config.ClusterName),
		Spec: v1.PodSpec{
			TerminationGracePeriodSeconds: config.TerminationGracePeriod,
			ActiveDeadlineSeconds:         config.ActiveDeadline,
			ServiceAccountName:            config.ServiceAccount,
			AutomountServiceAccountToken:  config.MountSAToken,
			NodeName:                      config.Node,
			ShareProcessNamespace:         config.ShareNamespace,
			SecurityContext:               createPodSecurityContext(config),
			SchedulerName:                 config.SchedulerName,
			PriorityClassName:             config.PriorityClass,
			Priority:                      config.PriorityValue,
			RuntimeClassName:              config.RuntimeClass,
			EnableServiceLinks:            config.ServiceLinks,
		},
	}

	fields := strings.SplitN(config.Hostname, ".", 2)
	if len(fields) == 1 {
		p.Spec.Hostname = config.Hostname
	} else {
		p.Spec.Hostname = fields[0]
		p.Spec.Subdomain = fields[1]
	}

	switch config.RestartPolicy {
	case api.RestartPolicyAlways:
		p.Spec.RestartPolicy = v1.RestartPolicyAlways
	case api.RestartPolicyOnFailure:
		p.Spec.RestartPolicy = v1.RestartPolicyOnFailure
	case api.RestartPolicyNever:
		p.Spec.RestartPolicy = v1.RestartPolicyNever
	}

	switch config.DNSPolicy {
	case api.DNSClusterFirstWithHostNet:
		p.Spec.DNSPolicy = v1.DNSClusterFirstWithHostNet
	case api.DNSClusterFirst:
		p.Spec.DNSPolicy = v1.DNSClusterFirst
	case api.DNSDefault:
		p.Spec.DNSPolicy = v1.DNSDefault
	default:
		p.Spec.DNSPolicy = v1.DNSClusterFirst
	}

	return &Pod{&p, MetadataFuncs{&p}}
}

func createPodSecurityContext(config api.PodConfig) *v1.PodSecurityContext {
	psc := &v1.PodSecurityContext{}

	psc.SELinuxOptions = createSELinux(config.SELinux)
	psc.RunAsUser = config.RunAsUser
	psc.RunAsGroup = config.RunAsGroup
	psc.RunAsNonRoot = config.ForceNonRoot
	psc.FSGroup = config.FSGID

	if reflect.DeepEqual(psc, &v1.PodSecurityContext{}) {
		return nil
	}

	return psc
}

// AddContainer adds a container to the pod
func (p *Pod) AddContainer(obj *Container) error {
	if p.findContainerPos(obj.Name, p.Spec.Containers) >= 0 {
		return fmt.Errorf("container with name %s already exists", obj.Name)
	}

	p.Spec.Containers = append(p.Spec.Containers, *obj.Container)

	return nil
}

// RemoveContainer removes a container from the pod
func (p *Pod) RemoveContainer(name string) error {
	loc := p.findContainerPos(name, p.Spec.Containers)
	if loc < 0 {
		return fmt.Errorf("container with name %s doesn't exist", name)
	}

	p.Spec.Containers = append(p.Spec.Containers[:loc], p.Spec.Containers[loc+1:]...)
	return nil
}

// AddInitContainer adds an init container to the pod
func (p *Pod) AddInitContainer(obj *Container) error {
	if p.findContainerPos(obj.Name, p.Spec.InitContainers) >= 0 {
		return fmt.Errorf("init container with name %s already exists", obj.Name)
	}

	p.Spec.InitContainers = append(p.Spec.InitContainers, *obj.Container)
	return nil
}

// RemoveInitContainer removes an init container from the pod
func (p *Pod) RemoveInitContainer(name string) error {
	loc := p.findContainerPos(name, p.Spec.InitContainers)
	if loc < 0 {
		return fmt.Errorf("init container with name %s doesn't exist", name)
	}

	p.Spec.Containers = append(p.Spec.InitContainers[:loc], p.Spec.InitContainers[loc+1:]...)
	return nil
}

func (p *Pod) findContainerPos(name string, containers []v1.Container) int {
	for i, c := range containers {
		if strings.Compare(c.Name, name) == 0 {
			return i
		}
	}

	return -1
}

// AddVolume adds a volume to the pod
func (p *Pod) AddVolume(obj *Volume) error {
	for _, v := range p.Spec.Volumes {
		if strings.EqualFold(v.Name, obj.Name) {
			return fmt.Errorf("volume %s already exists", obj.Name)
		}
	}

	p.Spec.Volumes = append(p.Spec.Volumes, *obj.Volume)
	return nil
}

// RemoveVolume removes a volume from the pod
func (p *Pod) RemoveVolume(name string) {
	for l, v := range p.Spec.Volumes {
		if strings.EqualFold(v.Name, name) {
			p.Spec.Volumes = append(p.Spec.Volumes[:l], p.Spec.Volumes[l+1:]...)
			break
		}
	}
}

// AddNodeAffinity adds a node affinity to the pod
func (p *Pod) AddNodeAffinity(affinityType api.AffinityType, config api.NodeAffinityConfig) error {
	if p.Spec.Affinity == nil {
		p.Spec.Affinity = &v1.Affinity{}
	}

	if p.Spec.Affinity.NodeAffinity == nil {
		p.Spec.Affinity.NodeAffinity = &v1.NodeAffinity{}
	}

	switch affinityType {
	case api.AffinityHard:
		term, err := generateNodeSelectorTerm(config)
		if err != nil {
			return fmt.Errorf("failed to add hard node affinity: %+v", err)
		}

		if p.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
			p.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &v1.NodeSelector{}
		}
		p.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = append(p.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, *term)

	case api.AffinitySoft:
		var term v1.PreferredSchedulingTerm

		nsTerm, err := generateNodeSelectorTerm(config)
		if err != nil {
			return fmt.Errorf("failed to add soft node affinity: %+v", err)
		}
		term.Weight = config.Weight
		term.Preference = *nsTerm

		p.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(p.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution, term)
	}

	return nil
}

// RemoveNodeAffinity removes a node affinity from the pod
func (p *Pod) RemoveNodeAffinity(affinityType api.AffinityType, config api.NodeAffinityConfig) error {
	if p.Spec.Affinity != nil && p.Spec.Affinity.NodeAffinity != nil {
		switch affinityType {
		case api.AffinityHard:
			if p.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
				term, err := generateNodeSelectorTerm(config)
				if err != nil {
					return fmt.Errorf("failed to generate hard node affinity for deletion: %+v", err)
				}

				for l, t := range p.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
					if reflect.DeepEqual(t, *term) {
						p.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = append(p.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[:l], p.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[l+1:]...)
						break
					}
				}
			}

		case api.AffinitySoft:
			var term v1.PreferredSchedulingTerm

			nsTerm, err := generateNodeSelectorTerm(config)
			if err != nil {
				return fmt.Errorf("failed to generate soft node affinity for deletion: %+v", err)
			}
			term.Weight = config.Weight
			term.Preference = *nsTerm

			for l, t := range p.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
				if reflect.DeepEqual(t, term) {
					p.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(p.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[:l], p.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[l+1:]...)
					break
				}
			}
		}
	}

	return nil
}

// AddPodAffinity adds a pod affinity to the pod
func (p *Pod) AddPodAffinity(affinityType api.AffinityType, config api.PodAffinityConfig) {
	if p.Spec.Affinity == nil {
		p.Spec.Affinity = &v1.Affinity{}
	}

	if p.Spec.Affinity.PodAffinity == nil {
		p.Spec.Affinity.PodAffinity = &v1.PodAffinity{}
	}

	switch affinityType {
	case api.AffinityHard:
		term := generatePodAffinityTerm(config)
		p.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(p.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution, term)

	case api.AffinitySoft:
		var term v1.WeightedPodAffinityTerm

		term.PodAffinityTerm = generatePodAffinityTerm(config)
		term.Weight = config.Weight

		p.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(p.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution, term)
	}
}

// RemovePodAffinity removes a pod affinity from the pod
func (p *Pod) RemovePodAffinity(affinityType api.AffinityType, config api.PodAffinityConfig) {
	if p.Spec.Affinity != nil && p.Spec.Affinity.PodAffinity != nil {
		switch affinityType {
		case api.AffinityHard:
			term := generatePodAffinityTerm(config)

			for l, t := range p.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
				if reflect.DeepEqual(t, term) {
					p.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(p.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution[:l], p.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution[l+1:]...)
					break
				}
			}

		case api.AffinitySoft:
			var term v1.WeightedPodAffinityTerm

			term.PodAffinityTerm = generatePodAffinityTerm(config)
			term.Weight = config.Weight

			for l, t := range p.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
				if reflect.DeepEqual(t, term) {
					p.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(p.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution[:l], p.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution[l+1:]...)
					break
				}
			}
		}
	}
}

// AddPodAntiAffinity adds a pod affinity to the pod
func (p *Pod) AddPodAntiAffinity(affinityType api.AffinityType, config api.PodAffinityConfig) {
	if p.Spec.Affinity == nil {
		p.Spec.Affinity = &v1.Affinity{}
	}

	if p.Spec.Affinity.PodAntiAffinity == nil {
		p.Spec.Affinity.PodAntiAffinity = &v1.PodAntiAffinity{}
	}

	switch affinityType {
	case api.AffinityHard:
		term := generatePodAffinityTerm(config)
		p.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(p.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution, term)

	case api.AffinitySoft:
		var term v1.WeightedPodAffinityTerm

		term.PodAffinityTerm = generatePodAffinityTerm(config)
		term.Weight = config.Weight

		p.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(p.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution, term)
	}
}

// RemovePodAntiAffinity removes a pod affinity from the pod
func (p *Pod) RemovePodAntiAffinity(affinityType api.AffinityType, config api.PodAffinityConfig) {
	if p.Spec.Affinity != nil && p.Spec.Affinity.PodAntiAffinity != nil {
		switch affinityType {
		case api.AffinityHard:
			term := generatePodAffinityTerm(config)

			for l, t := range p.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
				if reflect.DeepEqual(t, term) {
					p.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(p.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution[:l], p.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution[l+1:]...)
					break
				}
			}

		case api.AffinitySoft:
			var term v1.WeightedPodAffinityTerm

			term.PodAffinityTerm = generatePodAffinityTerm(config)
			term.Weight = config.Weight

			for l, t := range p.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
				if reflect.DeepEqual(t, term) {
					p.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(p.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[:l], p.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[l+1:]...)
					break
				}
			}
		}
	}
}

// AddHostMode will add a networking host mode to the pod
func (p *Pod) AddHostMode(mode api.HostModeType) {
	switch mode {
	case api.HostModeNet:
		p.Spec.HostNetwork = true
	case api.HostModePID:
		p.Spec.HostPID = true
	case api.HostModeIPC:
		p.Spec.HostIPC = true
	}
}

// RemoveHostMode will remove a networking host mode from the pod
func (p *Pod) RemoveHostMode(mode api.HostModeType) {
	switch mode {
	case api.HostModeNet:
		p.Spec.HostNetwork = false
	case api.HostModePID:
		p.Spec.HostPID = false
	case api.HostModeIPC:
		p.Spec.HostIPC = false
	}
}

// AddSupplementalGIDs will add supplemental GIDs to the pod
func (p *Pod) AddSupplementalGIDs(new []int64) {
	if p.Spec.SecurityContext == nil {
		p.Spec.SecurityContext = &v1.PodSecurityContext{}
	}
	for _, gid := range new {
		p.Spec.SecurityContext.SupplementalGroups = appendint64IfMissing(gid, p.Spec.SecurityContext.SupplementalGroups)
	}
}

// RemoveSupplementalGID will remove the provided GID from
// the list of supplemental GIDs on the pod
func (p *Pod) RemoveSupplementalGID(remove int64) {
	if p.Spec.SecurityContext != nil {
		for l, g := range p.Spec.SecurityContext.SupplementalGroups {
			if g == remove {
				p.Spec.SecurityContext.SupplementalGroups = append(p.Spec.SecurityContext.SupplementalGroups[:l], p.Spec.SecurityContext.SupplementalGroups[l+1:]...)
				break
			}
		}
	}
}

// AddImagePullSecrets will add image pull secrets to the pod
func (p *Pod) AddImagePullSecrets(new []string) {
	for _, name := range new {
		p.Spec.ImagePullSecrets = append(p.Spec.ImagePullSecrets, v1.LocalObjectReference{Name: name})
	}
}

// RemoveImagePullSecret will remove an image pull secret from the pod
func (p *Pod) RemoveImagePullSecret(remove string) {
	for l, s := range p.Spec.ImagePullSecrets {
		if strings.Compare(s.Name, remove) == 0 {
			p.Spec.ImagePullSecrets = append(p.Spec.ImagePullSecrets[:l], p.Spec.ImagePullSecrets[l+1:]...)
			break
		}
	}
}

// AddHostAliases will add host aliases to the pod
func (p *Pod) AddHostAliases(new []api.HostAliasConfig) {
	for _, conf := range new {
		alias := v1.HostAlias{
			IP:        conf.IP,
			Hostnames: conf.Hostnames,
		}
		p.Spec.HostAliases = append(p.Spec.HostAliases, alias)
	}
}

// RemoveHostAlias will remove a host alias from the pod
func (p *Pod) RemoveHostAlias(remove api.HostAliasConfig) {
	alias := v1.HostAlias{
		IP:        remove.IP,
		Hostnames: remove.Hostnames,
	}

	for l, a := range p.Spec.HostAliases {
		if reflect.DeepEqual(a, alias) {
			p.Spec.HostAliases = append(p.Spec.HostAliases[:l], p.Spec.HostAliases[l+1:]...)
			break
		}
	}
}

// AddTolerations will add tolerations to the pod
func (p *Pod) AddTolerations(config []api.TolerationConfig) {
	for _, conf := range config {
		p.Spec.Tolerations = append(p.Spec.Tolerations, p.createToleration(conf))
	}
}

// RemoveToleration will remove a toleration from the pod
func (p *Pod) RemoveToleration(remove api.TolerationConfig) {
	rt := p.createToleration(remove)

	for l, t := range p.Spec.Tolerations {
		if reflect.DeepEqual(t, rt) {
			p.Spec.Tolerations = append(p.Spec.Tolerations[:l], p.Spec.Tolerations[l+1:]...)
			break
		}
	}
}

func (p *Pod) createToleration(config api.TolerationConfig) v1.Toleration {
	toleration := v1.Toleration{
		Key:               config.Key,
		Value:             config.Value,
		TolerationSeconds: config.Duration,
	}

	switch config.Op {
	case api.TolerationOpExists:
		toleration.Operator = v1.TolerationOpExists
	case api.TolerationOpEqual:
		toleration.Operator = v1.TolerationOpEqual
	}

	switch config.Effect {
	case api.TolerationEffectNoSchedule:
		toleration.Effect = v1.TaintEffectNoSchedule
	case api.TolerationEffectPreferNoSchedule:
		toleration.Effect = v1.TaintEffectPreferNoSchedule
	case api.TolerationEffectNoExecute:
		toleration.Effect = v1.TaintEffectNoExecute
	}

	return toleration
}

// AddNodeSelectors will add the node selectors to the pod
func (p *Pod) AddNodeSelectors(selectors map[string]string) {
	mergo.Merge(&p.Spec.NodeSelector, selectors, mergo.WithOverride)
}

// RemoveNodeSelectors will remove the node selectors from the pod
func (p *Pod) RemoveNodeSelectors(selectors []string) {
	for _, s := range selectors {
		delete(p.Spec.NodeSelector, s)
	}
}

// AddDNSConfig will add the DNS configuration to the pod
func (p *Pod) AddDNSConfig(dns api.PodDNSConfig) {
	var opts []v1.PodDNSConfigOption

	for k, v := range dns.ResolverOptions {
		opt := v1.PodDNSConfigOption{
			Name: k,
		}
		if len(v) == 0 {
			opt.Value = &v
		}
		opts = append(opts, opt)
	}

	p.Spec.DNSConfig = &v1.PodDNSConfig{
		Nameservers: dns.Nameservers,
		Searches:    dns.SearchDomains,
		Options:     opts,
	}
}

// RemoveDNSConfig will remote the DNS configuration from the pod
func (p *Pod) RemoveDNSConfig() {
	p.Spec.DNSConfig = nil
}

// AddSysctls will add the sysctl key value pairs to the pod
func (p *Pod) AddSysctls(sysctls map[string]string) {
	if p.Spec.SecurityContext == nil {
		p.Spec.SecurityContext = &v1.PodSecurityContext{}
	}

	for k, v := range sysctls {
		p.Spec.SecurityContext.Sysctls = append(p.Spec.SecurityContext.Sysctls, v1.Sysctl{Name: k, Value: v})
	}
}

// RemoveSysctls will remove the sysctl keys from the pod
func (p *Pod) RemoveSysctls(sysctls []string) {
	if p.Spec.SecurityContext != nil {
		for _, k := range sysctls {
			for l, s := range p.Spec.SecurityContext.Sysctls {
				if strings.EqualFold(s.Name, k) {
					p.Spec.SecurityContext.Sysctls = append(p.Spec.SecurityContext.Sysctls[:l], p.Spec.SecurityContext.Sysctls[l+1:]...)
					break
				}
			}
		}
	}
}

// Deploy will deploy the pod to the cluster
func (p *Pod) Deploy(res api.DeployerResources) error {
	_, err := res.KubeClient.CoreV1().Pods(p.Namespace).Create(p.Pod)
	return err
}

// Undeploy will remove the pod from the cluster
func (p *Pod) Undeploy(res api.DeployerResources) error {
	return res.KubeClient.CoreV1().Pods(p.Namespace).Delete(p.Name, &metav1.DeleteOptions{})
}
