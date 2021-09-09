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

package soperator

import (
	"encoding/json"
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	horizoncomponents "github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	routev1 "github.com/openshift/api/route/v1"
	log "github.com/sirupsen/logrus"
)

// getOperatorDeployment creates a deployment for Synopsys Operaotor
func (specConfig *SpecConfig) getOperatorDeployment() (*horizoncomponents.Deployment, error) {
	// Add the Replication Controller to the Deployer
	var synopsysOperatorReplicas int32 = 1
	synopsysOperator := horizoncomponents.NewDeployment(horizonapi.DeploymentConfig{
		Name:      "synopsys-operator",
		Namespace: specConfig.Namespace,
		Replicas:  &synopsysOperatorReplicas,
	})

	synopsysOperator.AddMatchLabelsSelectors(map[string]string{"app": "synopsys-operator", "component": "operator"})

	synopsysOperatorPod := horizoncomponents.NewPod(horizonapi.PodConfig{
		APIVersion:     "v1",
		Name:           "synopsys-operator",
		Namespace:      specConfig.Namespace,
		ServiceAccount: "synopsys-operator",
	})

	synopsysOperatorContainer, err := horizoncomponents.NewContainer(horizonapi.ContainerConfig{
		Name:       "synopsys-operator",
		Args:       []string{"/etc/synopsys-operator/config.json"},
		Command:    []string{"./operator"},
		Image:      specConfig.Image,
		PullPolicy: horizonapi.PullAlways,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}
	synopsysOperatorContainer.AddPort(horizonapi.PortConfig{
		ContainerPort: int32(8080),
	})
	synopsysOperatorContainer.AddVolumeMount(horizonapi.VolumeMountConfig{
		MountPath: "/etc/synopsys-operator",
		Name:      "synopsys-operator",
	})
	synopsysOperatorContainer.AddVolumeMount(horizonapi.VolumeMountConfig{
		MountPath: "/opt/synopsys-operator/tls",
		Name:      "synopsys-operator-tls",
	})
	synopsysOperatorContainer.AddVolumeMount(horizonapi.VolumeMountConfig{
		MountPath: "/tmp",
		Name:      "tmp-logs",
	})
	synopsysOperatorContainer.AddEnv(horizonapi.EnvConfig{
		NameOrPrefix: "SEAL_KEY",
		Type:         horizonapi.EnvFromSecret,
		KeyOrVal:     "SEAL_KEY",
		FromName:     "blackduck-secret",
	})

	synopsysOperatorUIContainer, err := horizoncomponents.NewContainer(horizonapi.ContainerConfig{
		Name:       "synopsys-operator-ui",
		Command:    []string{"./app"},
		Image:      specConfig.Image,
		PullPolicy: horizonapi.PullAlways,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}
	synopsysOperatorUIContainer.AddPort(horizonapi.PortConfig{
		ContainerPort: int32(3000),
	})
	synopsysOperatorUIContainer.AddVolumeMount(horizonapi.VolumeMountConfig{
		MountPath: "/etc/synopsys-operator",
		Name:      "synopsys-operator",
	})
	synopsysOperatorUIContainer.AddEnv(horizonapi.EnvConfig{
		NameOrPrefix: "CONFIG_FILE_PATH",
		Type:         horizonapi.EnvVal,
		KeyOrVal:     "/etc/synopsys-operator/config.json",
	})
	synopsysOperatorUIContainer.AddEnv(horizonapi.EnvConfig{
		NameOrPrefix: "ADDR",
		Type:         horizonapi.EnvVal,
		KeyOrVal:     "0.0.0.0",
	})
	synopsysOperatorUIContainer.AddEnv(horizonapi.EnvConfig{
		NameOrPrefix: "PORT",
		Type:         horizonapi.EnvVal,
		KeyOrVal:     "3000",
	})
	synopsysOperatorUIContainer.AddEnv(horizonapi.EnvConfig{
		NameOrPrefix: "GO_ENV",
		Type:         horizonapi.EnvVal,
		KeyOrVal:     "development",
	})

	// Create config map volume
	var synopsysOperatorVolumeDefaultMode int32 = 420
	synopsysOperatorVolume := horizoncomponents.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "synopsys-operator",
		MapOrSecretName: "synopsys-operator",
		DefaultMode:     &synopsysOperatorVolumeDefaultMode,
	})

	synopsysOperatorPod.AddContainer(synopsysOperatorContainer)
	if specConfig.Expose != util.NONE && len(specConfig.Crds) > 0 && strings.Contains(strings.Join(specConfig.Crds, ","), util.BlackDuckCRDName) {
		synopsysOperatorPod.AddContainer(synopsysOperatorUIContainer)
	}
	synopsysOperatorPod.AddVolume(synopsysOperatorVolume)
	synopsysOperatorPod.AddLabels(map[string]string{"app": "synopsys-operator", "component": "operator"})

	synopsysOperatorTlSVolume := horizoncomponents.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "synopsys-operator-tls",
		MapOrSecretName: "synopsys-operator-tls",
		DefaultMode:     &synopsysOperatorVolumeDefaultMode,
	})
	synopsysOperatorPod.AddVolume(synopsysOperatorTlSVolume)

	// add temp log volume for glog issue
	synopsysOperatorLogVolume, err := horizoncomponents.NewEmptyDirVolume(horizonapi.EmptyDirVolumeConfig{
		VolumeName: "tmp-logs",
		Medium:     horizonapi.StorageMediumDefault,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}
	synopsysOperatorPod.AddVolume(synopsysOperatorLogVolume)

	synopsysOperator.AddPod(synopsysOperatorPod)

	synopsysOperator.AddLabels(map[string]string{"app": "synopsys-operator", "component": "operator"})
	return synopsysOperator, nil
}

