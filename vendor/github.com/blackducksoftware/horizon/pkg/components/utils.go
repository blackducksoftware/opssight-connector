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

package components

import (
	"strconv"
	"strings"

	"github.com/blackducksoftware/horizon/pkg/api"

	"github.com/koki/short/types"

	"k8s.io/apimachinery/pkg/util/intstr"
)

func createSELinuxObj(config api.SELinuxType) *types.SELinux {
	s := types.SELinux{
		Level: config.Level,
		Role:  config.Role,
		Type:  config.Type,
		User:  config.User,
	}

	return &s
}

func createIntOrStr(input string) *intstr.IntOrString {
	var val intstr.IntOrString
	intVal, err := strconv.ParseInt(input, 10, 32)

	if err == nil {
		val = intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: int32(intVal),
		}
	} else {
		val = intstr.IntOrString{
			Type:   intstr.String,
			StrVal: input,
		}
	}

	return &val
}

func appendIfMissing(new string, list []string) []string {
	for _, o := range list {
		if strings.Compare(new, o) == 0 {
			return list
		}
	}
	return append(list, new)
}
