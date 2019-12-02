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

// SelectorConfig defines the configuration for a selector
type SelectorConfig struct {
	Labels      map[string]string
	Expressions []ExpressionRequirementConfig
}

// ExpressionRequirementConfig defines the configuration for an expression
type ExpressionRequirementConfig struct {
	Key    string
	Op     ExpressionRequirementOp
	Values []string
}

// A ExpressionRequirementOp is the set of operators that can be used in a selector requirement
type ExpressionRequirementOp int

const (
	ExpressionRequirementOpIn ExpressionRequirementOp = iota + 1
	ExpressionRequirementOpNotIn
	ExpressionRequirementOpExists
	ExpressionRequirementOpDoesNotExist
)
