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

package util

// ComponentType defines the component type
type ComponentType string

const (
	DeploymentComponent              ComponentType = "Deployment"
	PodComponent                     ComponentType = "Pod"
	ConfigMapComponent               ComponentType = "ConfigMap"
	SecretComponent                  ComponentType = "Secret"
	ServiceComponent                 ComponentType = "Service"
	ServiceAccountComponent          ComponentType = "ServiceAccount"
	ClusterRoleComponent             ComponentType = "ClusterRole"
	ClusterRoleBindingComponent      ComponentType = "ClusterRoleBinding"
	ReplicationControllerComponent   ComponentType = "ReplicationController"
	CRDComponent                     ComponentType = "CustomResourceDefinition"
	Controller                       ComponentType = "Controller"
	NamespaceComponent               ComponentType = "Namespace"
	PersistentVolumeClaimComponent   ComponentType = "PersistentVolumeClaim"
	JobComponent                     ComponentType = "Job"
	HorizontalPodAutoscalerComponent ComponentType = "HorizontalPodAutoscaler"
	IngressComponent                 ComponentType = "Ingress"
	StatefulSetComponent             ComponentType = "StatefulSetIngress"
	DaemonSetComponent               ComponentType = "DaemonSetIngress"
)
