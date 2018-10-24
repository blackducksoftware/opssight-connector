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
	"strconv"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"

	log "github.com/sirupsen/logrus"
)

// CreateHubConfig will create the hub configMaps
func (hc *Creater) createHubConfig(createHub *v1.HubSpec, hubContainerFlavor *ContainerFlavor) map[string]*components.ConfigMap {
	configMaps := make(map[string]*components.ConfigMap)

	hubConfig := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: createHub.Namespace, Name: "hub-config"})
	hubData := map[string]string{
		"PUBLIC_HUB_WEBSERVER_HOST": "localhost",
		"PUBLIC_HUB_WEBSERVER_PORT": "443",
		"HUB_WEBSERVER_PORT":        "8443",
		"IPV4_ONLY":                 "0",
		"RUN_SECRETS_DIR":           "/tmp/secrets",
		"HUB_VERSION":               createHub.HubVersion,
		"HUB_PROXY_NON_PROXY_HOSTS": "solr",
	}

	for _, data := range createHub.Environs {
		hubData[data.Key] = data.Value
	}
	hubConfig.AddData(hubData)

	configMaps["hub-config"] = hubConfig

	hubDbConfig := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: createHub.Namespace, Name: "hub-db-config"})
	hubDbConfig.AddData(map[string]string{
		"HUB_POSTGRES_ADMIN": "blackduck",
		"HUB_POSTGRES_USER":  "blackduck_user",
		"HUB_POSTGRES_PORT":  "5432",
		"HUB_POSTGRES_HOST":  "postgres",
	})

	configMaps["hub-db-config"] = hubDbConfig

	hubConfigResources := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: createHub.Namespace, Name: "hub-config-resources"})
	hubConfigResources.AddData(map[string]string{
		"webapp-mem":    hubContainerFlavor.WebappHubMaxMemory,
		"jobrunner-mem": hubContainerFlavor.JobRunnerHubMaxMemory,
		"scan-mem":      hubContainerFlavor.ScanHubMaxMemory,
	})

	configMaps["hub-config-resources"] = hubConfigResources

	hubDbConfigGranular := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: createHub.Namespace, Name: "hub-db-config-granular"})
	hubDbConfigGranular.AddData(map[string]string{"HUB_POSTGRES_ENABLE_SSL": "false"})

	configMaps["hub-db-config-granular"] = hubDbConfigGranular

	postgresBootstrap := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: createHub.Namespace, Name: "postgres-bootstrap"})
	var backupInSeconds int
	var err error
	switch createHub.BackupUnit {
	case "Minute(s)":
		backupInSeconds, err = strconv.Atoi(createHub.BackupInterval)
		backupInSeconds = backupInSeconds * 60
	case "Hour(s)":
		backupInSeconds, err = strconv.Atoi(createHub.BackupInterval)
		backupInSeconds = backupInSeconds * 60 * 60
	case "Week(s)":
		backupInSeconds, err = strconv.Atoi(createHub.BackupInterval)
		backupInSeconds = backupInSeconds * 60 * 60 * 24 * 7
	default:
		backupInSeconds = 24 * 60 * 60
		err = nil
	}

	if err != nil {
		log.Errorf("unable to convert %s from string to integer due to %+v and hence defaults to 24 Hours", createHub.BackupInterval, err)
		backupInSeconds = 24 * 60 * 60
	}

	postgresBootstrap.AddData(map[string]string{"pgbootstrap.sh": fmt.Sprintf(`#!/bin/bash
		BACKUP_FILENAME="%s"
		CLONE_FILENAME="%s"
		NFS_PATH="%s"
		echo "Backup file name: $NFS_PATH/$BACKUP_FILENAME"
		echo "Clone file name: $NFS_PATH/$CLONE_FILENAME"
		if [ ! -f $NFS_PATH/$BACKUP_FILENAME.sql ] && [ -f $NFS_PATH/$CLONE_FILENAME.sql ]; then
			touch /tmp/BLACKDUCK_MIGRATING
			echo "clone data file found"
			while true; do
				if psql -c "SELECT 1" &>/dev/null; then
					echo "Migrating the data from clone !"
					psql < $NFS_PATH/$CLONE_FILENAME.sql
					break
				else
					echo "unable to execute the SELECT 1, sleeping 10 seconds... before trying to init db again."
					sleep 10
				fi
			done
			rm -f /tmp/BLACKDUCK_MIGRATING
		fi

		if [ -f $NFS_PATH/$BACKUP_FILENAME.sql ]; then
			touch /tmp/BLACKDUCK_MIGRATING
			echo "backup data file found"
			while true; do
				if psql -c "SELECT 1" &>/dev/null; then
					echo "Migrating the data from backup !"
					psql < $NFS_PATH/$BACKUP_FILENAME.sql
					break
				else
					echo "unable to execute the SELECT 1, sleeping 10 seconds... before trying migration again"
					sleep 10
				fi
			done
			rm -f /tmp/BLACKDUCK_MIGRATING
		fi

		if [ "%s" == "Yes" ]; then
			while true; do
				echo "Doing periodic data dump..."
				sleep %d
				if [ ! -f $NFS_PATH/$BACKUP_FILENAME_tmp.sql ]; then
					pg_dumpall -w > $NFS_PATH/$BACKUP_FILENAME_tmp.sql
					if [ $? -eq 0 ]; then
						mv $NFS_PATH/$BACKUP_FILENAME_tmp.sql $NFS_PATH/$BACKUP_FILENAME.sql
						if [ $? -eq 0 ]; then
							rm -f $NFS_PATH/$BACKUP_FILENAME_tmp.sql
						fi
					fi
				else
					echo "backup in progress... cannot start another instance of backup"
				fi
			done
		fi`, createHub.Namespace, createHub.DbPrototype, hc.Config.NFSPath, createHub.BackupSupport, backupInSeconds)})

	configMaps["postgres-bootstrap"] = postgresBootstrap

	postgresInit := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: createHub.Namespace, Name: "postgres-init"})
	postgresInit.AddData(map[string]string{"pginit.sh": `#!/bin/bash
		echo "executing bds init script"
    sh /usr/share/container-scripts/postgresql/pgbootstrap.sh &
    run-postgresql`})

	configMaps["postgres-init"] = postgresInit

	return configMaps
}
