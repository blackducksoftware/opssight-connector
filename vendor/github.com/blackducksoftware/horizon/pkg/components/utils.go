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

	"k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/util/intstr"
)

func createSELinux(config *api.SELinuxType) *v1.SELinuxOptions {
	if config == nil {
		return nil
	}

	return &v1.SELinuxOptions{
		Level: config.Level,
		Role:  config.Role,
		Type:  config.Type,
		User:  config.User,
	}
}

func createIntOrStr(input string) *intstr.IntOrString {
	var val intstr.IntOrString

	if len(input) == 0 {
		return nil
	}

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

func appendStringIfMissing(new string, list []string) []string {
	for _, o := range list {
		if strings.EqualFold(new, o) {
			return list
		}
	}
	return append(list, new)
}

func appendint64IfMissing(new int64, list []int64) []int64 {
	for _, o := range list {
		if new == o {
			return list
		}
	}
	return append(list, new)
}

func convertProtocol(orig api.ProtocolType) v1.Protocol {
	switch orig {
	case api.ProtocolTCP:
		return v1.ProtocolTCP
	case api.ProtocolUDP:
		return v1.ProtocolUDP
	case api.ProtocolSCTP:
		return v1.ProtocolSCTP
	}

	return v1.Protocol("")
}
