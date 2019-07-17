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

// ClusterRoleConfig defines the basic configuration for a cluster role
type ClusterRoleConfig struct {
	APIVersion  string
	ClusterName string
	Name        string
	Namespace   string
}

// ClusterRoleBindingConfig defines the base configuration for a
// cluster role binding
type ClusterRoleBindingConfig struct {
	APIVersion  string
	ClusterName string
	Name        string
	Namespace   string
}

// PolicyRuleConfig defines the configuration for a policy rule
type PolicyRuleConfig struct {
	Verbs           []string
	APIGroups       []string
	Resources       []string
	ResourceNames   []string
	NonResourceURLs []string
}

// SubjectConfig defines the configuration for a subject
type SubjectConfig struct {
	Kind      string
	APIGroup  string
	Name      string
	Namespace string
}

// RoleRefConfig defines the configuration for a role reference
type RoleRefConfig struct {
	APIGroup string
	Kind     string
	Name     string
}
