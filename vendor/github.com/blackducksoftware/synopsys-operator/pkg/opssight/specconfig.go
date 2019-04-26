/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownershia. The ASF licenses this file
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
	"encoding/json"
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	hubclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/juju/errors"
	"k8s.io/client-go/kubernetes"
)

// SpecConfig will contain the specification of OpsSight
type SpecConfig struct {
	config         *protoform.Config
	kubeClient     *kubernetes.Clientset
	opssightClient *opssightclientset.Clientset
	hubClient      *hubclientset.Clientset
	opssight       *opssightapi.OpsSight
	configMap      *MainOpssightConfigMap
	dryRun         bool
}

// NewSpecConfig will create the OpsSight object
func NewSpecConfig(config *protoform.Config, kubeClient *kubernetes.Clientset, opssightClient *opssightclientset.Clientset, hubClient *hubclientset.Clientset, opssight *opssightapi.OpsSight, dryRun bool) *SpecConfig {
	opssightSpec := &opssight.Spec
	configMap := &MainOpssightConfigMap{
		LogLevel: opssightSpec.LogLevel,
		BlackDuck: &BlackDuckConfig{
			ConnectionsEnvironmentVariableName: opssightSpec.Blackduck.ConnectionsEnvironmentVariableName,
			TLSVerification:                    opssightSpec.Blackduck.TLSVerification,
		},
		ImageFacade: &ImageFacadeConfig{
			CreateImagesOnly: false,
			Host:             "localhost",
			Port:             opssightSpec.ScannerPod.ImageFacade.Port,
			ImagePullerType:  opssightSpec.ScannerPod.ImageFacade.ImagePullerType,
		},
		Perceiver: &PerceiverConfig{
			Image: &ImagePerceiverConfig{},
			Pod: &PodPerceiverConfig{
				NamespaceFilter: opssightSpec.Perceiver.PodPerceiver.NamespaceFilter,
			},
			AnnotationIntervalSeconds: opssightSpec.Perceiver.AnnotationIntervalSeconds,
			DumpIntervalMinutes:       opssightSpec.Perceiver.DumpIntervalMinutes,
			Port:                      opssightSpec.Perceiver.Port,
		},
		Perceptor: &PerceptorConfig{
			Timings: &PerceptorTimingsConfig{
				CheckForStalledScansPauseHours: opssightSpec.Perceptor.CheckForStalledScansPauseHours,
				ClientTimeoutMilliseconds:      opssightSpec.Perceptor.ClientTimeoutMilliseconds,
				ModelMetricsPauseSeconds:       opssightSpec.Perceptor.ModelMetricsPauseSeconds,
				StalledScanClientTimeoutHours:  opssightSpec.Perceptor.StalledScanClientTimeoutHours,
				UnknownImagePauseMilliseconds:  opssightSpec.Perceptor.UnknownImagePauseMilliseconds,
			},
			Host:        opssightSpec.Perceptor.Name,
			Port:        opssightSpec.Perceptor.Port,
			UseMockMode: false,
		},
		Scanner: &ScannerConfig{
			BlackDuckClientTimeoutSeconds: opssightSpec.ScannerPod.Scanner.ClientTimeoutSeconds,
			ImageDirectory:                opssightSpec.ScannerPod.ImageDirectory,
			Port:                          opssightSpec.ScannerPod.Scanner.Port,
		},
		Skyfire: &SkyfireConfig{
			BlackDuckClientTimeoutSeconds: opssightSpec.Skyfire.HubClientTimeoutSeconds,
			BlackDuckDumpPauseSeconds:     opssightSpec.Skyfire.HubDumpPauseSeconds,
			KubeDumpIntervalSeconds:       opssightSpec.Skyfire.KubeDumpIntervalSeconds,
			PerceptorDumpIntervalSeconds:  opssightSpec.Skyfire.PerceptorDumpIntervalSeconds,
			Port:                          opssightSpec.Skyfire.Port,
			PrometheusPort:                opssightSpec.Skyfire.PrometheusPort,
			UseInClusterConfig:            true,
		},
	}
	return &SpecConfig{config: config, kubeClient: kubeClient, opssightClient: opssightClient, hubClient: hubClient, opssight: opssight, configMap: configMap, dryRun: dryRun}
}

func (p *SpecConfig) configMapVolume(volumeName string) *components.Volume {
	return components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      volumeName,
		MapOrSecretName: p.opssight.Spec.ConfigMapName,
	})
}

