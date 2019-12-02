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
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	appsutil "github.com/blackducksoftware/synopsys-operator/pkg/apps/util"
	hubclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	"k8s.io/client-go/kubernetes"
)

// SpecConfig will contain the specification of OpsSight
type SpecConfig struct {
	config                  *protoform.Config
	kubeClient              *kubernetes.Clientset
	opssightClient          *opssightclientset.Clientset
	hubClient               *hubclientset.Clientset
	opssight                *opssightapi.OpsSight
	configMap               *MainOpssightConfigMap
	names                   map[string]string
	images                  map[string]string
	isBlackDuckClusterScope bool
	dryRun                  bool
}

// NewSpecConfig will create the OpsSight object
func NewSpecConfig(config *protoform.Config, kubeClient *kubernetes.Clientset, opssightClient *opssightclientset.Clientset, hubClient *hubclientset.Clientset, opssight *opssightapi.OpsSight, isBlackDuckClusterScope bool, dryRun bool) *SpecConfig {
	opssightSpec := &opssight.Spec
	name := opssight.Name
	names := map[string]string{
		"perceptor":                 "core",
		"pod-perceiver":             "pod-processor",
		"image-perceiver":           "image-processor",
		"artifactory-perceiver":     "artifactory-processor",
		"quay-perceiver":            "quay-processor",
		"scanner":                   "scanner",
		"perceptor-imagefacade":     "image-getter",
		"skyfire":                   "skyfire",
		"prometheus":                "prometheus",
		"configmap":                 "opssight",
		"perceiver-service-account": "processor",
	}
	baseImageURL := "docker.io/blackducksoftware"
	version := "2.2.4"
	images := map[string]string{
		"perceptor":             fmt.Sprintf("%s/opssight-core:%s", baseImageURL, version),
		"pod-perceiver":         fmt.Sprintf("%s/opssight-pod-processor:%s", baseImageURL, version),
		"image-perceiver":       fmt.Sprintf("%s/opssight-image-processor:%s", baseImageURL, version),
		"artifactory-perceiver": fmt.Sprintf("%s/opssight-artifactory-processor:%s", baseImageURL, version),
		"quay-perceiver":        fmt.Sprintf("%s/opssight-quay-processor:%s", baseImageURL, version),
		"scanner":               fmt.Sprintf("%s/opssight-scanner:%s", baseImageURL, version),
		"perceptor-imagefacade": fmt.Sprintf("%s/opssight-image-getter:%s", baseImageURL, version),
		"skyfire":               "gcr.io/saas-hub-stg/blackducksoftware/pyfire:master",
		"prometheus":            "docker.io/prom/prometheus:v2.1.0",
	}
	if opssightSpec.IsUpstream {
		names = map[string]string{
			"perceptor":                 "perceptor",
			"pod-perceiver":             "pod-perceiver",
			"image-perceiver":           "image-perceiver",
			"artifactory-perceiver":     "artifactory-perceiver",
			"quay-perceiver":            "quay-perceiver",
			"scanner":                   "scanner",
			"perceptor-imagefacade":     "image-facade",
			"skyfire":                   "skyfire",
			"prometheus":                "prometheus",
			"configmap":                 "perceptor",
			"perceiver-service-account": "perceiver",
		}
		baseImageURL = "gcr.io/saas-hub-stg/blackducksoftware"
		version = "master"
		images = map[string]string{
			"perceptor":             fmt.Sprintf("%s/perceptor:%s", baseImageURL, version),
			"pod-perceiver":         fmt.Sprintf("%s/pod-perceiver:%s", baseImageURL, version),
			"image-perceiver":       fmt.Sprintf("%s/image-perceiver:%s", baseImageURL, version),
			"artifactory-perceiver": fmt.Sprintf("%s/artifactory-perceiver:%s", baseImageURL, version),
			"quay-perceiver":        fmt.Sprintf("%s/quay-perceiver:%s", baseImageURL, version),
			"scanner":               fmt.Sprintf("%s/perceptor-scanner:%s", baseImageURL, version),
			"perceptor-imagefacade": fmt.Sprintf("%s/perceptor-imagefacade:%s", baseImageURL, version),
			"skyfire":               "gcr.io/saas-hub-stg/blackducksoftware/pyfire:master",
			"prometheus":            "docker.io/prom/prometheus:v2.1.0"}
	}

	for componentName, componentImage := range images {
		image := appsutil.GenerateImageTag(componentImage, opssightSpec.ImageRegistries, opssightSpec.RegistryConfiguration)
		images[componentName] = image
	}

	configMap := &MainOpssightConfigMap{
		LogLevel: opssightSpec.LogLevel,
		BlackDuck: &BlackDuckConfig{
			ConnectionsEnvironmentVariableName: opssightSpec.Blackduck.ConnectionsEnvironmentVariableName,
			TLSVerification:                    opssightSpec.Blackduck.TLSVerification,
		},
		ImageFacade: &ImageFacadeConfig{
			CreateImagesOnly: false,
			Host:             "localhost",
			Port:             3004,
			ImagePullerType:  opssightSpec.ScannerPod.ImageFacade.ImagePullerType,
		},
		Perceiver: &PerceiverConfig{
			Certificate:    opssightSpec.Perceiver.Certificate,
			CertificateKey: opssightSpec.Perceiver.CertificateKey,
			Image:          &ImagePerceiverConfig{},
			Pod: &PodPerceiverConfig{
				NamespaceFilter: opssightSpec.Perceiver.PodPerceiver.NamespaceFilter,
			},
			Artifactory: &ArtifactoryPerceiverConfig{
				Dumper: opssightSpec.Perceiver.EnableArtifactoryPerceiverDumper,
			},
			AnnotationIntervalSeconds: opssightSpec.Perceiver.AnnotationIntervalSeconds,
			DumpIntervalMinutes:       opssightSpec.Perceiver.DumpIntervalMinutes,
			Port:                      3002,
		},
		Perceptor: &PerceptorConfig{
			Timings: &PerceptorTimingsConfig{
				CheckForStalledScansPauseHours: opssightSpec.Perceptor.CheckForStalledScansPauseHours,
				ClientTimeoutMilliseconds:      opssightSpec.Perceptor.ClientTimeoutMilliseconds,
				ModelMetricsPauseSeconds:       opssightSpec.Perceptor.ModelMetricsPauseSeconds,
				StalledScanClientTimeoutHours:  opssightSpec.Perceptor.StalledScanClientTimeoutHours,
				UnknownImagePauseMilliseconds:  opssightSpec.Perceptor.UnknownImagePauseMilliseconds,
			},
			Host:        util.GetResourceName(name, util.OpsSightName, names["perceptor"]),
			Port:        3001,
			UseMockMode: false,
		},
		Scanner: &ScannerConfig{
			BlackDuckClientTimeoutSeconds: opssightSpec.ScannerPod.Scanner.ClientTimeoutSeconds,
			ImageDirectory:                opssightSpec.ScannerPod.ImageDirectory,
			Port:                          3003,
		},
		Skyfire: &SkyfireConfig{
			BlackDuckClientTimeoutSeconds: opssightSpec.Skyfire.HubClientTimeoutSeconds,
			BlackDuckDumpPauseSeconds:     opssightSpec.Skyfire.HubDumpPauseSeconds,
			KubeDumpIntervalSeconds:       opssightSpec.Skyfire.KubeDumpIntervalSeconds,
			PerceptorDumpIntervalSeconds:  opssightSpec.Skyfire.PerceptorDumpIntervalSeconds,
			Port:                          3005,
			PrometheusPort:                3006,
			UseInClusterConfig:            true,
		},
	}
	return &SpecConfig{
		config:                  config,
		kubeClient:              kubeClient,
		opssightClient:          opssightClient,
		hubClient:               hubClient,
		opssight:                opssight,
		configMap:               configMap,
		isBlackDuckClusterScope: isBlackDuckClusterScope,
		dryRun:                  dryRun,
		names:                   names,
		images:                  images,
	}
}

