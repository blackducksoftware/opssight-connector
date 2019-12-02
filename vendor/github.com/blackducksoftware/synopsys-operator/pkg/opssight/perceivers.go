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

package opssight

import (
	"fmt"
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	routev1 "github.com/openshift/api/route/v1"
)

// PodPerceiverReplicationController creates a replication controller for the pod perceiver
func (p *SpecConfig) PodPerceiverReplicationController() (*components.ReplicationController, error) {
	name := p.names["pod-perceiver"]
	image := p.images["pod-perceiver"]

	rc := p.perceiverReplicationController(name, 1)

	pod, err := p.perceiverPod(name, image, util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["perceiver-service-account"]))
	if err != nil {
		return nil, errors.Annotate(err, "failed to create pod perceiver pod")
	}
	rc.AddPod(pod)

	return rc, nil
}

// ImagePerceiverReplicationController creates a replication controller for the image perceiver
func (p *SpecConfig) ImagePerceiverReplicationController() (*components.ReplicationController, error) {
	name := p.names["image-perceiver"]
	image := p.images["image-perceiver"]

	rc := p.perceiverReplicationController(name, 1)

	pod, err := p.perceiverPod(name, image, util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["perceiver-service-account"]))
	if err != nil {
		return nil, errors.Annotate(err, "failed to create image perceiver pod")
	}
	rc.AddPod(pod)

	return rc, nil
}

// ArtifactoryPerceiverReplicationController creates a replication controller for the artifactory perceiver
func (p *SpecConfig) ArtifactoryPerceiverReplicationController() (*components.ReplicationController, error) {
	name := p.names["artifactory-perceiver"]
	image := p.images["artifactory-perceiver"]

	rc := p.perceiverReplicationController(name, 1)

	pod, err := p.perceiverPod(name, image, "")
	if err != nil {
		return nil, errors.Annotate(err, "failed to create artifactory perceiver pod")
	}
	rc.AddPod(pod)
	return rc, nil
}

// QuayPerceiverReplicationController creates a replication controller for the quay perceiver
func (p *SpecConfig) QuayPerceiverReplicationController() (*components.ReplicationController, error) {
	name := p.names["quay-perceiver"]
	image := p.images["quay-perceiver"]

	rc := p.perceiverReplicationController(name, 1)

	pod, err := p.perceiverPod(name, image, "")
	if err != nil {
		return nil, errors.Annotate(err, "failed to create quay perceiver pod")
	}
	rc.AddPod(pod)
	return rc, nil
}

func (p *SpecConfig) perceiverReplicationController(name string, replicas int32) *components.ReplicationController {
	rc := components.NewReplicationController(horizonapi.ReplicationControllerConfig{
		Replicas:  &replicas,
		Name:      util.GetResourceName(p.opssight.Name, util.OpsSightName, name),
		Namespace: p.opssight.Spec.Namespace,
	})
	rc.AddSelectors(map[string]string{"component": name, "app": "opssight", "name": p.opssight.Name})
	rc.AddLabels(map[string]string{"component": name, "app": "opssight", "name": p.opssight.Name})
	return rc
}

func (p *SpecConfig) perceiverPod(name string, image string, account string) (*components.Pod, error) {
	pod := components.NewPod(horizonapi.PodConfig{
		Name:           util.GetResourceName(p.opssight.Name, util.OpsSightName, name),
		ServiceAccount: account,
	})

	pod.AddLabels(map[string]string{"component": name, "app": "opssight", "name": p.opssight.Name})
	container, err := p.perceiverContainer(name, image)
	if err != nil {
		return nil, errors.Trace(err)
	}
	pod.AddContainer(container)

	vols, err := p.perceiverVolumes(name)

	if err != nil {
		return nil, errors.Annotate(err, "unable to create volumes")
	}

	for _, v := range vols {
		err = pod.AddVolume(v)
		if err != nil {
			return nil, errors.Annotate(err, "unable to add volume to pod")
		}
	}

	if p.opssight.Spec.RegistryConfiguration != nil && len(p.opssight.Spec.RegistryConfiguration.PullSecrets) > 0 {
		pod.AddImagePullSecrets(p.opssight.Spec.RegistryConfiguration.PullSecrets)
	}

	return pod, nil
}

