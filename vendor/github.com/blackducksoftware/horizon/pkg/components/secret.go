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
	"github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/util"

	"github.com/koki/short/converter/converters"
	"github.com/koki/short/types"

	"k8s.io/apimachinery/pkg/runtime"
)

// Secret defines the secret component
type Secret struct {
	obj *types.Secret
}

// NewSecret creates a Secret object
func NewSecret(config api.SecretConfig) *Secret {
	s := &types.Secret{
		Version:   config.APIVersion,
		Cluster:   config.ClusterName,
		Name:      config.Name,
		Namespace: config.Namespace,
	}

	switch config.Type {
	case api.SecretTypeOpaque:
		s.SecretType = types.SecretTypeOpaque
	case api.SecretTypeServiceAccountToken:
		s.SecretType = types.SecretTypeServiceAccountToken
	case api.SecretTypeDockercfg:
		s.SecretType = types.SecretTypeDockercfg
	case api.SecretTypeDockerConfigJSON:
		s.SecretType = types.SecretTypeDockerConfigJson
	case api.SecretTypeBasicAuth:
		s.SecretType = types.SecretTypeBasicAuth
	case api.SecretTypeSSHAuth:
		s.SecretType = types.SecretTypeSSHAuth
	case api.SecretTypeTLS:
		s.SecretType = types.SecretTypeTLS
	}

	return &Secret{obj: s}
}

// GetObj returns the secret object in a format the deployer can use
func (s *Secret) GetObj() *types.Secret {
	return s.obj
}

// GetName returns the name of the secret
func (s *Secret) GetName() string {
	return s.obj.Name
}

// AddAnnotations adds annotations to the secret
func (s *Secret) AddAnnotations(new map[string]string) {
	s.obj.Annotations = util.MapMerge(s.obj.Annotations, new)
}

// RemoveAnnotations removes annotations from the secret
func (s *Secret) RemoveAnnotations(remove []string) {
	for _, k := range remove {
		s.obj.Annotations = util.RemoveElement(s.obj.Annotations, k)
	}
}

// AddLabels adds labels to the secret
func (s *Secret) AddLabels(new map[string]string) {
	s.obj.Labels = util.MapMerge(s.obj.Labels, new)
}

// RemoveLabels removes labels from the secret
func (s *Secret) RemoveLabels(remove []string) {
	for _, k := range remove {
		s.obj.Labels = util.RemoveElement(s.obj.Labels, k)
	}
}

// AddStringData adds string data to the secret
func (s *Secret) AddStringData(new map[string]string) {
	if !(len(s.obj.StringData) > 0) {
		s.obj.StringData = make(map[string]string)
	}
	s.obj.StringData = util.MapMerge(s.obj.StringData, new)
}

// RemoveStringData removes string data from the secret
func (s *Secret) RemoveStringData(remove []string) {
	for _, k := range remove {
		s.obj.StringData = util.RemoveElement(s.obj.StringData, k)
	}
}

// AddData adds data to the secret
func (s *Secret) AddData(new map[string][]byte) {
	if !(len(s.obj.Data) > 0) {
		s.obj.Data = make(map[string][]byte)
	}
	for k, v := range new {
		s.obj.Data[k] = v
	}
}

// RemoveData removes data from the secret
func (s *Secret) RemoveData(remove []string) {
	for _, k := range remove {
		if _, exists := s.obj.Data[k]; exists {
			delete(s.obj.Data, k)
		}
	}
}

// ToKube returns the kubernetes version of the secret
func (s *Secret) ToKube() (runtime.Object, error) {
	wrapper := &types.SecretWrapper{Secret: *s.obj}
	return converters.Convert_Koki_Secret_to_Kube_v1_Secret(wrapper)
}
