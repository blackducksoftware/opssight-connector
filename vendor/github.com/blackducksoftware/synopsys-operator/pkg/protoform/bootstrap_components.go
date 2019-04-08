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
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
)

// GetBootstrapComponents will return the authentication deployment
func GetBootstrapComponents(ns string, branch string, regKey string) (*components.ReplicationController, *components.Service, *components.ConfigMap, *components.ServiceAccount, *components.ClusterRoleBinding, *components.Service, *components.ReplicationController, *components.ConfigMap) {
	protoformVolume, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("synopsys-operator")
	protoformContainer := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{
			Name:       "synopsys-operator",
			Image:      fmt.Sprintf("gcr.io/saas-hub-stg/blackducksoftware/synopsys-operator:%v", regKey),
			PullPolicy: horizonapi.PullAlways,
		},
		EnvConfigs: []*horizonapi.EnvConfig{{Type: horizonapi.EnvFromConfigMap, FromName: "registration-key"}},
		VolumeMounts: []*horizonapi.VolumeMountConfig{
			{Name: "synopsys-operator", MountPath: "/etc/synopsys-operator"},
		},
		PortConfig: []*horizonapi.PortConfig{{ContainerPort: "8080", Protocol: horizonapi.ProtocolTCP}},
	}

	protoformRC := util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{
			Namespace: ns,
			Name:      "synopsys-operator",
			Replicas:  util.IntToInt32(1),
		},
		"default",
		[]*util.Container{protoformContainer},
		[]*components.Volume{protoformVolume, protoformVolume},
		[]*util.Container{},
		[]horizonapi.AffinityConfig{}, map[string]string{"app": "synopsys-operator"}, map[string]string{"app": "synopsys-operator"})

	protoformsvc := util.CreateService("synopsys-operator", map[string]string{"app": "synopsys-operator"}, ns, "8080", "8080", horizonapi.ClusterIPServiceTypeDefault, map[string]string{"app": "synopsys-operator"})

	// Config map

	protoformcfg := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: ns, Name: "synopsys-operator"})
	protoformcfg.AddData(map[string]string{
		"config.json": fmt.Sprintf(`"DryRun": false,
		"LogLevel":              "debug",
		"Namespace":             "%v",
		"Threadiness":           5,
		"PostgresRestartInMins": 10,
		"NFSPath":               "/kubenfs",
		"HubFederatorConfig": {
			"HubConfig": {
				"User":                      "sysadmin",
				"PasswordEnvVar":            "HUB_PASSWORD",
				"ClientTimeoutMilliseconds": 5000,
				"Port": 443,
				"FetchAllProjectsPauseSeconds": 60},
			"UseMockMode":  false,
			"Port":         3016,
			"Registry":     "gcr.io",
			"ImagePath":    "saas-hub-stg/blackducksoftware",
			"ImageName":    "federator",
			"ImageVersion": "master",
			},
		}`, ns)})

	// RBAC
	svcAcct := components.NewServiceAccount(horizonapi.ServiceAccountConfig{
		Name:      "synopsys-operator",
		Namespace: ns})
	clusterRoleBinding := components.NewClusterRoleBinding(horizonapi.ClusterRoleBindingConfig{
		Name: "protoform-admin",
	})
	clusterRoleBinding.AddSubject(horizonapi.SubjectConfig{
		Kind:      "ServiceAccount",
		Name:      "synopsys-operator",
		Namespace: ns,
	})
	clusterRoleBinding.AddRoleRef(horizonapi.RoleRefConfig{
		Kind:     "ClusterRole",
		Name:     "cluster-admin",
		APIGroup: "",
	})

	// prometheus
	prometheusService := util.CreateService("prometheus", map[string]string{"app": "prometheus"}, ns, "8080", "8080", horizonapi.ClusterIPServiceTypeDefault, map[string]string{"app": "prometheus"})

	prometheusContainer := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{
			Name:  "prometheus",
			Image: fmt.Sprintf("prom/prometheus:v2.1.0"),
		},
		VolumeMounts: []*horizonapi.VolumeMountConfig{
			{
				Name:      "data",
				MountPath: "/data",
			},
			{
				Name:      "config-volume",
				MountPath: "/etc/prometheus",
			},
		},
		PortConfig: []*horizonapi.PortConfig{{ContainerPort: "8080", Protocol: horizonapi.ProtocolTCP}},
	}
	prometheusVol1, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("data")
	prometheusVol2, _ := util.CreateConfigMapVolume("config-volume", "prometheus-config", 777)
	promCfg := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: ns, Name: "prometheus-config"})
	protoformcfg.AddData(map[string]string{
		"prometheus.yml": `{"global":{"scrape_interval":"5s"},"scrape_configs":[{"job_name":"synopsys-operator-scrape","scrape_interval":"5s","static_configs":[{"targets":["synopsys-operator:8080"]}]}]}`})

	prometheusRC := util.CreateReplicationControllerFromContainer(
		&horizonapi.ReplicationControllerConfig{
			Namespace: ns,
			Name:      "prometheus",
			Replicas:  util.IntToInt32(1),
		},
		"default",
		[]*util.Container{prometheusContainer},
		[]*components.Volume{prometheusVol1, prometheusVol2},
		[]*util.Container{},
		[]horizonapi.AffinityConfig{}, map[string]string{"app": "prometheus"}, map[string]string{"app": "prometheus"})

	return protoformRC, protoformsvc, protoformcfg, svcAcct, clusterRoleBinding, prometheusService, prometheusRC, promCfg
}
