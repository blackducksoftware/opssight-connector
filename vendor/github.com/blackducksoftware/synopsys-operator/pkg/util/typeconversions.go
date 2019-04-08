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

package util

import (
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	corev1 "k8s.io/api/core/v1"
)

// KubeSecretTypeToHorizon converts a kubernetes SecretType to Horizon's SecretType
func KubeSecretTypeToHorizon(secretType corev1.SecretType) (horizonapi.SecretType, error) {
	switch secretType {
	case corev1.SecretTypeOpaque:
		return horizonapi.SecretTypeOpaque, nil
	case corev1.SecretTypeServiceAccountToken:
		return horizonapi.SecretTypeServiceAccountToken, nil
	case corev1.SecretTypeDockercfg:
		return horizonapi.SecretTypeDockercfg, nil
	case corev1.SecretTypeDockerConfigJson:
		return horizonapi.SecretTypeDockerConfigJSON, nil
	case corev1.SecretTypeBasicAuth:
		return horizonapi.SecretTypeBasicAuth, nil
	case corev1.SecretTypeSSHAuth:
		return horizonapi.SecretTypeSSHAuth, nil
	case corev1.SecretTypeTLS:
		return horizonapi.SecretTypeTLS, nil
	default:
		return horizonapi.SecretTypeOpaque, fmt.Errorf("Invalid Secret Type: %+v", secretType)
	}
}

// SecretTypeNameToHorizon converts a SecretType Name String to Horizon's SecretType
func SecretTypeNameToHorizon(secretType string) (horizonapi.SecretType, error) {
	switch secretType {
	case "Opaque":
		return horizonapi.SecretTypeOpaque, nil
	case "ServiceAccountToken":
		return horizonapi.SecretTypeServiceAccountToken, nil
	case "Dockercfg":
		return horizonapi.SecretTypeDockercfg, nil
	case "DockerConfigJSON":
		return horizonapi.SecretTypeDockerConfigJSON, nil
	case "BasicAuth":
		return horizonapi.SecretTypeBasicAuth, nil
	case "SSHAuth":
		return horizonapi.SecretTypeSSHAuth, nil
	case "TypeTLS":
		return horizonapi.SecretTypeTLS, nil
	default:
		return horizonapi.SecretTypeOpaque, fmt.Errorf("Invalid Secret Type: %s", secretType)
	}
}
