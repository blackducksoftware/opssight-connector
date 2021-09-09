/*
Copyright (C) 2019 Synopsys, Inc.

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

// IngressConfig defines the basic configuration for an ingress
type IngressConfig struct {
	APIVersion  string
	ClusterName string
	Name        string
	Namespace   string
	ServiceName string
	ServicePort string
}

// IngressTLSConfig defines the configuration for TLS for the ingress
type IngressTLSConfig struct {
	Hosts      []string
	SecretName string
}

// IngressHostRuleConfig defines the rules mapping the paths under a specified host to
// the related backend services
type IngressHostRuleConfig struct {
	Host  string
	Paths []HTTPIngressPathConfig
}

// HTTPIngressPathConfig defines a collection of paths that map requests to backends
type HTTPIngressPathConfig struct {
	Path        string
	ServiceName string
	ServicePort string
}
