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

package crdupdater

import (
	"reflect"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type label struct {
	operator string
	value    []string
}

// getLabelsMap convert the label selector string to kbernetes label format
func getLabelsMap(labelSelectors string) map[string]label {
	labelSelectorArr := strings.SplitN(labelSelectors, ",", 3)
	expectedLabels := make(map[string]label, len(labelSelectorArr))
	for _, labelSelector := range labelSelectorArr {
		if strings.Contains(labelSelector, "!=") {
			labels := strings.Split(labelSelector, "!=")
			if len(labels) == 2 {
				expectedLabels[labels[0]] = label{operator: "!=", value: []string{labels[1]}}
			}
		} else if strings.Contains(labelSelector, "=") {
			labels := strings.Split(labelSelector, "=")
			if len(labels) == 2 {
				expectedLabels[labels[0]] = label{operator: "=", value: []string{labels[1]}}
			}
		} else if strings.Contains(labelSelector, " in ") {
			labels := strings.Split(labelSelector, " in (")
			if len(labels) == 2 {
				values := []string{}
				valueArr := strings.Split(labels[1][:len(labels[1])-1], ",")
				for _, value := range valueArr {
					values = append(values, value)
				}
				expectedLabels[labels[0]] = label{operator: "in", value: values}
			}
		} else if strings.Contains(labelSelector, " notin ") {
			labels := strings.Split(labelSelector, " notin (")
			if len(labels) == 2 {
				values := []string{}
				valueArr := strings.Split(labels[1][:len(labels[1])-1], ",")
				for _, value := range valueArr {
					values = append(values, value)
				}
				expectedLabels[labels[0]] = label{operator: "notin", value: values}
			}
		}

	}
	// fmt.Printf("expected Labels: %+v\n", expectedLabels)
	return expectedLabels
}

// isLabelsExist checks for the expected labels to match with the actual labels
func isLabelsExist(expectedLabels map[string]label, actualLabels map[string]string) bool {
	for key, expectedValue := range expectedLabels {
		actualValue, ok := actualLabels[key]
		if !ok {
			return false
		}
		switch expectedValue.operator {
		case "=":
			if !strings.EqualFold(expectedValue.value[0], actualValue) {
				return false
			}
		case "!=":
			if strings.EqualFold(expectedValue.value[0], actualValue) {
				return false
			}
		case "in":
			isExist := false
			for _, value := range expectedValue.value {
				if strings.EqualFold(value, actualValue) {
					isExist = true
					break
				}
			}
			if !isExist {
				return false
			}
		case "notin":
			isExist := false
			for _, value := range expectedValue.value {
				if strings.EqualFold(value, actualValue) {
					isExist = true
					break
				}
			}
			if isExist {
				return false
			}
		}
	}
	return true
}

type servicePort struct {
	Name       string             `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	Protocol   corev1.Protocol    `json:"protocol,omitempty" protobuf:"bytes,2,opt,name=protocol,casttype=Protocol"`
	Port       int32              `json:"port" protobuf:"varint,3,opt,name=port"`
	TargetPort intstr.IntOrString `json:"targetPort,omitempty" protobuf:"bytes,4,opt,name=targetPort"`
}

func sortPorts(ports []corev1.ServicePort) []servicePort {
	sort.Slice(ports, func(i, j int) bool { return ports[i].Name < ports[j].Name })
	servicePorts := []servicePort{}
	for _, port := range ports {
		servicePorts = append(servicePorts, servicePort{Name: port.Name, Protocol: port.Protocol, Port: port.Port, TargetPort: port.TargetPort})
	}
	return servicePorts
}

type policyRule struct {
	Verbs     []string `json:"verbs" protobuf:"bytes,1,rep,name=verbs"`
	APIGroups []string `json:"apiGroups,omitempty" protobuf:"bytes,2,rep,name=apiGroups"`
	Resources []string `json:"resources,omitempty" protobuf:"bytes,3,rep,name=resources"`
}

func sortPolicyRule(rules []rbacv1.PolicyRule) []policyRule {
	policyRules := []policyRule{}
	for _, rule := range rules {
		sort.Strings(rule.APIGroups)
		sort.Strings(rule.Resources)
		sort.Strings(rule.Verbs)
		policyRules = append(policyRules, policyRule{Verbs: rule.Verbs, APIGroups: rule.APIGroups, Resources: rule.APIGroups})
	}
	return policyRules
}

func sortEnvs(envs []corev1.EnvVar) []corev1.EnvVar {
	sort.Slice(envs, func(i, j int) bool { return envs[i].Name < envs[j].Name })
	return envs
}

func sortVolumeMounts(volumeMounts []corev1.VolumeMount) []corev1.VolumeMount {
	sort.Slice(volumeMounts, func(i, j int) bool { return volumeMounts[i].Name < volumeMounts[j].Name })
	return volumeMounts
}

func compareVolumes(oldVolume []corev1.Volume, newVolume []corev1.Volume) bool {
	for i, volume := range oldVolume {
		if volume.Secret != nil && !reflect.DeepEqual(volume.Secret.SecretName, newVolume[i].Secret.SecretName) && !reflect.DeepEqual(volume.Secret.Items, newVolume[i].Secret.Items) {
			return false
		} else if volume.ConfigMap != nil && !reflect.DeepEqual(volume.ConfigMap.Name, newVolume[i].ConfigMap.Name) && !reflect.DeepEqual(volume.ConfigMap.Items, newVolume[i].ConfigMap.Items) {
			return false
		} else if volume.Secret == nil && volume.ConfigMap == nil && !reflect.DeepEqual(volume, newVolume[i]) {
			return false
		}
	}
	return true
}

func compareProbes(oldProbe *corev1.Probe, newProbe *corev1.Probe) bool {
	if oldProbe == nil && newProbe == nil {
		return true
	} else if (oldProbe != nil && newProbe == nil) || (oldProbe == nil && newProbe != nil) {
		return false
	} else if !reflect.DeepEqual(oldProbe.Handler, newProbe.Handler) &&
		!reflect.DeepEqual(oldProbe.InitialDelaySeconds, newProbe.InitialDelaySeconds) &&
		!reflect.DeepEqual(oldProbe.TimeoutSeconds, newProbe.TimeoutSeconds) &&
		!reflect.DeepEqual(oldProbe.PeriodSeconds, newProbe.PeriodSeconds) {
		return false
	}
	return true
}

func sortVolumes(volumes []corev1.Volume) []corev1.Volume {
	for _, volume := range volumes {
		if volume.Secret != nil {
			sort.Slice(volume.Secret.Items, func(i, j int) bool { return volume.Secret.Items[i].Key < volume.Secret.Items[j].Key })
		}
		if volume.ConfigMap != nil {
			sort.Slice(volume.ConfigMap.Items, func(i, j int) bool { return volume.ConfigMap.Items[i].Key < volume.ConfigMap.Items[j].Key })
		}
	}
	sort.Slice(volumes, func(i, j int) bool { return volumes[i].Name < volumes[j].Name })
	return volumes
}

func compareAffinities(oldAffinity *corev1.Affinity, newAffinity *corev1.Affinity) bool {
	if oldAffinity == nil && newAffinity == nil {
		return true
	} else if (oldAffinity == nil && newAffinity != nil) || (oldAffinity != nil && newAffinity == nil) {
		return false
	} else {
		return compareNodeAffinities(oldAffinity.NodeAffinity, newAffinity.NodeAffinity)
	}
}

func compareNodeAffinities(oldNodeAffinity *corev1.NodeAffinity, newNodeAffinity *corev1.NodeAffinity) bool {
	log.Debugf("oldNodeAffinity: %v", oldNodeAffinity)
	log.Debugf("newNodeAffinity: %v", newNodeAffinity)

	if oldNodeAffinity == nil && newNodeAffinity == nil {
		return true
	} else if (oldNodeAffinity == nil && newNodeAffinity != nil) || (oldNodeAffinity != nil && newNodeAffinity == nil) {
		return false
	} else {
		return compareNodeSelector(oldNodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution, newNodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution) && comparePreferredSchedulingTerms(oldNodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution, newNodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution)
	}
}

func comparePreferredSchedulingTerms(oldTerms []corev1.PreferredSchedulingTerm, newTerms []corev1.PreferredSchedulingTerm) bool {
	if len(oldTerms) != len(newTerms) {
		return false
	}

	for idx, oldTerm := range oldTerms {
		newTerm := newTerms[idx]
		if !reflect.DeepEqual(oldTerm.Weight, newTerm.Weight) && !compareNodeSelectorTerm(&oldTerm.Preference, &newTerm.Preference) {
			return false
		}
	}
	return true
}

func compareNodeSelector(oldSelector *corev1.NodeSelector, newSelector *corev1.NodeSelector) bool {
	if oldSelector == nil && newSelector == nil {
		return true
	} else if (oldSelector == nil && newSelector != nil) || (oldSelector != nil && newSelector == nil) {
		return false
	} else {
		return compareNodeSelectorTerms(oldSelector.NodeSelectorTerms, newSelector.NodeSelectorTerms)
	}
}

func compareNodeSelectorTerms(oldTerms []corev1.NodeSelectorTerm, newTerms []corev1.NodeSelectorTerm) bool {
	log.Debugf("oldTerms: %v", oldTerms)
	log.Debugf("newTerms: %v", newTerms)

	if len(oldTerms) != len(newTerms) {
		return false
	}
	for idx, oldTerm := range oldTerms {
		if !compareNodeSelectorTerm(&oldTerm, &newTerms[idx]) {
			return false
		}
	}
	return true
}

func compareNodeSelectorTerm(oldTerm *corev1.NodeSelectorTerm, newTerm *corev1.NodeSelectorTerm) bool {
	log.Debugf("oldTerm: %v", oldTerm)
	log.Debugf("newTerm: %v", newTerm)

	if oldTerm == nil && newTerm == nil {
		return true
	} else if (oldTerm == nil && newTerm != nil) || (oldTerm != nil && newTerm == nil) {
		return false
	} else {
		return compareNodeSelectoreRequirements(oldTerm.MatchExpressions, newTerm.MatchExpressions) && compareNodeSelectoreRequirements(oldTerm.MatchFields, newTerm.MatchFields)
	}
}

func compareNodeSelectoreRequirements(oldReqmnts []corev1.NodeSelectorRequirement, newReqmnts []corev1.NodeSelectorRequirement) bool {
	if len(oldReqmnts) != len(newReqmnts) {
		return false
	}
	for idx, req := range oldReqmnts {
		oldReq := &req
		newReq := &newReqmnts[idx]

		oldKey := oldReq.Key
		newKey := newReq.Key

		log.Debugf("oldKey: %v", oldKey)
		log.Debugf("newKey: %v", newKey)

		oldOperator := oldReq.Operator
		newOperator := newReq.Operator

		log.Debugf("oldOperator: %v", oldOperator)
		log.Debugf("newOperator: %v", newOperator)

		oldValues := oldReq.Values
		newValues := newReq.Values

		log.Debugf("oldValues: %v", oldValues)
		log.Debugf("newValues: %v", newValues)

		if !reflect.DeepEqual(oldKey, newKey) && !reflect.DeepEqual(oldOperator, newOperator) && !reflect.DeepEqual(oldValues, newValues) {
			return false
		}
	}
	return true
}