func (p *SpecConfig) configMapVolume(volumeName string) *components.Volume {
	return components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      volumeName,
		MapOrSecretName: util.GetResourceName(p.opssight.Name, util.OpsSightName, p.names["configmap"]),
		DefaultMode:     util.IntToInt32(420),
	})
}

// GetComponents will return the list of components
func (p *SpecConfig) GetComponents() (*api.ComponentList, error) {
	components := &api.ComponentList{}
	name := p.opssight.Name
	// Add config map
	cm, err := p.configMap.horizonConfigMap(util.GetResourceName(name, util.OpsSightName, p.names["configmap"]), p.opssight.Spec.Namespace, fmt.Sprintf("%s.json", p.names["configmap"]))
	if err != nil {
		return nil, errors.Trace(err)
	}
	cm.AddLabels(map[string]string{"name": name})
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
	perceptorSvc, err := p.getPerceptorExposeService()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create perceptor service")
	}
	if perceptorSvc != nil {
		components.Services = append(components.Services, perceptorSvc)
	}
	secret, err := p.PerceptorSecret()
	if err != nil {
		return nil, errors.Trace(err)
	}
	if !p.dryRun {
		p.addSecretData(secret)
	}
	components.Secrets = append(components.Secrets, secret)

	route := p.GetPerceptorOpenShiftRoute()
	if route != nil {
		components.Routes = append(components.Routes, route)
	}

	// Add Perceptor Scanner
	scannerRC, err := p.ScannerReplicationController()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create scanner replication controller")
	}
	components.ReplicationControllers = append(components.ReplicationControllers, scannerRC)
	components.Services = append(components.Services, p.ScannerService(), p.ImageFacadeService())

	components.ServiceAccounts = append(components.ServiceAccounts, p.ScannerServiceAccount())
	if p.config.IsOpenshift && p.opssight.Spec.Perceiver.EnableImagePerceiver {
		clusterRoleBinding, err := p.ScannerClusterRoleBinding()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create scanner cluster role binding")
		}
		components.ClusterRoleBindings = append(components.ClusterRoleBindings, clusterRoleBinding)
	}

	// Add Pod Perceiver
	if p.opssight.Spec.Perceiver.EnablePodPerceiver {
		rc, err = p.PodPerceiverReplicationController()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create pod perceiver")
		}
		components.ReplicationControllers = append(components.ReplicationControllers, rc)
		components.Services = append(components.Services, p.PodPerceiverService())
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
		imageClusterRole := p.ImagePerceiverClusterRole()
		components.ClusterRoles = append(components.ClusterRoles, imageClusterRole)
		components.ClusterRoleBindings = append(components.ClusterRoleBindings, p.ImagePerceiverClusterRoleBinding(imageClusterRole))
	}

	// Add Artifactory Perceiver if enabled
	if p.opssight.Spec.Perceiver.EnableArtifactoryPerceiver {
		rc, err = p.ArtifactoryPerceiverReplicationController()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create artifactory perceiver")
		}

		components.ReplicationControllers = append(components.ReplicationControllers, rc)
		perceiverSvc, err := p.getPerceiverExposeService("artifactory")
		if err != nil {
			return nil, errors.Annotate(err, "failed to create artifactory perceiver service")
		}
		if perceiverSvc != nil {
			components.Services = append(components.Services, perceiverSvc)
		}
		secure := false
		if len(p.opssight.Spec.Perceiver.Certificate) > 0 && len(p.opssight.Spec.Perceiver.CertificateKey) > 0 {
			secure = true
		}
		route := p.GetPerceiverOpenShiftRoute("artifactory", secure)
		if route != nil {
			components.Services = append(components.Services, p.ArtifactoryPerceiverService())
			components.Routes = append(components.Routes, route)
		}
	}

	// Add Quay Perceiver if enabled
	if p.opssight.Spec.Perceiver.EnableQuayPerceiver {
		rc, err = p.QuayPerceiverReplicationController()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create quay perceiver")
		}

		components.ReplicationControllers = append(components.ReplicationControllers, rc)
		perceiverSvc, err := p.getPerceiverExposeService("quay")
		if err != nil {
			return nil, errors.Annotate(err, "failed to create quay perceiver service")
		}
		if perceiverSvc != nil {
			components.Services = append(components.Services, perceiverSvc)
		}
		secure := false
		if len(p.opssight.Spec.Perceiver.Certificate) > 0 && len(p.opssight.Spec.Perceiver.CertificateKey) > 0 {
			secure = true
		}
		route := p.GetPerceiverOpenShiftRoute("quay", secure)
		if route != nil {
			components.Services = append(components.Services, p.QuayPerceiverService())
			components.Routes = append(components.Routes, route)
		}
	}

	if p.opssight.Spec.Perceiver.EnablePodPerceiver || p.opssight.Spec.Perceiver.EnableImagePerceiver {
		// Use the same service account
		//components.ServiceAccounts = append(components.ServiceAccounts, p.PodPerceiverServiceAccount())
		components.ServiceAccounts = append(components.ServiceAccounts, p.ImagePerceiverServiceAccount())
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
		// deployments
		dep, err := p.PerceptorMetricsDeployment()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create metrics")
		}
		components.Deployments = append(components.Deployments, dep)

		// services
		prometheusService, err := p.PerceptorMetricsService()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create prometheus metrics service")
		}
		components.Services = append(components.Services, prometheusService)
		prometheusSvc, err := p.getPerceptorMetricsExposeService()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create prometheus metrics exposed service")
		}
		if prometheusSvc != nil {
			components.Services = append(components.Services, prometheusSvc)
		}

		// config map
		perceptorCm, err := p.PerceptorMetricsConfigMap()
		if err != nil {
			return nil, errors.Annotate(err, "failed to create perceptor config map")
		}
		components.ConfigMaps = append(components.ConfigMaps, perceptorCm)

		route := p.GetPrometheusOpenShiftRoute()
		if route != nil {
			components.Routes = append(components.Routes, route)
		}
	}

	return components, nil
}

