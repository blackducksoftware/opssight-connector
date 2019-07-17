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
	"github.com/blackducksoftware/horizon/pkg/util"

	"github.com/koki/short/converter/converters"
	"github.com/koki/short/types"

	"k8s.io/apimachinery/pkg/runtime"
)

// ServiceAccount defines the service account component
type ServiceAccount struct {
	obj *types.ServiceAccount
}

// NewServiceAccount creates a ServiceAccount object
func NewServiceAccount(config api.ServiceAccountConfig) *ServiceAccount {
	sa := &types.ServiceAccount{
		Version:                      config.APIVersion,
		Cluster:                      config.ClusterName,
		Name:                         config.Name,
		Namespace:                    config.Namespace,
		AutomountServiceAccountToken: config.AutomountToken,
	}

	return &ServiceAccount{obj: sa}
}

// GetObj returns the service account object in a format the deployer can use
func (sa *ServiceAccount) GetObj() *types.ServiceAccount {
	return sa.obj
}

// GetName returns the name of the service account
func (sa *ServiceAccount) GetName() string {
	return sa.obj.Name
}

// AddAnnotations adds annotations to the service account
func (sa *ServiceAccount) AddAnnotations(new map[string]string) {
	sa.obj.Annotations = util.MapMerge(sa.obj.Annotations, new)
}

// RemoveAnnotations removes annotations from the service account
func (sa *ServiceAccount) RemoveAnnotations(remove []string) {
	for _, k := range remove {
		sa.obj.Annotations = util.RemoveElement(sa.obj.Annotations, k)
	}
}

// AddLabels adds labels to the service account
func (sa *ServiceAccount) AddLabels(new map[string]string) {
	sa.obj.Labels = util.MapMerge(sa.obj.Labels, new)
}

// RemoveLabels removes labels from the service account
func (sa *ServiceAccount) RemoveLabels(remove []string) {
	for _, k := range remove {
		sa.obj.Labels = util.RemoveElement(sa.obj.Labels, k)
	}
}

// AddPullSecrets adds pull secrets to the service account
func (sa *ServiceAccount) AddPullSecrets(add []string) {
	sa.obj.ImagePullSecrets = append(sa.obj.ImagePullSecrets, add...)
}

// RemovePullSecret removes a pull secret from the service account
func (sa *ServiceAccount) RemovePullSecret(remove string) {
	for l, ps := range sa.obj.ImagePullSecrets {
		if strings.Compare(ps, remove) == 0 {
			sa.obj.ImagePullSecrets = append(sa.obj.ImagePullSecrets[:l], sa.obj.ImagePullSecrets[l+1:]...)
		}
	}
}

// AddObjectReference adds an object reference to the service account
func (sa *ServiceAccount) AddObjectReference(config api.ObjectReferenceConfig) {
	secret := types.ObjectReference{
		Kind:            config.Kind,
		Namespace:       config.Namespace,
		Name:            config.Name,
		UID:             config.UID,
		Version:         config.Version,
		ResourceVersion: config.ResourceVersion,
		FieldPath:       config.FieldPath,
	}
	sa.obj.Secrets = append(sa.obj.Secrets, secret)
}

// RemoveObjectReference will remove an object reference from a service account
func (sa *ServiceAccount) RemoveObjectReference(config api.ObjectReferenceConfig) {
	secret := types.ObjectReference{
		Kind:            config.Kind,
		Namespace:       config.Namespace,
		Name:            config.Name,
		UID:             config.UID,
		Version:         config.Version,
		ResourceVersion: config.ResourceVersion,
		FieldPath:       config.FieldPath,
	}

	for l, s := range sa.obj.Secrets {
		if reflect.DeepEqual(s, secret) {
			sa.obj.Secrets = append(sa.obj.Secrets[:l], sa.obj.Secrets[l+1:]...)
			break
		}
	}
}

// ToKube returns the kubernetes version of the service account
func (sa *ServiceAccount) ToKube() (runtime.Object, error) {
	wrapper := &types.ServiceAccountWrapper{ServiceAccount: *sa.obj}
	return converters.Convert_Koki_ServiceAccount_to_Kube_ServiceAccount(wrapper)
}
