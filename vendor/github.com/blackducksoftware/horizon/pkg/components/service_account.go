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
	"reflect"
	"strings"

	"github.com/blackducksoftware/horizon/pkg/api"

	"k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// ServiceAccount defines the service account component
type ServiceAccount struct {
	*v1.ServiceAccount
	MetadataFuncs
}

// NewServiceAccount creates a ServiceAccount object
func NewServiceAccount(config api.ServiceAccountConfig) *ServiceAccount {
	version := "v1"
	if len(config.APIVersion) > 0 {
		version = config.APIVersion
	}

	sa := v1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: version,
		},
		ObjectMeta:                   generateObjectMeta(config.Name, config.Namespace, config.ClusterName),
		AutomountServiceAccountToken: config.AutomountToken,
	}

	return &ServiceAccount{&sa, MetadataFuncs{&sa}}
}

// AddPullSecrets adds pull secrets to the service account
func (sa *ServiceAccount) AddPullSecrets(add []string) {
	for _, name := range add {
		sa.ImagePullSecrets = append(sa.ImagePullSecrets, v1.LocalObjectReference{Name: name})
	}
}

// RemovePullSecret removes a pull secret from the service account
func (sa *ServiceAccount) RemovePullSecret(remove string) {
	for l, ps := range sa.ImagePullSecrets {
		if strings.Compare(ps.Name, remove) == 0 {
			sa.ImagePullSecrets = append(sa.ImagePullSecrets[:l], sa.ImagePullSecrets[l+1:]...)
		}
	}
}

// AddSecret adds a usable secret to the service account
func (sa *ServiceAccount) AddSecret(config api.ServiceAccountSecretConfig) {
	secret := v1.ObjectReference{
		Kind:            config.Kind,
		Namespace:       config.Namespace,
		Name:            config.Name,
		UID:             types.UID(config.UID),
		APIVersion:      config.Version,
		ResourceVersion: config.ResourceVersion,
		FieldPath:       config.FieldPath,
	}
	sa.Secrets = append(sa.Secrets, secret)
}

// RemoveObjectReference will remove a usable secret from a service account
func (sa *ServiceAccount) RemoveObjectReference(config api.ServiceAccountSecretConfig) {
	secret := v1.ObjectReference{
		Kind:            config.Kind,
		Namespace:       config.Namespace,
		Name:            config.Name,
		UID:             types.UID(config.UID),
		APIVersion:      config.Version,
		ResourceVersion: config.ResourceVersion,
		FieldPath:       config.FieldPath,
	}

	for l, s := range sa.Secrets {
		if reflect.DeepEqual(s, secret) {
			sa.Secrets = append(sa.Secrets[:l], sa.Secrets[l+1:]...)
			break
		}
	}
}

// Deploy will deploy the service account to the cluster
func (sa *ServiceAccount) Deploy(res api.DeployerResources) error {
	_, err := res.KubeClient.CoreV1().ServiceAccounts(sa.Namespace).Create(sa.ServiceAccount)
	return err
}

// Undeploy will remove the service account from the cluster
func (sa *ServiceAccount) Undeploy(res api.DeployerResources) error {
	return res.KubeClient.CoreV1().ServiceAccounts(sa.Namespace).Delete(sa.Name, &metav1.DeleteOptions{})
}