func (p *SpecConfig) getPerceiverExposeService(perceiverName string) (*components.Service, error) {
	var svc *components.Service
	var err error
	switch strings.ToUpper(p.opssight.Spec.Perceiver.Expose) {
	case util.NODEPORT:
		svc, err = p.PerceiverNodePortService(perceiverName)
		break
	case util.LOADBALANCER:
		svc, err = p.PerceiverLoadBalancerService(perceiverName)
		break
	default:
	}
	return svc, err
}

func (p *SpecConfig) getPerceptorExposeService() (*components.Service, error) {
	var svc *components.Service
	var err error
	switch strings.ToUpper(p.opssight.Spec.Perceptor.Expose) {
	case util.NODEPORT:
		svc, err = p.PerceptorNodePortService()
		break
	case util.LOADBALANCER:
		svc, err = p.PerceptorLoadBalancerService()
		break
	default:
	}
	return svc, err
}

func (p *SpecConfig) getPerceptorMetricsExposeService() (*components.Service, error) {
	var svc *components.Service
	var err error
	switch strings.ToUpper(p.opssight.Spec.Prometheus.Expose) {
	case util.NODEPORT:
		svc, err = p.PerceptorMetricsNodePortService()
		break
	case util.LOADBALANCER:
		svc, err = p.PerceptorMetricsLoadBalancerService()
		break
	default:
	}
	return svc, err
}

func (p *SpecConfig) addSecretData(secret *components.Secret) error {
	blackduckHosts := make(map[string]*opssightapi.Host)
	// adding External Black Duck credentials
	for _, host := range p.opssight.Spec.Blackduck.ExternalHosts {
		blackduckHosts[host.Domain] = host
	}

	// adding Internal Black Duck credentials
	secretEditor := NewUpdater(p.config, p.kubeClient, p.hubClient, p.opssightClient)
	hubType := p.opssight.Spec.Blackduck.BlackduckSpec.Type
	blackduckPassword, err := util.Base64Decode(p.opssight.Spec.Blackduck.BlackduckPassword)
	if err != nil {
		return errors.Annotatef(err, "unable to decode blackduckPassword")
	}

	allHubs := secretEditor.getAllHubs(hubType, blackduckPassword)
	blackduckPasswords := secretEditor.appendBlackDuckSecrets(blackduckHosts, p.opssight.Status.InternalHosts, allHubs)

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
	p.opssight.Status.InternalHosts = secretEditor.appendBlackDuckHosts(p.opssight.Status.InternalHosts, allHubs)
	return nil
}