// GetComponents will return the list of components
func (p *SpecConfig) GetComponents() (*api.ComponentList, error) {
	components := &api.ComponentList{}

	// Add config map
	cm, err := p.configMap.horizonConfigMap(
		p.opssight.Spec.ConfigMapName,
		p.opssight.Spec.Namespace,
		fmt.Sprintf("%s.json", p.opssight.Spec.ConfigMapName))
	if err != nil {
		return nil, errors.Trace(err)
	}
	components.ConfigMaps = append(components.ConfigMaps, cm)

	// Add Perceptor
	rc, err := p.PerceptorReplicationController()
	if err != nil {
		return nil, errors.Trace(err)
	}
	components.ReplicationControllers = append(components.ReplicationControllers, rc)
	service, err := p.PerceptorService()
	if err != nil {
		return nil, errors.Trace(err)
	}
	components.Services = append(components.Services, service)
	secret := p.PerceptorSecret()
	if !p.dryRun {
		p.addSecretData(secret)
	}
	components.Secrets = append(components.Secrets, secret)

	// Add Perceptor Scanner
	scannerRC, err := p.ScannerReplicationController()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create scanner replication controller")
	}
	components.ReplicationControllers = append(components.ReplicationControllers, scannerRC)
	components.Services = append(components.Services, p.ScannerService(), p.ImageFacadeService())

	components.ServiceAccounts = append(components.ServiceAccounts, p.ScannerServiceAccount())
	components.ClusterRoleBindings = append(components.ClusterRoleBindings, p.ScannerClusterRoleBinding())

	// Add Pod Perceiver
	if p.opssight.Spec.Perceiver.EnablePodPerceiver {
		rc, err = p.PodPerceiverReplicationController()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create pod perceiver")
		}
		components.ReplicationControllers = append(components.ReplicationControllers, rc)
		components.Services = append(components.Services, p.PodPerceiverService())
		components.ServiceAccounts = append(components.ServiceAccounts, p.PodPerceiverServiceAccount())
		podClusterRole := p.PodPerceiverClusterRole()
		components.ClusterRoles = append(components.ClusterRoles, podClusterRole)
		components.ClusterRoleBindings = append(components.ClusterRoleBindings, p.PodPerceiverClusterRoleBinding(podClusterRole))
	}

	// Add Image Perceiver
	if p.opssight.Spec.Perceiver.EnableImagePerceiver {
		rc, err = p.ImagePerceiverReplicationController()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create image perceiver")
		}
		components.ReplicationControllers = append(components.ReplicationControllers, rc)
		components.Services = append(components.Services, p.ImagePerceiverService())
		components.ServiceAccounts = append(components.ServiceAccounts, p.ImagePerceiverServiceAccount())
		imageClusterRole := p.ImagePerceiverClusterRole()
		components.ClusterRoles = append(components.ClusterRoles, imageClusterRole)
		components.ClusterRoleBindings = append(components.ClusterRoleBindings, p.ImagePerceiverClusterRoleBinding(imageClusterRole))
	}

	// Add skyfire
	if p.opssight.Spec.EnableSkyfire {
		skyfireRC, err := p.PerceptorSkyfireReplicationController()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create skyfire")
		}
		components.ReplicationControllers = append(components.ReplicationControllers, skyfireRC)
		components.Services = append(components.Services, p.PerceptorSkyfireService())
		components.ServiceAccounts = append(components.ServiceAccounts, p.PerceptorSkyfireServiceAccount())
		skyfireClusterRole := p.PerceptorSkyfireClusterRole()
		components.ClusterRoles = append(components.ClusterRoles, skyfireClusterRole)
		components.ClusterRoleBindings = append(components.ClusterRoleBindings, p.PerceptorSkyfireClusterRoleBinding(skyfireClusterRole))
	}

	// Add Metrics
	if p.opssight.Spec.EnableMetrics {
		dep, err := p.PerceptorMetricsDeployment()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create metrics")
		}
		components.Deployments = append(components.Deployments, dep)
		components.Services = append(components.Services, p.PerceptorMetricsService())
		perceptorCm, err := p.PerceptorMetricsConfigMap()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create perceptor config map")
		}
		components.ConfigMaps = append(components.ConfigMaps, perceptorCm)
	}

	return components, nil
}

func (p *SpecConfig) addSecretData(secret *components.Secret) error {
	blackduckHosts := make(map[string]*opssightapi.Host)
	// adding External Black Duck credentials
	for _, host := range p.opssight.Spec.Blackduck.ExternalHosts {
		blackduckHosts[host.Domain] = host
	}

	// adding Internal Black Duck credentials
	configMapEditor := NewUpdater(p.config, p.kubeClient, p.hubClient, p.opssightClient)
	allHubs := configMapEditor.getAllHubs(p.opssight.Spec.Blackduck.BlackduckSpec.Type)
	blackduckPasswords := configMapEditor.appendBlackDuckSecrets(blackduckHosts, allHubs)

	// marshal the blackduck credentials to bytes
	bytes, err := json.Marshal(blackduckPasswords)
	if err != nil {
		return errors.Annotatef(err, "unable to marshal blackduck passwords")
	}
	secret.AddData(map[string][]byte{p.opssight.Spec.Blackduck.ConnectionsEnvironmentVariableName: bytes})

	// adding Secured registries credentials
	securedRegistries := make(map[string]*opssightapi.RegistryAuth)
	for _, internalRegistry := range p.opssight.Spec.ScannerPod.ImageFacade.InternalRegistries {
		securedRegistries[internalRegistry.URL] = internalRegistry
	}
	// marshal the Secured registries credentials to bytes
	bytes, err = json.Marshal(securedRegistries)
	if err != nil {
		return errors.Annotatef(err, "unable to marshal secured registries")
	}
	secret.AddData(map[string][]byte{"securedRegistries.json": bytes})

	// add internal hosts to status
	p.opssight.Status.InternalHosts = configMapEditor.appendBlackDuckHosts(p.opssight.Status.InternalHosts, allHubs)
	return nil
}
