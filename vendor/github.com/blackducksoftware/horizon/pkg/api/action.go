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

// ActionType defines the type of action
type ActionType int

const (
	ActionTypeCommand ActionType = iota + 1
	ActionTypeHTTP
	ActionTypeHTTPS
	ActionTypeTCP
)

// ActionConfig defines the configuration for an action (ie on-start, pre-stop)
type ActionConfig struct {
	Type    ActionType
	Command []string
	Headers map[string]string
	Host    string
	Port    string
	Path    string
}
