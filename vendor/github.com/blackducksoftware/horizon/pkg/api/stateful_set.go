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

// StatefulSetConfig defines the basic configuration for a stateful set
type StatefulSetConfig struct {
	APIVersion           string
	ClusterName          string
	Name                 string
	Namespace            string
	Replicas             *int32
	UpdateStrategy       StatefulSetUpdateStrategyType
	Partition            *int32
	RevisionHistoryLimit *int32
	PodManagementPolicy  PodManagementPolicyType
	Service              string
}

// StatefulSetUpdateStrategyType defines the update strategy for the stateful set
type StatefulSetUpdateStrategyType int

const (
	StatefulSetUpdateStrategyRollingUpdate StatefulSetUpdateStrategyType = iota + 1
	StatefulSetUpdateStrategyOnDelete
)

// PodManagementPolicyType defines the pod managagement policy for the stateful set
type PodManagementPolicyType int

const (
	PodManagementPolicyOrdered PodManagementPolicyType = iota + 1
	PodManagementPolicyParallel
)