func (p *SpecConfig) perceiverContainer(name string, image string) (*components.Container, error) {
	cmd := fmt.Sprintf("./%s", name)
	if strings.Contains(name, "processor") {
		cmd = fmt.Sprintf("./opssight-%s", name)
	}
	container, err := components.NewContainer(horizonapi.ContainerConfig{
		Name:    name,
		Image:   image,
		Command: []string{cmd},
		Args:    []string{fmt.Sprintf("/etc/%s/%s.json", name, p.names["configmap"])},
		MinCPU:  p.opssight.Spec.DefaultCPU,
		MinMem:  p.opssight.Spec.DefaultMem,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	container.AddPort(horizonapi.PortConfig{
		ContainerPort: int32(3002),
		Protocol:      horizonapi.ProtocolTCP,
	})

	err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      name,
		MountPath: fmt.Sprintf("/etc/%s", name),
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = container.AddVolumeMount(horizonapi.VolumeMountConfig{
		Name:      "logs",
		MountPath: "/tmp",
	})
	if err != nil {
		return nil, errors.Annotatef(err, "unable to add the volume mount to %s container", name)
	}

	if strings.Contains(name, "artifactory") || strings.Contains(name, "quay") {
		container.AddEnv(horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, FromName: util.GetResourceName(p.opssight.Name, util.OpsSightName, "blackduck")})
	}
	return container, nil
}

func (p *SpecConfig) perceiverVolumes(name string) ([]*components.Volume, error) {
	vols := []*components.Volume{p.configMapVolume(name)}

	vol, err := components.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "logs",
		Medium:     horizonapi.StorageMediumDefault,
	})

	if err != nil {
		return nil, errors.Annotate(err, "failed to create empty dir volume")
	}
	vols = append(vols, vol)

	return vols, nil
}

// PerceiverNodePortService creates a nodeport service for perceiver
func (p *SpecConfig) PerceiverNodePortService(perceiverName string) (*components.Service, error) {
	name := fmt.Sprintf("%s-exposed", p.names[fmt.Sprintf("%s-perceiver", perceiverName)])
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      util.GetResourceName(p.opssight.Name, util.OpsSightName, name),
		Namespace: p.opssight.Spec.Namespace,
		Type:      horizonapi.ServiceTypeNodePort,
	})

	err := service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(3002),
		TargetPort: fmt.Sprintf("%d", 3002),
		Protocol:   horizonapi.ProtocolTCP,
		Name:       fmt.Sprintf("port-%s", name),
	})
	if err != nil {
		return nil, errors.Trace(err)
	}

	service.AddLabels(map[string]string{"component": p.names[fmt.Sprintf("%s-perceiver", perceiverName)], "app": "opssight", "name": p.opssight.Name})
	service.AddSelectors(map[string]string{"component": p.names[fmt.Sprintf("%s-perceiver", perceiverName)], "app": "opssight", "name": p.opssight.Name})

	return service, nil
}

// PerceiverLoadBalancerService creates a loadbalancer service for perceiver
func (p *SpecConfig) PerceiverLoadBalancerService(perceiverName string) (*components.Service, error) {
	name := fmt.Sprintf("%s-exposed", p.names[fmt.Sprintf("%s-perceiver", perceiverName)])
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      util.GetResourceName(p.opssight.Name, util.OpsSightName, name),
		Namespace: p.opssight.Spec.Namespace,
		Type:      horizonapi.ServiceTypeLoadBalancer,
	})

	err := service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(3002),
		TargetPort: fmt.Sprintf("%d", 3002),
		Protocol:   horizonapi.ProtocolTCP,
		Name:       fmt.Sprintf("port-%s", name),
	})
	if err != nil {
		return nil, errors.Trace(err)

	}

	service.AddLabels(map[string]string{"component": p.names[fmt.Sprintf("%s-perceiver", perceiverName)], "app": "opssight", "name": p.opssight.Name})
	service.AddSelectors(map[string]string{"component": p.names[fmt.Sprintf("%s-perceiver", perceiverName)], "app": "opssight", "name": p.opssight.Name})

	return service, nil
}

// GetPerceiverOpenShiftRoute creates the OpenShift route component for the perceiver model
func (p *SpecConfig) GetPerceiverOpenShiftRoute(perceiverName string, secure bool) *api.Route {
	perceiverName = fmt.Sprintf("%s-perceiver", perceiverName)
	name := p.names[perceiverName]
	namespace := p.opssight.Spec.Namespace
	if strings.ToUpper(p.opssight.Spec.Perceiver.Expose) == util.OPENSHIFT {
		route := &api.Route{
			Name:        util.GetResourceName(p.opssight.Name, util.OpsSightName, name),
			Namespace:   namespace,
			Kind:        "Service",
			ServiceName: util.GetResourceName(p.opssight.Name, util.OpsSightName, name),
			PortName:    fmt.Sprintf("port-%s", name),
			Labels:      map[string]string{"app": "opssight", "name": p.opssight.Name, "component": fmt.Sprintf("%s-exposed", name)},
		}
		if secure {
			route.TLSTerminationType = routev1.TLSTerminationPassthrough
		}
		return route
	}
	return nil
}

