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

package hub

import (
	"fmt"
	"time"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"
	"github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
	"github.com/blackducksoftware/perceptor-protoform/pkg/util"
	log "github.com/sirupsen/logrus"
)

// createDeployer will create an entire hub for you.  TODO add flavor parameters !
// To create the returned hub, run 	CreateHub().Run().
func (hc *Creater) createDeployer(deployer *horizon.Deployer, createHub *v1.HubSpec, hubContainerFlavor *ContainerFlavor, allConfigEnv []*horizonapi.EnvConfig) {

	// Hub ConfigMap environment variables
	hubConfigEnv := []*horizonapi.EnvConfig{
		{Type: horizonapi.EnvFromConfigMap, FromName: "hub-config"},
	}

	dbSecretVolume := components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "db-passwords",
		MapOrSecretName: "db-creds",
		Items: map[string]horizonapi.KeyAndMode{
			"HUB_POSTGRES_ADMIN_PASSWORD_FILE": {KeyOrPath: "HUB_POSTGRES_ADMIN_PASSWORD_FILE", Mode: util.IntToInt32(420)},
			"HUB_POSTGRES_USER_PASSWORD_FILE":  {KeyOrPath: "HUB_POSTGRES_USER_PASSWORD_FILE", Mode: util.IntToInt32(420)},
		},
		DefaultMode: util.IntToInt32(420),
	})

	dbEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("cloudsql")

	// cfssl
	// cfsslGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-cfssl", fmt.Sprintf("%s-%s", "cfssl-disk", createHub.Namespace), "ext4")
	cfsslEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-cfssl")
	cfsslContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "cfssl", Image: fmt.Sprintf("%s/%s/hub-cfssl:%s", createHub.DockerRegistry, createHub.DockerRepo, createHub.HubVersion),
			PullPolicy: horizonapi.PullAlways, MinMem: hubContainerFlavor.CfsslMemoryLimit, MaxMem: hubContainerFlavor.CfsslMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   hubConfigEnv,
		VolumeMounts: []*horizonapi.VolumeMountConfig{{Name: "dir-cfssl", MountPath: "/etc/cfssl", Propagation: horizonapi.MountPropagationNone}},
		PortConfig:   &horizonapi.PortConfig{ContainerPort: cfsslPort, Protocol: horizonapi.ProtocolTCP},
	}
	cfssl := util.CreateDeploymentFromContainer(&horizonapi.DeploymentConfig{Namespace: createHub.Namespace, Name: "cfssl", Replicas: util.IntToInt32(1)},
		[]*util.Container{cfsslContainerConfig}, []*components.Volume{cfsslEmptyDir}, []*util.Container{},
		[]horizonapi.AffinityConfig{})
	// log.Infof("cfssl : %v\n", cfssl.GetObj())
	deployer.AddDeployment(cfssl)

	// webserver
	// webServerGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-webserver", fmt.Sprintf("%s-%s", "webserver-disk", createHub.Namespace), "ext4")
	for {
		secret, err := util.GetSecret(hc.KubeClient, createHub.Namespace, "blackduck-certificate")
		if err != nil {
			log.Errorf("unable to get the secret in %s due to %+v", createHub.Namespace, err)
			break
		}
		data := secret.Data
		if len(data) > 0 {
			break
		}
		time.Sleep(10 * time.Second)
	}
	webServerEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-webserver")
	webServerSecretVol, _ := util.CreateSecretVolume("certificate", "blackduck-certificate", 0777)
	webServerContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "webserver", Image: fmt.Sprintf("%s/%s/hub-nginx:%s", createHub.DockerRegistry, createHub.DockerRepo, createHub.HubVersion),
			PullPolicy: horizonapi.PullAlways, MinMem: hubContainerFlavor.WebserverMemoryLimit, MaxMem: hubContainerFlavor.WebserverMemoryLimit, MinCPU: "", MaxCPU: "", UID: util.IntToInt64(1000)},
		EnvConfigs: hubConfigEnv,
		VolumeMounts: []*horizonapi.VolumeMountConfig{
			{Name: "dir-webserver", MountPath: "/opt/blackduck/hub/webserver/security", Propagation: horizonapi.MountPropagationNone},
			{Name: "certificate", MountPath: "/tmp/secrets", Propagation: horizonapi.MountPropagationNone},
		},
		PortConfig: &horizonapi.PortConfig{ContainerPort: webserverPort, Protocol: horizonapi.ProtocolTCP},
	}
	webserver := util.CreateDeploymentFromContainer(&horizonapi.DeploymentConfig{Namespace: createHub.Namespace, Name: "webserver", Replicas: util.IntToInt32(1)},
		[]*util.Container{webServerContainerConfig}, []*components.Volume{webServerEmptyDir, webServerSecretVol}, []*util.Container{},
		[]horizonapi.AffinityConfig{})
	// log.Infof("webserver : %v\n", webserver.GetObj())
	deployer.AddDeployment(webserver)

	// documentation
	documentationContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "documentation", Image: fmt.Sprintf("%s/%s/hub-documentation:%s", createHub.DockerRegistry, createHub.DockerRepo, createHub.HubVersion),
			PullPolicy: horizonapi.PullAlways, MinMem: hubContainerFlavor.DocumentationMemoryLimit, MaxMem: hubContainerFlavor.DocumentationMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs: hubConfigEnv,
		PortConfig: &horizonapi.PortConfig{ContainerPort: documentationPort, Protocol: horizonapi.ProtocolTCP},
	}
	documentation := util.CreateDeploymentFromContainer(&horizonapi.DeploymentConfig{Namespace: createHub.Namespace, Name: "documentation", Replicas: util.IntToInt32(1)},
		[]*util.Container{documentationContainerConfig}, []*components.Volume{}, []*util.Container{}, []horizonapi.AffinityConfig{})
	// log.Infof("documentation : %v\n", documentation.GetObj())
	deployer.AddDeployment(documentation)

	// solr
	// solrGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-solr", fmt.Sprintf("%s-%s", "solr-disk", createHub.Namespace), "ext4")
	solrEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-solr")
	solrContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "solr", Image: fmt.Sprintf("%s/%s/hub-solr:%s", createHub.DockerRegistry, createHub.DockerRepo, createHub.HubVersion),
			PullPolicy: horizonapi.PullAlways, MinMem: hubContainerFlavor.SolrMemoryLimit, MaxMem: hubContainerFlavor.SolrMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   hubConfigEnv,
		VolumeMounts: []*horizonapi.VolumeMountConfig{{Name: "dir-solr", MountPath: "/opt/blackduck/hub/solr/cores.data", Propagation: horizonapi.MountPropagationNone}},
		PortConfig:   &horizonapi.PortConfig{ContainerPort: solrPort, Protocol: horizonapi.ProtocolTCP},
	}
	solr := util.CreateDeploymentFromContainer(&horizonapi.DeploymentConfig{Namespace: createHub.Namespace, Name: "solr", Replicas: util.IntToInt32(1)},
		[]*util.Container{solrContainerConfig}, []*components.Volume{solrEmptyDir}, []*util.Container{},
		[]horizonapi.AffinityConfig{})
	// log.Infof("solr : %v\n", solr.GetObj())
	deployer.AddDeployment(solr)

	// registration
	// registrationGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-registration", fmt.Sprintf("%s-%s", "registration-disk", createHub.Namespace), "ext4")
	registrationEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-registration")
	registrationContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "registration", Image: fmt.Sprintf("%s/%s/hub-registration:%s", createHub.DockerRegistry, createHub.DockerRepo, createHub.HubVersion),
			PullPolicy: horizonapi.PullAlways, MinMem: hubContainerFlavor.RegistrationMemoryLimit, MaxMem: hubContainerFlavor.RegistrationMemoryLimit, MinCPU: "1", MaxCPU: ""},
		EnvConfigs:   hubConfigEnv,
		VolumeMounts: []*horizonapi.VolumeMountConfig{{Name: "dir-registration", MountPath: "/opt/blackduck/hub/hub-registration/config", Propagation: horizonapi.MountPropagationNone}},
		PortConfig:   &horizonapi.PortConfig{ContainerPort: registrationPort, Protocol: horizonapi.ProtocolTCP},
	}
	registration := util.CreateDeploymentFromContainer(&horizonapi.DeploymentConfig{Namespace: createHub.Namespace, Name: "registration", Replicas: util.IntToInt32(1)},
		[]*util.Container{registrationContainerConfig}, []*components.Volume{registrationEmptyDir}, []*util.Container{},
		[]horizonapi.AffinityConfig{})
	// log.Infof("registration : %v\n", registration.GetObj())
	deployer.AddDeployment(registration)

	// zookeeper
	zookeeperEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-zookeeper")
	zookeeperContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "zookeeper", Image: fmt.Sprintf("%s/%s/hub-zookeeper:%s", createHub.DockerRegistry, createHub.DockerRepo, createHub.HubVersion),
			PullPolicy: horizonapi.PullAlways, MinMem: hubContainerFlavor.ZookeeperMemoryLimit, MaxMem: hubContainerFlavor.ZookeeperMemoryLimit, MinCPU: "1", MaxCPU: ""},
		EnvConfigs:   hubConfigEnv,
		VolumeMounts: []*horizonapi.VolumeMountConfig{{Name: "dir-zookeeper", MountPath: "/opt/blackduck/hub/logs", Propagation: horizonapi.MountPropagationNone}},
		PortConfig:   &horizonapi.PortConfig{ContainerPort: zookeeperPort, Protocol: horizonapi.ProtocolTCP},
	}
	zookeeper := util.CreateDeploymentFromContainer(&horizonapi.DeploymentConfig{Namespace: createHub.Namespace, Name: "zookeeper", Replicas: util.IntToInt32(1)},
		[]*util.Container{zookeeperContainerConfig}, []*components.Volume{zookeeperEmptyDir}, []*util.Container{}, []horizonapi.AffinityConfig{})
	// log.Infof("zookeeper : %v\n", zookeeper.GetObj())
	deployer.AddDeployment(zookeeper)

	// jobRunner
	jobRunnerEnvs := allConfigEnv
	jobRunnerEnvs = append(jobRunnerEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromConfigMap, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: "jobrunner-mem", FromName: "hub-config-resources"})
	jobRunnerContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "jobrunner", Image: fmt.Sprintf("%s/%s/hub-jobrunner:%s", createHub.DockerRegistry, createHub.DockerRepo, createHub.HubVersion),
			PullPolicy: horizonapi.PullAlways, MinMem: hubContainerFlavor.JobRunnerMemoryLimit, MaxMem: hubContainerFlavor.JobRunnerMemoryLimit, MinCPU: "1", MaxCPU: "1"},
		EnvConfigs:   jobRunnerEnvs,
		VolumeMounts: []*horizonapi.VolumeMountConfig{{Name: "db-passwords", MountPath: "/tmp/secrets", Propagation: horizonapi.MountPropagationNone}},
		PortConfig:   &horizonapi.PortConfig{ContainerPort: jobRunnerPort, Protocol: horizonapi.ProtocolTCP},
	}

	jobRunner := util.CreateDeploymentFromContainer(&horizonapi.DeploymentConfig{Namespace: createHub.Namespace, Name: "jobrunner", Replicas: hubContainerFlavor.JobRunnerReplicas},
		[]*util.Container{jobRunnerContainerConfig}, []*components.Volume{dbSecretVolume, dbEmptyDir}, []*util.Container{},
		[]horizonapi.AffinityConfig{})
	// log.Infof("jobRunner : %v\n", jobRunner.GetObj())
	deployer.AddDeployment(jobRunner)

	// hub-scan
	scannerEnvs := allConfigEnv
	scannerEnvs = append(scannerEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromConfigMap, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: "scan-mem", FromName: "hub-config-resources"})
	hubScanEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-scan")
	hubScanContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "hub-scan", Image: fmt.Sprintf("%s/%s/hub-scan:%s", createHub.DockerRegistry, createHub.DockerRepo, createHub.HubVersion),
			PullPolicy: horizonapi.PullAlways, MinMem: hubContainerFlavor.ScanMemoryLimit, MaxMem: hubContainerFlavor.ScanMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs: scannerEnvs,
		VolumeMounts: []*horizonapi.VolumeMountConfig{
			{Name: "db-passwords", MountPath: "/tmp/secrets", Propagation: horizonapi.MountPropagationNone},
			{Name: "dir-scan", MountPath: "/opt/blackduck/hub/hub-scan/security", Propagation: horizonapi.MountPropagationNone}},
		PortConfig: &horizonapi.PortConfig{ContainerPort: scannerPort, Protocol: horizonapi.ProtocolTCP},
	}
	hubScan := util.CreateDeploymentFromContainer(&horizonapi.DeploymentConfig{Namespace: createHub.Namespace, Name: "hub-scan", Replicas: hubContainerFlavor.ScanReplicas},
		[]*util.Container{hubScanContainerConfig}, []*components.Volume{hubScanEmptyDir, dbSecretVolume, dbEmptyDir}, []*util.Container{}, []horizonapi.AffinityConfig{})
	// log.Infof("hubScan : %v\n", hubScan.GetObj())
	deployer.AddDeployment(hubScan)

	// hub-authentication
	authEnvs := allConfigEnv
	authEnvs = append(authEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: hubContainerFlavor.AuthenticationHubMaxMemory})
	// hubAuthGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-authentication", fmt.Sprintf("%s-%s", "authentication-disk", createHub.Namespace), "ext4")
	hubAuthEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-authentication")
	hubAuthContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "hub-authentication", Image: fmt.Sprintf("%s/%s/hub-authentication:%s", createHub.DockerRegistry, createHub.DockerRepo, createHub.HubVersion),
			PullPolicy: horizonapi.PullAlways, MinMem: hubContainerFlavor.AuthenticationMemoryLimit, MaxMem: hubContainerFlavor.AuthenticationMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs: authEnvs,
		VolumeMounts: []*horizonapi.VolumeMountConfig{
			{Name: "db-passwords", MountPath: "/tmp/secrets", Propagation: horizonapi.MountPropagationNone},
			{Name: "dir-authentication", MountPath: "/opt/blackduck/hub/hub-authentication/security", Propagation: horizonapi.MountPropagationNone}},
		PortConfig: &horizonapi.PortConfig{ContainerPort: authenticationPort, Protocol: horizonapi.ProtocolTCP},
	}
	hubAuth := util.CreateDeploymentFromContainer(&horizonapi.DeploymentConfig{Namespace: createHub.Namespace, Name: "hub-authentication", Replicas: util.IntToInt32(1)},
		[]*util.Container{hubAuthContainerConfig}, []*components.Volume{hubAuthEmptyDir, dbSecretVolume, dbEmptyDir}, []*util.Container{},
		[]horizonapi.AffinityConfig{})
	// log.Infof("hubAuth : %v\n", hubAuthc.GetObj())
	deployer.AddDeployment(hubAuth)

	// webapp-logstash
	webappEnvs := allConfigEnv
	webappEnvs = append(webappEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromConfigMap, NameOrPrefix: "HUB_MAX_MEMORY", KeyOrVal: "webapp-mem", FromName: "hub-config-resources"})
	// webappGCEPersistentDiskVol := CreateGCEPersistentDiskVolume("dir-webapp", fmt.Sprintf("%s-%s", "webapp-disk", createHub.Namespace), "ext4")
	webappEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-webapp")
	webappContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "webapp", Image: fmt.Sprintf("%s/%s/hub-webapp:%s", createHub.DockerRegistry, createHub.DockerRepo, createHub.HubVersion),
			PullPolicy: horizonapi.PullAlways, MinMem: hubContainerFlavor.WebappMemoryLimit, MaxMem: hubContainerFlavor.WebappMemoryLimit, MinCPU: hubContainerFlavor.WebappCPULimit,
			MaxCPU: hubContainerFlavor.WebappCPULimit},
		EnvConfigs: webappEnvs,
		VolumeMounts: []*horizonapi.VolumeMountConfig{
			{Name: "db-passwords", MountPath: "/tmp/secrets", Propagation: horizonapi.MountPropagationNone},
			{Name: "dir-webapp", MountPath: "/opt/blackduck/hub/hub-webapp/security", Propagation: horizonapi.MountPropagationNone},
			{Name: "dir-logstash", MountPath: "/opt/blackduck/hub/logs", Propagation: horizonapi.MountPropagationNone}},
		PortConfig: &horizonapi.PortConfig{ContainerPort: webappPort, Protocol: horizonapi.ProtocolTCP},
	}
	logstashEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("dir-logstash")
	logstashContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "logstash", Image: fmt.Sprintf("%s/%s/hub-logstash:%s", createHub.DockerRegistry, createHub.DockerRepo, createHub.HubVersion),
			PullPolicy: horizonapi.PullAlways, MinMem: hubContainerFlavor.LogstashMemoryLimit, MaxMem: hubContainerFlavor.LogstashMemoryLimit, MinCPU: "", MaxCPU: ""},
		EnvConfigs:   hubConfigEnv,
		VolumeMounts: []*horizonapi.VolumeMountConfig{{Name: "dir-logstash", MountPath: "/var/lib/logstash/data", Propagation: horizonapi.MountPropagationNone}},
		PortConfig:   &horizonapi.PortConfig{ContainerPort: logstashPort, Protocol: horizonapi.ProtocolTCP},
	}
	webappLogstash := util.CreateDeploymentFromContainer(&horizonapi.DeploymentConfig{Namespace: createHub.Namespace, Name: "webapp-logstash", Replicas: util.IntToInt32(1)}, []*util.Container{webappContainerConfig, logstashContainerConfig},
		[]*components.Volume{webappEmptyDir, logstashEmptyDir, dbSecretVolume, dbEmptyDir}, []*util.Container{},
		[]horizonapi.AffinityConfig{})
	// log.Infof("webappLogstash : %v\n", webappLogstashc.GetObj())
	deployer.AddDeployment(webappLogstash)

	deployer.AddService(util.CreateService("cfssl", "cfssl", createHub.Namespace, cfsslPort, cfsslPort, horizonapi.ClusterIPServiceTypeDefault))
	deployer.AddService(util.CreateService("zookeeper", "zookeeper", createHub.Namespace, zookeeperPort, zookeeperPort, horizonapi.ClusterIPServiceTypeDefault))
	deployer.AddService(util.CreateService("webserver", "webserver", createHub.Namespace, "443", webserverPort, horizonapi.ClusterIPServiceTypeDefault))
	deployer.AddService(util.CreateService("webserver-np", "webserver", createHub.Namespace, "443", webserverPort, horizonapi.ClusterIPServiceTypeNodePort))
	deployer.AddService(util.CreateService("webserver-lb", "webserver", createHub.Namespace, "443", webserverPort, horizonapi.ClusterIPServiceTypeLoadBalancer))
	deployer.AddService(util.CreateService("webapp", "webapp-logstash", createHub.Namespace, webappPort, webappPort, horizonapi.ClusterIPServiceTypeDefault))
	deployer.AddService(util.CreateService("logstash", "webapp-logstash", createHub.Namespace, logstashPort, logstashPort, horizonapi.ClusterIPServiceTypeDefault))
	deployer.AddService(util.CreateService("solr", "solr", createHub.Namespace, solrPort, solrPort, horizonapi.ClusterIPServiceTypeDefault))
	deployer.AddService(util.CreateService("documentation", "documentation", createHub.Namespace, documentationPort, documentationPort, horizonapi.ClusterIPServiceTypeDefault))
	deployer.AddService(util.CreateService("scan", "hub-scan", createHub.Namespace, scannerPort, scannerPort, horizonapi.ClusterIPServiceTypeDefault))
	deployer.AddService(util.CreateService("authentication", "hub-authentication", createHub.Namespace, authenticationPort, authenticationPort, horizonapi.ClusterIPServiceTypeDefault))
	deployer.AddService(util.CreateService("registration", "registration", createHub.Namespace, registrationPort, registrationPort, horizonapi.ClusterIPServiceTypeDefault))
}
