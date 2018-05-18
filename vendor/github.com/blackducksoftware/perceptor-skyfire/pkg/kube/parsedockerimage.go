/*
Copyright (C) 2018 Black Duck Software, Inc.

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

package kube

import (
	"fmt"
	"regexp"
)

var digestRegexp = regexp.MustCompile("^docker-pullable://(.+)@sha256:([a-zA-Z0-9]+)$")

// ParseImageIDString parses an ImageID pulled from kubernetes.
// Example image id:
//   docker-pullable://registry.kipp.blackducksoftware.com/blackducksoftware/hub-registration@sha256:cb4983d8399a59bb5ee6e68b6177d878966a8fe41abe18a45c3b1d8809f1d043
func ParseImageIDString(imageID string) (string, string, error) {
	match := digestRegexp.FindStringSubmatch(imageID)
	if len(match) != 3 {
		return "", "", fmt.Errorf("unable to match digestRegexp regex <%s> to input <%s>", digestRegexp.String(), imageID)
	}
	name := match[1]
	digest := match[2]
	return name, digest, nil
}
