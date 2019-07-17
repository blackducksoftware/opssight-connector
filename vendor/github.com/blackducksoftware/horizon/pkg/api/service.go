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

// ServiceConfig defines the basic configuration for a service
type ServiceConfig struct {
	APIVersion               string
	ClusterName              string
	Name                     string
	Namespace                string
	ExternalName             string
	Type                     ServiceType
	ClusterIP                string
	PublishNotReadyAddresses bool
	TrafficPolicy            TrafficPolicyType
	Affinity                 ServiceAffinityType
	IPTimeout                *int32
}

// ServiceType defines how the service is exposed service
type ServiceType int

const (
	ServiceTypeServiceIP ServiceType = iota + 1
	ServiceTypeNodePort
	ServiceTypeLoadBalancer
	ServiceTypeExternalName
)

// ServicePortConfig defines the configuration for a service port
type ServicePortConfig struct {
	Name       string
	Port       int32
	TargetPort string
	NodePort   int32
	Protocol   ProtocolType
}

// TrafficPolicyType defines the external traffic policy for the service
type TrafficPolicyType int

const (
	ServiceTrafficPolicyLocal TrafficPolicyType = iota + 1
	ServiceTrafficPolicyCluster
)

// LoadBalancerConfig defines the configuration for a load balancer
// to use with a service
type LoadBalancerConfig struct {
	IP                  string
	AllowedIPs          []string
	HealthCheckNodePort int32
}

type ServiceAffinityType int

const (
	ServiceAffinityTypeClientIP ServiceAffinityType = iota + 1
	ServiceAffinityTypeNone
)