// getOperatorService creates a exposed UI Service Horizon component for Synopsys Operaotor
func (specConfig *SpecConfig) getOperatorService() []*horizoncomponents.Service {

	services := []*horizoncomponents.Service{}
	if specConfig.Expose != util.NONE && len(specConfig.Crds) > 0 && strings.Contains(strings.Join(specConfig.Crds, ","), util.BlackDuckCRDName) {
		if strings.EqualFold(specConfig.Expose, util.NODEPORT) || strings.EqualFold(specConfig.Expose, util.LOADBALANCER) {

			var exposedServiceType horizonapi.ServiceType
			if strings.EqualFold(specConfig.Expose, util.NODEPORT) {
				exposedServiceType = horizonapi.ServiceTypeNodePort
			} else {
				exposedServiceType = horizonapi.ServiceTypeLoadBalancer
			}

			// Synopsys Operator UI exposed service
			synopsysOperatorExposedService := horizoncomponents.NewService(horizonapi.ServiceConfig{
				APIVersion: "v1",
				Name:       "synopsys-operator-exposed",
				Namespace:  specConfig.Namespace,
				Type:       exposedServiceType,
			})
			synopsysOperatorExposedService.AddSelectors(map[string]string{"app": "synopsys-operator", "component": "operator"})
			synopsysOperatorExposedService.AddPort(horizonapi.ServicePortConfig{
				Name:       "synopsys-operator-ui",
				Port:       80,
				TargetPort: "3000",
				Protocol:   horizonapi.ProtocolTCP,
			})
			synopsysOperatorExposedService.AddLabels(map[string]string{"app": "synopsys-operator", "component": "operator"})
			services = append(services, synopsysOperatorExposedService)
		}
	}

	return services
}

// GetOperatorConfigMap creates a ConfigMap Horizon component for Synopsys Operaotor
func (specConfig *SpecConfig) GetOperatorConfigMap() (*horizoncomponents.ConfigMap, error) {
	// Config Map
	synopsysOperatorConfigMap := horizoncomponents.NewConfigMap(horizonapi.ConfigMapConfig{
		APIVersion: "v1",
		Name:       "synopsys-operator",
		Namespace:  specConfig.Namespace,
	})

	cmData := map[string]string{}
	configData := map[string]interface{}{
		"Namespace":                     specConfig.Namespace,
		"Image":                         specConfig.Image,
		"Expose":                        specConfig.Expose,
		"ClusterType":                   specConfig.ClusterType,
		"DryRun":                        specConfig.DryRun,
		"LogLevel":                      specConfig.LogLevel,
		"Threadiness":                   specConfig.Threadiness,
		"PostgresRestartInMins":         specConfig.PostgresRestartInMins,
		"PodWaitTimeoutSeconds":         specConfig.PodWaitTimeoutSeconds,
		"ResyncIntervalInSeconds":       specConfig.ResyncIntervalInSeconds,
		"TerminationGracePeriodSeconds": specConfig.TerminationGracePeriodSeconds,
		"AdmissionWebhookListener":      specConfig.AdmissionWebhookListener,
		"CrdNames":                      strings.Join(specConfig.Crds, ","),
		"IsClusterScoped":               specConfig.IsClusterScoped,
	}
	bytes, err := json.Marshal(configData)
	if err != nil {
		return nil, errors.Trace(err)
	}

	cmData["config.json"] = string(bytes)
	synopsysOperatorConfigMap.AddData(cmData)

	synopsysOperatorConfigMap.AddLabels(map[string]string{"app": "synopsys-operator", "component": "operator"})
	return synopsysOperatorConfigMap, nil
}

