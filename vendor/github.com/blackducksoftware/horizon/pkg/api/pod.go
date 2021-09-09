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
	MountSAToken           *bool
	ShareNamespace         *bool
	RuntimeClass           *string
	ServiceLinks           *bool
}

// DNSPolicyType defines the pod DNS policy
type DNSPolicyType int

const (
	DNSClusterFirstWithHostNet DNSPolicyType = iota + 1
	DNSClusterFirst
	DNSDefault
)

// RestartPolicyType defines the pod restart policy
type RestartPolicyType int

const (
	RestartPolicyAlways RestartPolicyType = iota + 1
	RestartPolicyOnFailure
	RestartPolicyNever
)

// HostModeType defines the host mode for the pod
type HostModeType int

const (
	HostModeNet HostModeType = iota + 1
	HostModePID
	HostModeIPC
)

// TolerationConfig defines the configuration for a pod toleration
type TolerationConfig struct {
	Duration *int64
	Effect   TolerationEffectType
	Key      string
	Value    string
	Op       TolerationOpType
}

// TolerationEffectType defines the effect of the toleration
type TolerationEffectType int

const (
	TolerationEffectNoSchedule TolerationEffectType = iota + 1
	TolerationEffectPreferNoSchedule
	TolerationEffectNoExecute
)

// TolerationOpType defines the toleration operator
type TolerationOpType int

const (
	TolerationOpExists TolerationOpType = iota + 1
	TolerationOpEqual
)

// PodDNSConfig defines the dns configuration for a pod
type PodDNSConfig struct {
	Nameservers     []string
	SearchDomains   []string
	ResolverOptions map[string]string
}

// HostAliasConfig defines the configuration for a host alias on a pod
type HostAliasConfig struct {
	IP        string
	Hostnames []string
}
