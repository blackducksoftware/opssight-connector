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

// PodAffinityConfig defines the configuration for a pod affinity or antiaffinity
type PodAffinityConfig struct {
	Weight     int32
	Selector   SelectorConfig
	Topology   string
	Namespaces []string
}

// NodeAffinityConfig defines the configuration for a node affinity
type NodeAffinityConfig struct {
	Weight      int32
	Expressions []NodeExpression
	Fields      []NodeExpression
}

// AffinityType defines the type of affinity
type AffinityType int

const (
	AffinityHard AffinityType = iota + 1
	AffinitySoft
)

// NodeExpression defines the configuration for a node expresion
type NodeExpression struct {
	Key    string
	Op     NodeOperator
	Values []string
}

// NodeOperator defines the valid operators in a node expression
type NodeOperator int

const (
	NodeOperatorIn NodeOperator = iota + 1
	NodeOperatorNotIn
	NodeOperatorExists
	NodeOperatorDoesNotExist
	NodeOperatorGt
	NodeOperatorLt
)
