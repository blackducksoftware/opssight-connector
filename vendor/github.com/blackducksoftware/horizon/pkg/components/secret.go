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

	"k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/imdario/mergo"
)

// Secret defines the secret component
type Secret struct {
	*v1.Secret
	MetadataFuncs
}

// NewSecret creates a Secret object
func NewSecret(config api.SecretConfig) *Secret {
	version := "v1"
	if len(config.APIVersion) > 0 {
		version = config.APIVersion
	}

	s := v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: version,
		},
		ObjectMeta: generateObjectMeta(config.Name, config.Namespace, config.ClusterName),
	}

	switch config.Type {
	case api.SecretTypeOpaque:
		s.Type = v1.SecretTypeOpaque
	case api.SecretTypeServiceAccountToken:
		s.Type = v1.SecretTypeServiceAccountToken
	case api.SecretTypeDockercfg:
		s.Type = v1.SecretTypeDockercfg
	case api.SecretTypeDockerConfigJSON:
		s.Type = v1.SecretTypeDockerConfigJson
	case api.SecretTypeBasicAuth:
		s.Type = v1.SecretTypeBasicAuth
	case api.SecretTypeSSHAuth:
		s.Type = v1.SecretTypeSSHAuth
	case api.SecretTypeTLS:
		s.Type = v1.SecretTypeTLS
	case api.SecretTypeBootstrapToken:
		s.Type = v1.SecretTypeBootstrapToken
	}

	return &Secret{&s, MetadataFuncs{&s}}
}

// AddStringData adds string data to the secret
func (s *Secret) AddStringData(new map[string]string) {
	if !(len(s.StringData) > 0) {
		s.StringData = make(map[string]string)
	}
	mergo.Merge(&s.StringData, new, mergo.WithOverride)
}

// RemoveStringData removes string data from the secret
func (s *Secret) RemoveStringData(remove []string) {
	for _, k := range remove {
		delete(s.StringData, k)
	}
}

// AddData adds data to the secret
func (s *Secret) AddData(new map[string][]byte) {
	if !(len(s.Data) > 0) {
		s.Data = make(map[string][]byte)
	}
	mergo.Merge(&s.Data, new, mergo.WithOverride)
}

// RemoveData removes data from the secret
func (s *Secret) RemoveData(remove []string) {
	for _, k := range remove {
		delete(s.Data, k)
	}
}

// Deploy will deploy the secret to the cluster
func (s *Secret) Deploy(res api.DeployerResources) error {
	_, err := res.KubeClient.CoreV1().Secrets(s.Namespace).Create(s.Secret)
	return err
}

// Undeploy will remove the secret from the cluster
func (s *Secret) Undeploy(res api.DeployerResources) error {
	return res.KubeClient.CoreV1().Secrets(s.Namespace).Delete(s.Name, &metav1.DeleteOptions{})
}
