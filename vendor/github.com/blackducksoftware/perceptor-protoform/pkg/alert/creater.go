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

package alert

import (
	"time"

	alertclientset "github.com/blackducksoftware/perceptor-protoform/pkg/alert/client/clientset/versioned"
	"github.com/blackducksoftware/perceptor-protoform/pkg/api/alert/v1"
	"github.com/blackducksoftware/perceptor-protoform/pkg/util"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	log "github.com/sirupsen/logrus"
)

// Creater will store the configuration to create the Hub
type Creater struct {
	kubeConfig  *rest.Config
	kubeClient  *kubernetes.Clientset
	alertClient *alertclientset.Clientset
}

// NewCreater will instantiate the Creater
func NewCreater(kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, alertClient *alertclientset.Clientset) *Creater {
	return &Creater{kubeConfig: kubeConfig, kubeClient: kubeClient, alertClient: alertClient}
}

// DeleteAlert will delete the Black Duck Alert
func (ac *Creater) DeleteAlert(namespace string) {
	log.Debugf("Delete Alert details for %s", namespace)
	var err error
	// Verify whether the namespace exist
	_, err = util.GetNamespace(ac.kubeClient, namespace)
	if err != nil {
		log.Errorf("Unable to find the namespace %+v due to %+v", namespace, err)
	} else {
		// Delete a namespace
		err = util.DeleteNamespace(ac.kubeClient, namespace)
		if err != nil {
			log.Errorf("Unable to delete the namespace %+v due to %+v", namespace, err)
		}

		for {
			// Verify whether the namespace deleted
			ns, err := util.GetNamespace(ac.kubeClient, namespace)
			log.Infof("Namespace: %v, status: %v", namespace, ns.Status)
			time.Sleep(10 * time.Second)
			if err != nil {
				log.Infof("Deleted the namespace %+v", namespace)
				break
			}
		}
	}
}

// CreateAlert will create the Black Duck Alert
func (ac *Creater) CreateAlert(createAlert *v1.AlertSpec) error {
	log.Debugf("Create Alert details for %s: %+v", createAlert.Namespace, createAlert)
	alert := NewAlert(createAlert)
	components, err := alert.GetComponents()
	if err != nil {
		log.Errorf("unable to get alert components for %s due to %+v", createAlert.Namespace, err)
		return err
	}
	deployer, err := util.NewDeployer(ac.kubeConfig)
	if err != nil {
		log.Errorf("unable to get deployer object for %s due to %+v", createAlert.Namespace, err)
		return err
	}
	deployer.PreDeploy(components, createAlert.Namespace)
	err = deployer.Run()
	if err != nil {
		log.Errorf("unable to deploy alert app due to %+v", err)
	}
	deployer.StartControllers()
	return nil
}
