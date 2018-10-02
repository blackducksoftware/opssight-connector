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
	"strings"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"

	"github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
	"github.com/blackducksoftware/perceptor-protoform/pkg/util"

	log "github.com/sirupsen/logrus"
)

func (hc *Creater) init(deployer *horizon.Deployer, createHub *v1.HubSpec, hubContainerFlavor *ContainerFlavor, allConfigEnv []*horizonapi.EnvConfig, adminPassword string, userPassword string) error {

	// Create a namespaces
	_, err := util.GetNamespace(hc.KubeClient, createHub.Namespace)
	if err != nil {
		log.Debugf("unable to find the namespace %s", createHub.Namespace)
		deployer.AddNamespace(components.NewNamespace(horizonapi.NamespaceConfig{Name: createHub.Namespace}))
	}

	// Create a secret
	secrets := hc.createHubSecrets(createHub, adminPassword, userPassword)

	for _, secret := range secrets {
		deployer.AddSecret(secret)
	}

	// Create a service account
	serviceAccount := util.CreateServiceAccount(createHub.Namespace, createHub.Namespace)
	deployer.AddServiceAccount(serviceAccount)

	// Create a cluster role binding and associated it to a service account
	clusterRoleBinding := util.CreateClusterRoleBinding(createHub.Namespace, createHub.Namespace, createHub.Namespace, "", "ClusterRole", "cluster-admin")
	deployer.AddClusterRoleBinding(clusterRoleBinding)

	// Create ConfigMaps
	configMaps := hc.createHubConfig(createHub, hubContainerFlavor)

	for _, configMap := range configMaps {
		deployer.AddConfigMap(configMap)
	}

	var storageClass string
	if strings.EqualFold(createHub.PVCStorageClass, "none") {
		storageClass = ""
	} else {
		storageClass = createHub.PVCStorageClass
	}

	if strings.EqualFold(createHub.PVCStorageClass, "none") {
		// Postgres PV
		if strings.EqualFold(createHub.NFSServer, "") {
			return fmt.Errorf("unable to create the PV %s due to missing NFS server path", createHub.Namespace)
		}

		_, err = util.CreatePersistentVolume(hc.KubeClient, createHub.Namespace, storageClass, createHub.PVCClaimSize, "/data/bds/backup", createHub.NFSServer)

		if err != nil {
			return fmt.Errorf("unable to create the PV %s due to %+v", createHub.Namespace, err)
		}
	}

	postgresEnvs := allConfigEnv
	postgresEnvs = append(postgresEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "POSTGRESQL_USER", KeyOrVal: "blackduck"})
	postgresEnvs = append(postgresEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, NameOrPrefix: "POSTGRESQL_PASSWORD", KeyOrVal: "HUB_POSTGRES_USER_PASSWORD_FILE", FromName: "db-creds"})
	postgresEnvs = append(postgresEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvVal, NameOrPrefix: "POSTGRESQL_DATABASE", KeyOrVal: "blackduck"})
	postgresEnvs = append(postgresEnvs, &horizonapi.EnvConfig{Type: horizonapi.EnvFromSecret, NameOrPrefix: "POSTGRESQL_ADMIN_PASSWORD", KeyOrVal: "HUB_POSTGRES_ADMIN_PASSWORD_FILE", FromName: "db-creds"})

	postgresVolumes := []*components.Volume{}
	postgresEmptyDir, _ := util.CreateEmptyDirVolumeWithoutSizeLimit("postgres-persistent-vol")
	postgresInitConfigVol, _ := util.CreateConfigMapVolume("postgres-init-vol", "postgres-init", 0777)
	postgresBootstrapConfigVol, _ := util.CreateConfigMapVolume("postgres-bootstrap-vol", "postgres-bootstrap", 0777)
	postgresVolumes = append(postgresVolumes, postgresEmptyDir, postgresInitConfigVol, postgresBootstrapConfigVol)

	postgresVolumeMounts := []*horizonapi.VolumeMountConfig{}
	postgresVolumeMounts = append(postgresVolumeMounts, &horizonapi.VolumeMountConfig{Name: "postgres-persistent-vol", MountPath: "/var/lib/pgsql/data", Propagation: horizonapi.MountPropagationHostToContainer})
	postgresVolumeMounts = append(postgresVolumeMounts, &horizonapi.VolumeMountConfig{Name: "postgres-bootstrap-vol:pgbootstrap.sh", MountPath: "/usr/share/container-scripts/postgresql/pgbootstrap.sh", Propagation: horizonapi.MountPropagationHostToContainer})
	postgresVolumeMounts = append(postgresVolumeMounts, &horizonapi.VolumeMountConfig{Name: "postgres-init-vol:pginit.sh", MountPath: "/usr/share/container-scripts/postgresql/pginit.sh", Propagation: horizonapi.MountPropagationHostToContainer})

	if strings.EqualFold(createHub.BackupSupport, "Yes") || !strings.EqualFold(createHub.DbPrototype, "empty") {
		// Postgres PVC
		postgresPVC, err := util.CreatePersistentVolumeClaim(createHub.Namespace, createHub.Namespace, createHub.PVCClaimSize, storageClass, horizonapi.ReadWriteOnce)
		if err != nil {
			return fmt.Errorf("failed to create the postgres PVC for %s due to %+v", createHub.Namespace, err)
		}
		deployer.AddPVC(postgresPVC)

		postgresBackupDir, _ := util.CreatePersistentVolumeClaimVolume("postgres-backup-vol", createHub.Namespace)
		postgresVolumes = append(postgresVolumes, postgresBackupDir)
		postgresVolumeMounts = append(postgresVolumeMounts, &horizonapi.VolumeMountConfig{Name: "postgres-backup-vol", MountPath: "/data/bds/backup", Propagation: horizonapi.MountPropagationHostToContainer})
	}

	postgresExternalContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "postgres", Image: "registry.access.redhat.com/rhscl/postgresql-96-rhel7:1", PullPolicy: horizonapi.PullAlways,
			MinMem: hubContainerFlavor.PostgresMemoryLimit, MaxMem: "", MinCPU: hubContainerFlavor.PostgresCPULimit, MaxCPU: "",
			Command: []string{"/usr/share/container-scripts/postgresql/pginit.sh"}},
		EnvConfigs:   postgresEnvs,
		VolumeMounts: postgresVolumeMounts,
		PortConfig:   &horizonapi.PortConfig{ContainerPort: postgresPort, Protocol: horizonapi.ProtocolTCP},
	}
	initContainers := []*util.Container{}
	// If the PV storage is other than NFS or if the backup is enabled and PV storage is other than NFS, add the init container
	if !strings.EqualFold(createHub.PVCStorageClass, "") && !strings.EqualFold(createHub.PVCStorageClass, "none") {
		postgresInitContainerConfig := &util.Container{
			ContainerConfig: &horizonapi.ContainerConfig{Name: "alpine", Image: "alpine", Command: []string{"sh", "-c", "chmod -cR 777 /data/bds/backup"}},
			VolumeMounts: []*horizonapi.VolumeMountConfig{
				{Name: "postgres-backup-vol", MountPath: "/data/bds/backup", Propagation: horizonapi.MountPropagationHostToContainer},
			},
			PortConfig: &horizonapi.PortConfig{ContainerPort: "3001", Protocol: horizonapi.ProtocolTCP},
		}
		initContainers = append(initContainers, postgresInitContainerConfig)
	}

	postgres := util.CreateDeploymentFromContainer(&horizonapi.DeploymentConfig{Namespace: createHub.Namespace, Name: "postgres", Replicas: util.IntToInt32(1)}, "",
		[]*util.Container{postgresExternalContainerConfig}, postgresVolumes, initContainers, []horizonapi.AffinityConfig{})
	// log.Infof("postgres : %+v\n", postgres.GetObj())
	deployer.AddDeployment(postgres)
	deployer.AddService(util.CreateService("postgres", "postgres", createHub.Namespace, postgresPort, postgresPort, horizonapi.ClusterIPServiceTypeDefault))
	// deployer.AddService(util.CreateService("postgres-exposed", "postgres", createHub.Spec.Namespace, postgresPort, postgresPort, horizonapi.ClusterIPServiceTypeLoadBalancer))
	return nil
}
