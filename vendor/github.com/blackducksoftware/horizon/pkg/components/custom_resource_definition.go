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

	"github.com/blackducksoftware/horizon/pkg/api"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CustomResourceDefinition defines a custom resource
type CustomResourceDefinition struct {
	*v1beta1.CustomResourceDefinition
	MetadataFuncs
}

// NewCustomResourceDefintion returns a CustomerResrouceDefinition object
func NewCustomResourceDefintion(config api.CRDConfig) *CustomResourceDefinition {
	version := "apiextensions/v1beta1"
	if len(config.APIVersion) > 0 {
		version = config.APIVersion
	}

	crdVersions := []v1beta1.CustomResourceDefinitionVersion{}
	for _, v := range config.Versions {
		crdVersion := v1beta1.CustomResourceDefinitionVersion{
			Name:    v.Name,
			Served:  v.Enabled,
			Storage: v.Storage,
		}

		if v.Schema != nil {
			crdVersion.Schema = &v1beta1.CustomResourceValidation{
				OpenAPIV3Schema: v.Schema,
			}
		}

		if v.ScaleSubresources != nil {
			crdVersion.Subresources = createSubresources(v.ScaleSubresources)
		}

		crdVersion.AdditionalPrinterColumns = createColumns(v.ExtraColumns)
		crdVersions = append(crdVersions, crdVersion)
	}

	crd := v1beta1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CustomResourceDefinition",
			APIVersion: version,
		},
		ObjectMeta: generateObjectMeta(config.Name, config.Namespace, config.ClusterName),
		Spec: v1beta1.CustomResourceDefinitionSpec{
			Group:   config.Group,
			Version: config.CRDVersion,
			Names: v1beta1.CustomResourceDefinitionNames{
				Plural:     config.Plural,
				Singular:   config.Singular,
				ShortNames: config.ShortNames,
				Kind:       config.Kind,
				ListKind:   config.ListKind,
				Categories: config.Categories,
			},
			Subresources:             createSubresources(config.ScaleSubresources),
			Versions:                 crdVersions,
			AdditionalPrinterColumns: createColumns(config.ExtraColumns),
			Conversion:               createConversionStrategy(config),
		},
	}

	if config.Validation != nil {
		crd.Spec.Validation = &v1beta1.CustomResourceValidation{
			OpenAPIV3Schema: config.Validation,
		}
	}

	switch config.Scope {
	case api.CRDClusterScoped:
		crd.Spec.Scope = v1beta1.ClusterScoped
	case api.CRDNamespaceScoped:
		crd.Spec.Scope = v1beta1.NamespaceScoped
	}

	return &CustomResourceDefinition{&crd, MetadataFuncs{&crd}}
}

func createSubresources(config *api.CRDScaleSubresources) *v1beta1.CustomResourceSubresources {
	if config == nil {
		return nil
	}

	return &v1beta1.CustomResourceSubresources{
		Scale: &v1beta1.CustomResourceSubresourceScale{
			SpecReplicasPath:   config.SpecPath,
			StatusReplicasPath: config.StatusPath,
			LabelSelectorPath:  config.SelectorPath,
		},
	}
}

func createColumns(config []api.CRDColumn) []v1beta1.CustomResourceColumnDefinition {
	columns := []v1beta1.CustomResourceColumnDefinition{}

	for _, c := range config {
		column := v1beta1.CustomResourceColumnDefinition{
			Name:        c.Name,
			Type:        c.Type,
			Format:      c.Format,
			Description: c.Description,
			Priority:    c.Priority,
			JSONPath:    c.Path,
		}
		columns = append(columns, column)
	}

	return columns
}

func createConversionStrategy(config api.CRDConfig) *v1beta1.CustomResourceConversion {
	strategy := &v1beta1.CustomResourceConversion{
		ConversionReviewVersions: config.ConversionReviewVersions,
	}

	if len(config.ConversionWebhookServiceNamespace) > 0 || len(config.ConversionWebhookServiceName) > 0 ||
		len(config.ConversionWebhookCABundle) > 0 || config.ConversionWebhookServicePath != nil ||
		config.ConversionWebhookURL != nil {
		strategy.WebhookClientConfig = &v1beta1.WebhookClientConfig{
			URL: config.ConversionWebhookURL,
			Service: &v1beta1.ServiceReference{
				Namespace: config.ConversionWebhookServiceNamespace,
				Name:      config.ConversionWebhookServiceName,
				Path:      config.ConversionWebhookServicePath,
			},
			CABundle: config.ConversionWebhookCABundle,
		}
	}

	switch config.ConversionStrategy {
	case api.CRDConversionStraegyTypeNone:
		strategy.Strategy = v1beta1.NoneConverter
	case api.CRDConversionStraegyTypeWebhook:
		strategy.Strategy = v1beta1.WebhookConverter
	}

	if reflect.DeepEqual(strategy, &v1beta1.CustomResourceConversion{}) {
		return nil
	}

	return strategy
}

// Deploy will deploy the custom resource definition to the cluster
func (crd *CustomResourceDefinition) Deploy(res api.DeployerResources) error {
	_, err := res.KubeExtensionsClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd.CustomResourceDefinition)
	return err
}

// Undeploy will remove the custom resource definition from the cluster
func (crd *CustomResourceDefinition) Undeploy(res api.DeployerResources) error {
	return res.KubeExtensionsClient.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(crd.Name, &metav1.DeleteOptions{})
}
