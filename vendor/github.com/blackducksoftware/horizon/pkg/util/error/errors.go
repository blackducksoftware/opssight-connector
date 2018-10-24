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

package error

import (
	"fmt"

	"github.com/blackducksoftware/horizon/pkg/util"
)

// DeployErrors defines an object that contains multiple
// errors across one or more components
type DeployErrors interface {
	error
	Errors() map[util.ComponentType][]error
}

type deployErrors map[util.ComponentType][]error

// NewDeployErrors converts a map of ComponentType:[]error to
// a DeployErrors interface
func NewDeployErrors(errMap map[util.ComponentType][]error) DeployErrors {
	if len(errMap) == 0 {
		return nil
	}

	errs := make(map[util.ComponentType][]error)
	for k, v := range errMap {
		errs[k] = v
	}
	return deployErrors(errs)
}

// Error implements the error interface Error function
func (d deployErrors) Error() string {
	strErrors := ""
	for k, v := range d {
		if len(strErrors) > 0 {
			strErrors += "; "
		}
		strErrors += fmt.Sprintf("%s: %s", k, v[0].Error())
		for i := 1; i < len(v); i++ {
			strErrors += fmt.Sprintf(", %s", v[i].Error())
		}
	}
	return strErrors
}

func (d deployErrors) Errors() map[util.ComponentType][]error {
	return d
}

func ComponentErrorCount(errs error, component util.ComponentType) int {
	if errs == nil {
		return 0
	}
	if dErrs, ok := errs.(DeployErrors); ok {
		return len(dErrs.Errors()[component])
	}
	return -1
}
