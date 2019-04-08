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
	"fmt"
	"strings"

	"github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/util"

	"github.com/koki/short/converter/converters"
	"github.com/koki/short/types"

	"k8s.io/apimachinery/pkg/runtime"
)

// ReplicationController defines the replication controller component
type ReplicationController struct {
	obj *types.ReplicationController
}

// NewReplicationController creates a ReplicationController object
func NewReplicationController(config api.ReplicationControllerConfig) *ReplicationController {
	rc := &types.ReplicationController{
		Version:         config.APIVersion,
		Name:            config.Name,
		Cluster:         config.ClusterName,
		Namespace:       config.Namespace,
		Replicas:        config.Replicas,
		MinReadySeconds: config.ReadySeconds,
	}

	return &ReplicationController{obj: rc}
}

// GetObj returns the replication controller object in a format the deployer can use
func (rc *ReplicationController) GetObj() *types.ReplicationController {
	return rc.obj
}

// GetName returns the name of the replication controller
func (rc *ReplicationController) GetName() string {
	return rc.obj.Name
}

// AddAnnotations adds annotations to the replication controller
func (rc *ReplicationController) AddAnnotations(new map[string]string) {
	rc.obj.Annotations = util.MapMerge(rc.obj.Annotations, new)
}

// RemoveAnnotations removes annotations from the replication controller
func (rc *ReplicationController) RemoveAnnotations(remove []string) {
	for _, k := range remove {
		rc.obj.Annotations = util.RemoveElement(rc.obj.Annotations, k)
	}
}

// AddLabels adds labels to the replication controller
func (rc *ReplicationController) AddLabels(new map[string]string) {
	rc.obj.Labels = util.MapMerge(rc.obj.Labels, new)
}

// RemoveLabels removes labels from the replication controller
func (rc *ReplicationController) RemoveLabels(remove []string) {
	for _, k := range remove {
		rc.obj.Labels = util.RemoveElement(rc.obj.Labels, k)
	}
}

// AddPod adds a pod to the replication controller
func (rc *ReplicationController) AddPod(obj *Pod) error {
	o := obj.GetObj()
	rc.obj.TemplateMetadata = &o.PodTemplateMeta
	rc.obj.PodTemplate = o.PodTemplate

	return nil
}

// RemovePod removes a pod from the replication controller
func (rc *ReplicationController) RemovePod(name string) error {
	if strings.Compare(rc.obj.TemplateMetadata.Name, name) != 0 {
		return fmt.Errorf("pod with name %s doesn't exist on replication controller", name)
	}
	rc.obj.TemplateMetadata = nil
	rc.obj.PodTemplate = types.PodTemplate{}
	return nil
}

// AddLabelSelectors adds the given label selectors to the replication controller
func (rc *ReplicationController) AddLabelSelectors(new map[string]string) {
	rc.obj.Selector = util.MapMerge(rc.obj.Selector, new)
}

// RemoveLabelSelectors removes the given label selectors from the replication controller
func (rc *ReplicationController) RemoveLabelSelectors(remove []string) {
	for _, k := range remove {
		rc.obj.Selector = util.RemoveElement(rc.obj.Selector, k)
	}
}

// ToKube returns the kubernetes version of the replication controller
func (rc *ReplicationController) ToKube() (runtime.Object, error) {
	wrapper := &types.ReplicationControllerWrapper{ReplicationController: *rc.obj}
	return converters.Convert_Koki_ReplicationController_to_Kube_v1_ReplicationController(wrapper)
}
