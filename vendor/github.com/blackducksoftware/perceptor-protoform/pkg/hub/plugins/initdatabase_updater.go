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

package plugins

import (
	"fmt"
	"strings"
	"time"

	hubv1 "github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
	"github.com/blackducksoftware/perceptor-protoform/pkg/hub"
	hubclient "github.com/blackducksoftware/perceptor-protoform/pkg/hub/client/clientset/versioned"
	"github.com/blackducksoftware/perceptor-protoform/pkg/model"
	"github.com/blackducksoftware/perceptor-protoform/pkg/util"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// InitDatabaseUpdater will hold the configuration to initialize the Postgres database
type InitDatabaseUpdater struct {
	Config     *model.Config
	KubeClient *kubernetes.Clientset
	HubClient  *hubclient.Clientset
	Hubs       map[string]chan struct{}
}

// Run is a BLOCKING function which should be run by the framework .
func (i *InitDatabaseUpdater) Run(ch <-chan struct{}) {

	i.verifyHubsPostgresRestart()

	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return i.HubClient.SynopsysV1().Hubs(i.Config.Namespace).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return i.HubClient.SynopsysV1().Hubs(i.Config.Namespace).Watch(options)
		},
	}
	_, ctrl := cache.NewInformer(lw,
		&hubv1.Hub{},
		2*time.Second,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				log.Debugf("init database hub added event ! %v ", obj)
				i.addHub(obj)
			},

			DeleteFunc: func(obj interface{}) {
				log.Debugf("init database hub deleted event ! %v ", obj)
				hub := obj.(*hubv1.Hub)
				stopCh := i.Hubs[hub.Name]
				close(stopCh)
			},
		},
	)
	log.Infof("Starting controller for hub<->postgres updates... this blocks, so running in a go func.")

	// make sure this is called from a go func.
	// This blocks!
	go ctrl.Run(ch)
}

func (i *InitDatabaseUpdater) addHub(obj interface{}) {
	hub := obj.(*hubv1.Hub)
	if i.isHubThreadAlreadyExist(hub) {
		return
	}
	if strings.EqualFold(hub.Spec.BackupSupport, "No") {
		for j := 0; j < 20; j++ {
			hub, err := util.GetHub(i.HubClient, i.Config.Namespace, hub.Name)
			if err != nil {
				log.Errorf("unable to get hub %s due to %+v", hub.Name, err)
			}

			addHubSpec := hub.Spec
			if strings.EqualFold(hub.Status.State, "running") {
				i.Hubs[hub.Name] = i.startInitDatabaseUpdater(&addHubSpec)
				break
			}
			time.Sleep(10 * time.Second)
		}
	}
}

// getHubPasswords will get the hub password from the db-creds secret
func (i *InitDatabaseUpdater) getHubPasswords(hubSpec *hubv1.HubSpec) (adminPassword string, userPassword string, err error) {
	secret, err := util.GetSecret(i.KubeClient, hubSpec.Namespace, "db-creds")

	if err != nil {
		return "", "", err
	}

	adminPassword = string(secret.Data["HUB_POSTGRES_ADMIN_PASSWORD_FILE"])
	userPassword = string(secret.Data["HUB_POSTGRES_USER_PASSWORD_FILE"])
	return adminPassword, userPassword, nil
}

// startInitDatabaseUpdater will check every 3 minutes for Hub postgres restart, if so, then initialize the DB
func (i *InitDatabaseUpdater) startInitDatabaseUpdater(hubSpec *hubv1.HubSpec) chan struct{} {
	stopCh := make(chan struct{})
	go func() {
		var checks int32
		for {
			log.Debugf("%v: Waiting %d minutes before running repair check.", hubSpec.Namespace, i.Config.PostgresRestartInMins)
			select {
			case <-stopCh:
				return
			case <-time.After(time.Duration(i.Config.PostgresRestartInMins) * time.Minute):
				log.Debugf("%v: running postgres schema repair check # %v...", hubSpec.Namespace, checks)
				// name == namespace (before the namespace is set, it might be empty, but name wont be)
				hostName := fmt.Sprintf("postgres.%s.svc.cluster.local", hubSpec.Namespace)
				_, _, postgresPassword, err := hub.GetDefaultPasswords(i.KubeClient, i.Config.Namespace)
				adminPassword, userPassword, err := i.getHubPasswords(hubSpec)
				if err != nil {
					log.Errorf("password mismatch for %s because %+v", hubSpec.Namespace, err)
				}
				dbNeedsInitBecause := ""

				log.Debugf("%v : Checking connection now...", hubSpec.Namespace)
				db, err := hub.OpenDatabaseConnection(hostName, "bds_hub", "postgres", postgresPassword, "postgres")
				log.Debugf("%v : Done checking [ error status == %v ] ...", hubSpec.Namespace, err)
				if err != nil {
					dbNeedsInitBecause = "couldnt connect !"
				} else {
					_, err := db.Exec("SELECT * FROM USER;")
					if err != nil {
						dbNeedsInitBecause = "couldnt select!"
					}
				}
				db.Close()

				if dbNeedsInitBecause != "" {
					log.Warnf("%v: database needs init because (%v), ::: %v ", hubSpec.Namespace, dbNeedsInitBecause, err)
					err := hub.InitDatabase(hubSpec, adminPassword, userPassword, postgresPassword)
					if err != nil {
						log.Errorf("%v: error: %+v", hubSpec.Namespace, err)
					}
				} else {
					log.Debugf("%v Database connection and USER table query  succeeded, not fixing ", hubSpec.Namespace)
				}
				checks++
			}
		}
	}()
	return stopCh
}

// verifyHubsPostgresRestart will retrieve all Backup disabled hubs and send it to startInitDatabaseUpdater
func (i *InitDatabaseUpdater) verifyHubsPostgresRestart() {
	hubs, err := util.ListHubs(i.HubClient, i.Config.Namespace)
	if err != nil {
		log.Errorf("unable to list the hubs due to %+v", err)
	}

	for _, hub := range hubs.Items {
		verifyHub := hub
		if i.isHubThreadAlreadyExist(&verifyHub) {
			continue
		}
		if strings.EqualFold(hub.Spec.BackupSupport, "No") {
			i.Hubs[hub.Name] = i.startInitDatabaseUpdater(&verifyHub.Spec)
		}
	}
}

func (i *InitDatabaseUpdater) isHubThreadAlreadyExist(hub *hubv1.Hub) bool {
	if _, ok := i.Hubs[hub.Name]; ok {
		return true
	}
	return false
}
