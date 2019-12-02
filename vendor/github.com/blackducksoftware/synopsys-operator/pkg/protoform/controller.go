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

package protoform

import (
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"
	"github.com/blackducksoftware/synopsys-operator/pkg/soperator"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus" //_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

// NewController will initialize the input config file, create the hub informers, initiantiate all rest api
func NewController(configPath string, version string) (*Deployer, error) {
	config, err := GetConfig(configPath, version)
	if err != nil {
		return nil, errors.Annotate(err, "Failed to load configuration")
	}
	if config == nil {
		return nil, errors.Errorf("expected non-nil config, but got nil")
	}

	level, err := config.GetLogLevel()
	if err != nil {
		return nil, errors.Annotate(err, "unable to get log level")
	}
	log.SetLevel(level)

	log.Debugf("config: %+v", config)

	kubeConfig, err := GetKubeConfig("", false)
	if err != nil {
		return nil, errors.Annotate(err, "unable to create config for both in-cluster and external to cluster")
	}

	kubeClientSet, err := GetKubeClientSet(kubeConfig)
	if err != nil {
		return nil, errors.Annotate(err, "unable to create Kubernetes clientset")
	}

	// check for the existence of operator configmap, if not create it
	_, err = util.GetConfigMap(kubeClientSet, config.Namespace, "synopsys-operator")
	if err != nil {
		deployer, err := horizon.NewDeployer(kubeConfig)
		if err != nil {
			return nil, errors.Annotate(err, "unable to create deployer object")
		}

		config.IsOpenshift = util.IsOpenshift(kubeClientSet)
		clusterType := soperator.KubernetesClusterType
		if config.IsOpenshift {
			clusterType = soperator.OpenshiftClusterType
		}

		operatorConfig := soperator.SpecConfig{
			Namespace:                     config.Namespace,
			Expose:                        util.NONE,
			ClusterType:                   clusterType,
			DryRun:                        config.DryRun,
			LogLevel:                      config.LogLevel,
			Threadiness:                   config.Threadiness,
			PostgresRestartInMins:         config.PostgresRestartInMins,
			PodWaitTimeoutSeconds:         config.PodWaitTimeoutSeconds,
			ResyncIntervalInSeconds:       config.ResyncIntervalInSeconds,
			TerminationGracePeriodSeconds: config.TerminationGracePeriodSeconds,
			IsClusterScoped:               config.IsClusterScoped,
			Crds:                          strings.Split(config.CrdNames, ","),
		}
		operatorCm, err := operatorConfig.GetOperatorConfigMap()
		if err != nil {
			return nil, errors.Annotate(err, "unable to create operator configmap")
		}
		deployer.AddComponent(horizonapi.ConfigMapComponent, operatorCm)
		deployer.Run()
	}

	return NewDeployer(config, kubeConfig, kubeClientSet)
}
