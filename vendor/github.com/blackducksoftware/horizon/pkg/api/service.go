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
	IPServiceType            ClusterIPServiceType
	ClusterIP                string
	PublishNotReadyAddresses bool
	TrafficPolicy            TrafficPolicyType

	// Valid Affinity values are true or a valid integer.
	// If set to true, it configures the session affinity
	// of the service based on client ip addresses.
	// If set to a number then, along with configuring the
	// session affinity, it also configures the number of
	// seconds the session affinity sticks to a client ip address
	// before it expires.
	Affinity string
}

// ClusterIPServiceType defines the type of IP service
type ClusterIPServiceType int

const (
	ClusterIPServiceTypeDefault ClusterIPServiceType = iota
	ClusterIPServiceTypeNodePort
	ClusterIPServiceTypeLoadBalancer
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
	ServiceTrafficPolicyNil TrafficPolicyType = iota
	ServiceTrafficPolicyLocal
	ServiceTrafficPolicyCluster
)

// LoadBalancerConfig defines the configuration for a load balancer
// to use with a service
type LoadBalancerConfig struct {
	IP                  string
	AllowedIPs          []string
	HealthCheckNodePort int32
}