func (p *SpecConfig) perceiverService(name string) *components.Service {
	service := components.NewService(horizonapi.ServiceConfig{
		Name:      util.GetResourceName(p.opssight.Name, util.OpsSightName, name),
		Namespace: p.opssight.Spec.Namespace,
		Type:      horizonapi.ServiceTypeServiceIP,
	})

	service.AddPort(horizonapi.ServicePortConfig{
		Port:       int32(3002),
		TargetPort: fmt.Sprintf("%d", 3002),
		Protocol:   horizonapi.ProtocolTCP,
		Name:       fmt.Sprintf("port-%s", name),
	})

	service.AddLabels(map[string]string{"component": name, "app": "opssight", "name": p.opssight.Name})
	service.AddSelectors(map[string]string{"component": name, "app": "opssight", "name": p.opssight.Name})

	return service
}

// PodPerceiverService creates a service for the pod perceiver
func (p *SpecConfig) PodPerceiverService() *components.Service {
	return p.perceiverService(p.names["pod-perceiver"])
}

// ImagePerceiverService creates a service for the image perceiver
func (p *SpecConfig) ImagePerceiverService() *components.Service {
	return p.perceiverService(p.names["image-perceiver"])
}

// ArtifactoryPerceiverService creates a service for the Artifactory perceiver
func (p *SpecConfig) ArtifactoryPerceiverService() *components.Service {
	return p.perceiverService(p.names["artifactory-perceiver"])
}

// QuayPerceiverService creates a service for the Quay perceiver
func (p *SpecConfig) QuayPerceiverService() *components.Service {
	return p.perceiverService(p.names["quay-perceiver"])
}

func (p *SpecConfig) perceiverServiceAccount(name string) *components.ServiceAccount {
	serviceAccount := components.NewServiceAccount(horizonapi.ServiceAccountConfig{
		Name:      util.GetResourceName(p.opssight.Name, util.OpsSightName, name),
		Namespace: p.opssight.Spec.Namespace,
	})
	serviceAccount.AddLabels(map[string]string{"component": name, "app": "opssight", "name": p.opssight.Name})
	return serviceAccount
}

// PodPerceiverServiceAccount creates a service account for the pod perceiver
func (p *SpecConfig) PodPerceiverServiceAccount() *components.ServiceAccount {
	return p.perceiverServiceAccount(p.names["perceiver-service-account"])
}

// ImagePerceiverServiceAccount creates a service account for the image perceiver
func (p *SpecConfig) ImagePerceiverServiceAccount() *components.ServiceAccount {
	return p.perceiverServiceAccount(p.names["perceiver-service-account"])
}

// PodPerceiverClusterRole creates a cluster role for the pod perceiver
func (p *SpecConfig) PodPerceiverClusterRole() *components.ClusterRole {
	clusterRole := components.NewClusterRole(horizonapi.ClusterRoleConfig{
		Name:       util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["pod-perceiver"]),
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	clusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		APIGroups: []string{""},
		Resources: []string{"pods"},
		Verbs:     []string{"get", "watch", "list", "update"},
	})
	clusterRole.AddLabels(map[string]string{"component": p.names["pod-perceiver"], "app": "opssight", "name": p.opssight.Name})

	return clusterRole
}

// ImagePerceiverClusterRole creates a cluster role for the image perceiver
func (p *SpecConfig) ImagePerceiverClusterRole() *components.ClusterRole {
	clusterRole := components.NewClusterRole(horizonapi.ClusterRoleConfig{
		Name:       util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["image-perceiver"]),
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	clusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		APIGroups: []string{"image.openshift.io"},
		Resources: []string{"images"},
		Verbs:     []string{"get", "watch", "list", "update"},
	})
	clusterRole.AddLabels(map[string]string{"component": p.names["image-perceiver"], "app": "opssight", "name": p.opssight.Name})

	return clusterRole
}

// PodPerceiverClusterRoleBinding creates a cluster role binding for the pod perceiver
func (p *SpecConfig) PodPerceiverClusterRoleBinding(clusterRole *components.ClusterRole) *components.ClusterRoleBinding {
	clusterRoleBinding := components.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		Name:       util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["pod-perceiver"]),
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	clusterRoleBinding.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["perceiver-service-account"]),
		Namespace: p.opssight.Spec.Namespace,
	})
	clusterRoleBinding.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     clusterRole.GetName(),
	})
	clusterRoleBinding.AddLabels(map[string]string{"component": p.names["pod-perceiver"], "app": "opssight", "name": p.opssight.Name})

	return clusterRoleBinding
}

// ImagePerceiverClusterRoleBinding creates a cluster role binding for the image perceiver
func (p *SpecConfig) ImagePerceiverClusterRoleBinding(clusterRole *components.ClusterRole) *components.ClusterRoleBinding {
	clusterRoleBinding := components.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		Name:       util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["image-perceiver"]),
		APIVersion: "rbac.authorization.k8s.io/v1",
	})
	clusterRoleBinding.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["perceiver-service-account"]),
		Namespace: p.opssight.Spec.Namespace,
	})
	clusterRoleBinding.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     clusterRole.GetName(),
	})
	clusterRoleBinding.AddLabels(map[string]string{"component": p.names["image-perceiver"], "app": "opssight", "name": p.opssight.Name})

	return clusterRoleBinding
}