// getOperatorServiceAccount creates a ServiceAccount Horizon component for Synopsys Operaotor
func (specConfig *SpecConfig) getOperatorServiceAccount() *horizoncomponents.ServiceAccount {
	// Service Account
	synopsysOperatorServiceAccount := horizoncomponents.NewServiceAccount(horizonapi.ServiceAccountConfig{
		APIVersion: "v1",
		Name:       "synopsys-operator",
		Namespace:  specConfig.Namespace,
	})

	synopsysOperatorServiceAccount.AddLabels(map[string]string{"app": "synopsys-operator", "component": "operator"})
	return synopsysOperatorServiceAccount
}

// getOperatorClusterRoleBinding creates a ClusterRoleBinding Horizon component for Synopsys Operaotor
func (specConfig *SpecConfig) getOperatorClusterRoleBinding() *horizoncomponents.ClusterRoleBinding {
	// Cluster Role Binding
	synopsysOperatorClusterRoleBinding := horizoncomponents.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		APIVersion: "rbac.authorization.k8s.io/v1beta1",
		Name:       "synopsys-operator-admin",
		Namespace:  specConfig.Namespace,
	})
	synopsysOperatorClusterRoleBinding.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      "synopsys-operator",
		Namespace: specConfig.Namespace,
	})
	synopsysOperatorClusterRoleBinding.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "ClusterRole",
		Name:     "synopsys-operator-admin",
	})

	synopsysOperatorClusterRoleBinding.AddLabels(map[string]string{"app": "synopsys-operator", "component": "operator"})
	return synopsysOperatorClusterRoleBinding
}

// getOperatorRoleBinding creates a RoleBinding Horizon component for Synopsys Operator
func (specConfig *SpecConfig) getOperatorRoleBinding() *horizoncomponents.RoleBinding {
	// Role Binding
	synopsysOperatorRoleBinding := horizoncomponents.NewRoleBinding(horizonapi.RoleBindingConfig{
		APIVersion: "rbac.authorization.k8s.io/v1beta1",
		Name:       "synopsys-operator-admin",
		Namespace:  specConfig.Namespace,
	})
	synopsysOperatorRoleBinding.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      "synopsys-operator",
		Namespace: specConfig.Namespace,
	})
	synopsysOperatorRoleBinding.AddRoleRef(horizonapi.RoleRefConfig{
		APIGroup: "",
		Kind:     "Role",
		Name:     "synopsys-operator-admin",
	})

	synopsysOperatorRoleBinding.AddLabels(map[string]string{"app": "synopsys-operator", "component": "operator"})
	return synopsysOperatorRoleBinding
}

