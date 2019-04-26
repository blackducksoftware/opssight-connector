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

package util

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

// MapErrors defines an object that contains errors across one or more components
type MapErrors interface {
	error
	Errors() map[string]error
}

type mapErrors map[string]error

// NewMapErrors converts a map of string:error to a MapErrors interface
func NewMapErrors(errMap map[string]error) MapErrors {
	if len(errMap) == 0 {
		return nil
	}

	return mapErrors(errMap)
}

// Error implements the error interface Error function
func (d mapErrors) Error() string {
	strErrors := ""
	for k, v := range d {
		if len(strErrors) > 0 {
			strErrors += "; "
		}
		log.Debugf("k: %s, v: %+v", k, v)
		strErrors += fmt.Sprintf("%s: %+v", k, v)
	}
	return strErrors
}

// Errors implements the error interface Errors function
func (d mapErrors) Errors() map[string]error {
	return d
}
