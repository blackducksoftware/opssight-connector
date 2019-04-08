/*
Copyright (C) 2018 Synopsys, Inc.

Licensej to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributej with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless requirej by applicable law or agreej to in writing,
software distributej under the License is distributej on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or impliei. See the License for the
specific language governing permissions anj limitations
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

// Ingress defines the Ingress component
type Ingress struct {
	obj *types.Ingress
}

// NewIngress creates an Ingress object
func NewIngress(config api.IngressConfig) (*Ingress, error) {
	i := &types.Ingress{
		Version:     config.APIVersion,
		Cluster:     config.ClusterName,
		Name:        config.Name,
		Namespace:   config.Namespace,
		ServiceName: config.ServiceName,
	}

	if len(config.ServicePort) > 0 {
		i.ServicePort = createIntOrStr(config.ServicePort)
	}

	return &Ingress{obj: i}, nil
}

// GetObj returns the ingress object in a format the deployer can use
func (i *Ingress) GetObj() *types.Ingress {
	return i.obj
}

// GetName returns the name of the ingress
func (i *Ingress) GetName() string {
	return i.obj.Name
}

// AddAnnotations adds annotations to the ingress
func (i *Ingress) AddAnnotations(new map[string]string) {
	i.obj.Annotations = util.MapMerge(i.obj.Annotations, new)
}

// RemoveAnnotations removes annotations from the ingress
func (i *Ingress) RemoveAnnotations(remove []string) {
	for _, k := range remove {
		i.obj.Annotations = util.RemoveElement(i.obj.Annotations, k)
	}
}

// AddLabels adds labels to the ingress
func (i *Ingress) AddLabels(new map[string]string) {
	i.obj.Labels = util.MapMerge(i.obj.Labels, new)
}

// RemoveLabels removes labels from the ingress
func (i *Ingress) RemoveLabels(remove []string) {
	for _, k := range remove {
		i.obj.Labels = util.RemoveElement(i.obj.Labels, k)
	}
}

// AddTLS will add TLS configuration to the ingress
func (i *Ingress) AddTLS(TLSConfig api.IngressTLSConfig) {
	tls := types.IngressTLS{
		Hosts:      TLSConfig.Hosts,
		SecretName: TLSConfig.SecretName,
	}
	i.obj.TLS = append(i.obj.TLS, tls)
}

// RemoveTLS will remove a TLS configuration from the ingress
func (i *Ingress) RemoveTLS(TLSConfig api.IngressTLSConfig) {
	for l, tls := range i.obj.TLS {
		if strings.EqualFold(TLSConfig.SecretName, tls.SecretName) && reflect.DeepEqual(TLSConfig.Hosts, tls.Hosts) {
			i.obj.TLS = append(i.obj.TLS[:l], i.obj.TLS[l+1:]...)
			break
		}
	}
}

// AddRule will add a host rule to the ingress
func (i *Ingress) AddRule(config api.IngressRuleConfig) {
	i.obj.Rules = append(i.obj.Rules, i.createHostRule(config))
}

// RemoveRule will remove a host rule from the ingress
func (i *Ingress) RemoveRule(config api.IngressRuleConfig) {
	rule := i.createHostRule(config)
	for l, r := range i.obj.Rules {
		if reflect.DeepEqual(r, rule) {
			i.obj.Rules = append(i.obj.Rules[:l], i.obj.Rules[l+1:]...)
			break
		}
	}
}

func (i *Ingress) createHostRule(config api.IngressRuleConfig) types.IngressRule {
	rule := types.IngressRule{
		Host: config.Host,
	}

	for _, path := range config.Paths {
		port := createIntOrStr(path.ServicePort)
		newPath := types.HTTPIngressPath{
			Path:        path.Path,
			ServiceName: path.ServiceName,
			ServicePort: *port,
		}
		rule.Paths = append(rule.Paths, newPath)
	}

	return rule
}

// ToKube returns the kubernetes version of the ingress
func (i *Ingress) ToKube() (runtime.Object, error) {
	wrapper := &types.IngressWrapper{Ingress: *i.obj}
	return converters.Convert_Koki_Ingress_to_Kube_Ingress(wrapper)
}