// getOperatorClusterRole creates a ClusterRole Horizon component for the Synopsys Operator
func (specConfig *SpecConfig) getOperatorClusterRole() *horizoncomponents.ClusterRole {
	synopsysOperatorClusterRole := horizoncomponents.NewClusterRole(horizonapi.ClusterRoleConfig{
		APIVersion: "rbac.authorization.k8s.io/v1beta1",
		Name:       "synopsys-operator-admin",
		Namespace:  specConfig.Namespace,
	})

	synopsysOperatorClusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		Verbs:     []string{"get", "list"},
		APIGroups: []string{"apiextensions.k8s.io"},
		Resources: []string{"customresourcedefinitions"},
	})

	synopsysOperatorClusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		Verbs:     []string{"get", "list", "create", "update", "patch", "delete"},
		APIGroups: []string{"rbac.authorization.k8s.io"},
		Resources: []string{"clusterrolebindings", "clusterroles"},
	})

	synopsysOperatorClusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		Verbs:     []string{"get", "list", "create", "update", "patch", "delete"},
		APIGroups: []string{"rbac.authorization.k8s.io"},
		Resources: []string{"rolebindings", "roles"},
	})

	synopsysOperatorClusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		Verbs:     []string{"get", "list", "create", "update", "patch", "delete"},
		APIGroups: []string{"apps"},
		Resources: []string{"deployments"},
	})

	synopsysOperatorClusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		Verbs:     []string{"get", "list", "create", "update", "patch", "delete"},
		APIGroups: []string{""},
		Resources: []string{"namespaces", "persistentvolumeclaims", "services", "replicationcontrollers", "serviceaccounts"},
	})

	synopsysOperatorClusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
		APIGroups: []string{""},
		Resources: []string{"configmaps", "secrets"},
	})

	synopsysOperatorClusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		Verbs:     []string{"get", "list", "update"},
		APIGroups: []string{""},
		Resources: []string{"pods"},
	})

	if len(specConfig.Crds) > 0 && strings.Contains(strings.Join(specConfig.Crds, ","), util.OpsSightCRDName) {
		synopsysOperatorClusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
			Verbs:     []string{"watch"},
			APIGroups: []string{""},
			Resources: []string{"pods"},
		})
	}

	synopsysOperatorClusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		Verbs:     []string{"create"},
		APIGroups: []string{""},
		Resources: []string{"pods/exec"},
	})

	synopsysOperatorClusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete", "deletecollection"},
		APIGroups: []string{"synopsys.com"},
		Resources: []string{"*"},
	})

	if specConfig.Expose != util.NONE && len(specConfig.Crds) > 0 && strings.Contains(strings.Join(specConfig.Crds, ","), util.BlackDuckCRDName) {
		synopsysOperatorClusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
			Verbs:     []string{"get", "list"},
			APIGroups: []string{"storage.k8s.io"},
			Resources: []string{"storageclasses"},
		})
	}

	// Add Openshift rules
	if specConfig.ClusterType == OpenshiftClusterType {
		synopsysOperatorClusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
			Verbs:     []string{"get", "update", "patch"},
			APIGroups: []string{"security.openshift.io"},
			Resources: []string{"securitycontextconstraints"},
		})

		synopsysOperatorClusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
			Verbs:     []string{"get", "list", "create", "delete"},
			APIGroups: []string{"route.openshift.io"},
			Resources: []string{"routes"},
		})

		synopsysOperatorClusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
			Verbs:     []string{"get", "list", "watch", "update"},
			APIGroups: []string{"image.openshift.io"},
			Resources: []string{"images"},
		})

		// add layers rule to pull the image from OpenShift internal registry
		if len(specConfig.Crds) > 0 && strings.Contains(strings.Join(specConfig.Crds, ","), util.OpsSightCRDName) {
			synopsysOperatorClusterRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
				Verbs:     []string{"get"},
				APIGroups: []string{"image.openshift.io", ""},
				Resources: []string{"imagestreams/layers"},
			})
		}
	} else { // Kube or Error
		log.Debug("Skipping Openshift Cluster Role Rules")
	}

	synopsysOperatorClusterRole.AddLabels(map[string]string{"app": "synopsys-operator", "component": "operator"})
	return synopsysOperatorClusterRole
}

