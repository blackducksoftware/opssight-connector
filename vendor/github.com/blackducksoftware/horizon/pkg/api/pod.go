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

package api

// PodConfig defines the basic configuration for a pod
type PodConfig struct {
	APIVersion             string
	ClusterName            string
	Name                   string
	Namespace              string
	ServiceAccount         string
	RestartPolicy          RestartPolicyType
	TerminationGracePeriod *int64
	ActiveDeadline         *int64
	Node                   string
	FSGID                  *int64
	Hostname               string
	SchedulerName          string
	DNSPolicy              DNSPolicyType
	PriorityValue          *int32
	PriorityClass          string
	SELinux                *SELinuxType
	RunAsUser              *int64
	RunAsGroup             *int64
	ForceNonRoot           *bool
}

// AffinityConfig defines the configuration for an affinity
type AffinityConfig struct {
	NodeAffinity    string
	PodAffinity     string
	PodAntiAffinity string
	Topology        string
	Namespaces      []string
}

// DNSPolicyType defines the pod DNS policy
type DNSPolicyType int

const (
	DNSClusterFirstWithHostNet DNSPolicyType = iota
	DNSClusterFirst
	DNSDefault
)

// RestartPolicyType defines the pod restart policy
type RestartPolicyType int

const (
	RestartPolicyAlways RestartPolicyType = iota
	RestartPolicyOnFailure
	RestartPolicyNever
)

// HostModeType defines the host mode for the pod
type HostModeType int

const (
	HostModeNet HostModeType = iota
	HostModePID
	HostModeIPC
)

// TolerationConfig defines the configuration for a pod toleration
type TolerationConfig struct {
	Expires *int64
	Effect  TolerationEffectType
	Key     string
	Value   string
	Op      TolerationOpType
}

// TolerationEffectType defines the effect of the toleration
type TolerationEffectType int

const (
	TolerationEffectNone TolerationEffectType = iota
	TolerationEffectNoSchedule
	TolerationEffectPreferNoSchedule
	TolerationEffectNoExecute
)

// TolerationOpType defines the toleration operator
type TolerationOpType int

const (
	TolerationOpExists TolerationOpType = iota
	TolerationOpEqual
)
