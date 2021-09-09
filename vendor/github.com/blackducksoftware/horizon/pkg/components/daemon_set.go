/*
Copyright (C) 2019 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreementd. See the NOTICE file
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

	"k8s.io/api/apps/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DaemonSet defines the daemon set component
type DaemonSet struct {
	*v1.DaemonSet
	MetadataFuncs
	LabelSelectorFuncs
	PodFuncs
}

// NewDaemonSet creates a DaemonSet object
func NewDaemonSet(config api.DaemonSetConfig) *DaemonSet {
	version := "apps/v1"
	if len(config.APIVersion) > 0 {
		version = config.APIVersion
	}

	d := v1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: version,
		},
		ObjectMeta: generateObjectMeta(config.Name, config.Namespace, config.ClusterName),
		Spec: v1.DaemonSetSpec{
			MinReadySeconds:      config.MinReadySeconds,
			RevisionHistoryLimit: config.RevisionHistoryLimit,
		},
	}

	switch config.Strategy {
	case api.DaemonSetUpdateStrategyRollingUpdate:
		d.Spec.UpdateStrategy = v1.DaemonSetUpdateStrategy{
			Type: v1.RollingUpdateDaemonSetStrategyType,
			RollingUpdate: &v1.RollingUpdateDaemonSet{
				MaxUnavailable: createIntOrStr(config.MaxUnavailable),
			},
		}

	case api.DaemonSetUpdateStrategyOnDelete:
		d.Spec.UpdateStrategy = v1.DaemonSetUpdateStrategy{
			Type: v1.OnDeleteDaemonSetStrategyType,
		}
	}

	return &DaemonSet{&d, MetadataFuncs{&d}, LabelSelectorFuncs{&d}, PodFuncs{&d}}
}

// Deploy will deploy the daemon set to the cluster
func (ds *DaemonSet) Deploy(res api.DeployerResources) error {
	_, err := res.KubeClient.AppsV1().DaemonSets(ds.Namespace).Create(ds.DaemonSet)
	return err
}

// Undeploy will remove the daemon set from the cluster
func (ds *DaemonSet) Undeploy(res api.DeployerResources) error {
	return res.KubeClient.AppsV1().DaemonSets(ds.Namespace).Delete(ds.Name, &metav1.DeleteOptions{})
}