// getOperatorRole creates a Role Horizon component for Synopsys Operaotor
func (specConfig *SpecConfig) getOperatorRole() *horizoncomponents.Role {
	synopsysOperatorRole := horizoncomponents.NewRole(horizonapi.RoleConfig{
		APIVersion: "rbac.authorization.k8s.io/v1beta1",
		Name:       "synopsys-operator-admin",
		Namespace:  specConfig.Namespace,
	})

	// TODO: is watch/deletecollection needed currently?
	synopsysOperatorRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		Verbs:     []string{"get", "list", "create", "update", "patch", "delete"},
		APIGroups: []string{"rbac.authorization.k8s.io"},
		Resources: []string{"rolebindings", "roles"},
	})

	synopsysOperatorRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		Verbs:     []string{"get", "list", "create", "update", "patch", "delete"},
		APIGroups: []string{"apps"},
		Resources: []string{"deployments"},
	})

	synopsysOperatorRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		Verbs:     []string{"get", "update", "patch"},
		APIGroups: []string{""},
		Resources: []string{"namespaces"},
	})

	synopsysOperatorRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		Verbs:     []string{"get", "list", "create", "update", "patch", "delete"},
		APIGroups: []string{""},
		Resources: []string{"persistentvolumeclaims", "services", "replicationcontrollers", "replicationcontrollers/scale", "serviceaccounts"},
	})

	synopsysOperatorRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
		APIGroups: []string{""},
		Resources: []string{"configmaps", "secrets"},
	})

	// TODO: is logs needed?
	// TODO: is endpoints needed?
	synopsysOperatorRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		Verbs:     []string{"get", "list", "update"},
		APIGroups: []string{""},
		Resources: []string{"pods"},
	})

	synopsysOperatorRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		Verbs:     []string{"create"},
		APIGroups: []string{""},
		Resources: []string{"pods/exec"},
	})

	synopsysOperatorRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
		Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete", "deletecollection"},
		APIGroups: []string{"synopsys.com"},
		Resources: []string{"*"},
	})

	if specConfig.Expose != util.NONE && len(specConfig.Crds) > 0 && strings.Contains(strings.Join(specConfig.Crds, ","), util.BlackDuckCRDName) {
		synopsysOperatorRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
			Verbs:     []string{"get", "list"},
			APIGroups: []string{"storage.k8s.io"},
			Resources: []string{"storageclasses"},
		})
	}

	// Add Openshift rules
	if specConfig.ClusterType == OpenshiftClusterType {

		synopsysOperatorRole.AddPolicyRule(horizonapi.PolicyRuleConfig{
			Verbs:     []string{"get", "list", "create", "delete"},
			APIGroups: []string{"route.openshift.io"},
			Resources: []string{"routes"},
		})

	} else { // Kube or Error
		log.Debug("Skipping Openshift Cluster Role Rules")
	}

	synopsysOperatorRole.AddLabels(map[string]string{"app": "synopsys-operator", "component": "operator"})
	return synopsysOperatorRole
}

// GetTLSCertificateSecret creates a TLS certificate in horizon format
func (specConfig *SpecConfig) GetTLSCertificateSecret() *horizoncomponents.Secret {
	tlsSecret := horizoncomponents.NewSecret(horizonapi.SecretConfig{
		Name:      "synopsys-operator-tls",
		Namespace: specConfig.Namespace,
		Type:      horizonapi.SecretTypeOpaque,
	})
	tlsSecret.AddData(map[string][]byte{
		"cert.crt": []byte(specConfig.Certificate),
		"cert.key": []byte(specConfig.CertificateKey),
	})

	tlsSecret.AddLabels(map[string]string{"app": "synopsys-operator", "component": "operator"})
	return tlsSecret
}

// getOperatorSecret creates a Secret Horizon component for Synopsys Operaotor
func (specConfig *SpecConfig) getOperatorSecret() *horizoncomponents.Secret {
	// create a secret
	synopsysOperatorSecret := horizoncomponents.NewSecret(horizonapi.SecretConfig{
		APIVersion: "v1",
		Name:       "blackduck-secret",
		Namespace:  specConfig.Namespace,
		Type:       horizonapi.SecretTypeOpaque,
	})
	synopsysOperatorSecret.AddData(map[string][]byte{
		"SEAL_KEY": []byte(specConfig.SealKey),
	})

	synopsysOperatorSecret.AddLabels(map[string]string{"app": "synopsys-operator", "component": "operator"})
	return synopsysOperatorSecret
}

// getOpenShiftRoute creates the OpenShift route component for Synopsys Operator
func (specConfig *SpecConfig) getOpenShiftRoute() *api.Route {
	if strings.ToUpper(specConfig.Expose) == util.OPENSHIFT {
		return &api.Route{
			Name:               "synopsys-operator-ui",
			Namespace:          specConfig.Namespace,
			Kind:               "Service",
			ServiceName:        "synopsys-operator",
			PortName:           "synopsys-operator-ui",
			Labels:             map[string]string{"app": "synopsys-operator", "component": "operator"},
			TLSTerminationType: routev1.TLSTerminationEdge,
		}
	}
	return nil
}
