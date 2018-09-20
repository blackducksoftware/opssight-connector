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

// ContainerConfig defines the basic configuration for a container
type ContainerConfig struct {
	Name                     string
	Args                     []string
	Command                  []string
	Image                    string
	PullPolicy               PullPolicyType
	MinCPU                   string
	MaxCPU                   string
	MinMem                   string
	MaxMem                   string
	Privileged               *bool
	AllowPrivilegeEscalation *bool
	ReadOnlyFS               *bool
	ForceNonRoot             *bool
	SELinux                  *SELinuxType
	UID                      *int64
	AllocateStdin            bool
	StdinOnce                bool
	AllocateTTY              bool
	WorkingDirectory         string
	TerminationMsgPath       string
	TerminationMsgPolicy     TerminationMessagePolicyType
}

// PullPolicyType defines the type of pull policy
type PullPolicyType int

const (
	PullAlways PullPolicyType = iota
	PullNever
	PullIfNotPresent
)

// PortConfig defines the configuration for a port
type PortConfig struct {
	Name          string
	Protocol      ProtocolType
	IP            string
	HostPort      string
	ContainerPort string
}

// ProtocolType defines the type of protocols
type ProtocolType int

const (
	ProtocolTCP ProtocolType = iota
	ProtocolUDP
)

// VolumeMountConfig defines the configuration for a volume mount
type VolumeMountConfig struct {
	MountPath   string
	Propagation MountPropagationType
	Name        string
	SubPath     string
	ReadOnly    *bool
}

// MountPropagationType defines the type of mount propagation
// for the volume mount
type MountPropagationType string

const (
	MountPropagationHostToContainer MountPropagationType = "host-to-container"
	MountPropagationBidirectional   MountPropagationType = "bidirectional"
	MountPropagationNone            MountPropagationType = "none"
)

// ProbeConfig defines the configuration for a probe
type ProbeConfig struct {
	ActionConfig
	Delay           int32
	Interval        int32
	MinCountSuccess int32
	MinCountFailure int32
	Timeout         int32
}

// TerminationMessagePolicyType defines the policy for the termination message
type TerminationMessagePolicyType int

const (
	TerminationMessageReadFile TerminationMessagePolicyType = iota
	TerminationMessageFallbackToLogsOnError
)
