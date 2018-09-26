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
	"os"
	"strings"
	"time"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"

	"github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
	hubclientset "github.com/blackducksoftware/perceptor-protoform/pkg/hub/client/clientset/versioned"
	"github.com/blackducksoftware/perceptor-protoform/pkg/model"
	"github.com/blackducksoftware/perceptor-protoform/pkg/util"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	log "github.com/sirupsen/logrus"
)

// Creater will store the configuration to create the Hub
type Creater struct {
	Config     *model.Config
	KubeConfig *rest.Config
	KubeClient *kubernetes.Clientset
	HubClient  *hubclientset.Clientset
}

// NewCreater will instantiate the Creater
func NewCreater(config *model.Config, kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, hubClient *hubclientset.Clientset) *Creater {
	return &Creater{Config: config, KubeConfig: kubeConfig, KubeClient: kubeClient, HubClient: hubClient}
}

// DeleteHub will delete the Black Duck Hub
func (hc *Creater) DeleteHub(namespace string) {
	var err error
	// Verify whether the namespace exist
	_, err = util.GetNamespace(hc.KubeClient, namespace)
	if err != nil {
		log.Errorf("Unable to find the namespace %+v due to %+v", namespace, err)
	} else {
		// Delete a namespace
		err = util.DeleteNamespace(hc.KubeClient, namespace)
		if err != nil {
			log.Errorf("Unable to delete the namespace %+v due to %+v", namespace, err)
		}

		for {
			// Verify whether the namespace deleted
			ns, err := util.GetNamespace(hc.KubeClient, namespace)
			log.Infof("Namespace: %v, status: %v", namespace, ns.Status)
			time.Sleep(10 * time.Second)
			if err != nil {
				log.Infof("Deleted the namespace %+v", namespace)
				break
			}
		}
	}

	// Delete a persistent volume
	err = util.DeletePersistentVolume(hc.KubeClient, namespace)
	if err != nil {
		log.Errorf("unable to delete the pv for %+v", namespace)
	}
}

// CreateHub will create the Black Duck Hub
func (hc *Creater) CreateHub(createHub *v1.HubSpec) (string, string, bool, error) {
	log.Debugf("Create Hub details for %s: %+v", createHub.Namespace, createHub)

	// Create a horizon deployer for each hub
	deployer, err := horizon.NewDeployer(hc.KubeConfig)
	if err != nil {
		return "", "", true, fmt.Errorf("unable to create the horizon deployer due to %+v", err)
	}

	// Get Containers Flavor
	hubContainerFlavor := GetContainersFlavor(createHub.Flavor)
	log.Debugf("Hub Container Flavor: %+v", hubContainerFlavor)

	if hubContainerFlavor == nil {
		return "", "", true, fmt.Errorf("invalid flavor type, Expected: Small, Medium, Large (or) OpsSight, Actual: %s", createHub.Flavor)
	}

	// All ConfigMap environment variables
	allConfigEnv := []*horizonapi.EnvConfig{
		{Type: horizonapi.EnvFromConfigMap, FromName: "hub-config"},
		{Type: horizonapi.EnvFromConfigMap, FromName: "hub-db-config"},
		{Type: horizonapi.EnvFromConfigMap, FromName: "hub-db-config-granular"},
	}

	var adminPassword, userPassword, postgresPassword string
	for {
		blackduckSecret, err := util.GetSecret(hc.KubeClient, hc.Config.Namespace, "blackduck-secret")
		if err != nil {
			log.Infof("Aborting: You need to first create a 'blackduck-secret' in this namespace with ADMIN_PASSWORD,USER_PASSWORD,POSTGRES_PASSWORD and retry")
		} else {
			adminPassword = string(blackduckSecret.Data["ADMIN_PASSWORD"])
			userPassword = string(blackduckSecret.Data["USER_PASSWORD"])
			postgresPassword = string(blackduckSecret.Data["POSTGRES_PASSWORD"])
			break
		}
		time.Sleep(5 * time.Second)
	}

	log.Debugf("Before init: %+v", &createHub)
	// Create the config-maps, secrets and postgres container
	err = hc.init(deployer, createHub, hubContainerFlavor, allConfigEnv, adminPassword, userPassword)
	if err != nil {
		return "", "", true, err
	}
	// Deploy config-maps, secrets and postgres container
	err = deployer.Run()
	if err != nil {
		log.Errorf("deployments failed because %+v", err)
	}
	// time.Sleep(20 * time.Second)
	// Get all pods corresponding to the hub namespace
	pods, err := util.GetAllPodsForNamespace(hc.KubeClient, createHub.Namespace)
	if err != nil {
		return "", "", true, fmt.Errorf("unable to list the pods in namespace %s due to %+v", createHub.Namespace, err)
	}
	// Validate all pods are in running state
	util.ValidatePodsAreRunning(hc.KubeClient, pods)
	// Initialize the hub database
	if strings.EqualFold(createHub.DbPrototype, "empty") {
		InitDatabase(createHub, adminPassword, userPassword, postgresPassword)
	}

	// Create all hub deployments
	deployer, _ = horizon.NewDeployer(hc.KubeConfig)
	hc.createDeployer(deployer, createHub, hubContainerFlavor, allConfigEnv)
	log.Debugf("%+v", deployer)
	// Deploy all hub containers
	err = deployer.Run()
	if err != nil {
		log.Errorf("deployments failed because %+v", err)
		return "", "", true, fmt.Errorf("unable to deploy the hub in %s due to %+v", createHub.Namespace, err)
	}
	time.Sleep(10 * time.Second)
	// Get all pods corresponding to the hub namespace
	pods, err = util.GetAllPodsForNamespace(hc.KubeClient, createHub.Namespace)
	if err != nil {
		return "", "", true, fmt.Errorf("unable to list the pods in namespace %s due to %+v", createHub.Namespace, err)
	}
	// Validate all pods are in running state
	util.ValidatePodsAreRunning(hc.KubeClient, pods)

	// Filter the registration pod to auto register the hub using the registration key from the environment variable
	registrationPod := util.FilterPodByNamePrefix(pods, "registration")
	log.Debugf("registration pod: %+v", registrationPod)
	registrationKey := os.Getenv("REGISTRATION_KEY")
	// log.Debugf("registration key: %s", registrationKey)

	if registrationPod != nil && !strings.EqualFold(registrationKey, "") {
		for {
			// Create the exec into kubernetes pod request
			req := util.CreateExecContainerRequest(hc.KubeClient, registrationPod)
			// Exec into the kubernetes pod and execute the commands
			err = hc.execContainer(req, []string{fmt.Sprintf("curl -k -X POST https://127.0.0.1:8443/registration/HubRegistration?action=activate\\&registrationid=%s", registrationKey)})
			if err != nil {
				log.Infof("error in Stream: %v", err)
			} else {
				// Hub created and auto registered. Exit!!!!
				break
			}
			time.Sleep(10 * time.Second)
		}
	}

	// Retrieve the PVC volume name
	pvcVolumeName := ""
	if strings.EqualFold(createHub.BackupSupport, "Yes") || !strings.EqualFold(createHub.DbPrototype, "empty") {
		pvcVolumeName, err = hc.getPVCVolumeName(createHub.Namespace)
		if err != nil {
			return "", "", false, err
		}
	}

	ipAddress, err := hc.getLoadBalancerIPAddress(createHub.Namespace, "webserver-lb")
	if err != nil {
		ipAddress, err = hc.getNodePortIPAddress(createHub.Namespace, "webserver-np")
		if err != nil {
			return "", pvcVolumeName, false, err
		}
	}
	log.Infof("hub Ip address: %s", ipAddress)
	return ipAddress, pvcVolumeName, false, nil
}

func (hc *Creater) getPVCVolumeName(namespace string) (string, error) {
	for i := 0; i < 60; i++ {
		time.Sleep(10 * time.Second)
		pvc, err := util.GetPVC(hc.KubeClient, namespace, namespace)
		if err != nil {
			return "", fmt.Errorf("unable to get pvc in %s namespace due to %s", namespace, err.Error())
		}

		log.Debugf("pvc: %v", pvc)

		if strings.EqualFold(pvc.Spec.VolumeName, "") {
			continue
		} else {
			return pvc.Spec.VolumeName, nil
		}
	}
	return "", fmt.Errorf("timeout: unable to get pvc %s in %s namespace", namespace, namespace)
}

func (hc *Creater) getLoadBalancerIPAddress(namespace string, serviceName string) (string, error) {
	for i := 0; i < 10; i++ {
		time.Sleep(10 * time.Second)
		service, err := util.GetService(hc.KubeClient, namespace, serviceName)
		if err != nil {
			return "", fmt.Errorf("unable to get service %s in %s namespace due to %s", serviceName, namespace, err.Error())
		}

		log.Debugf("Service: %v", service)

		if len(service.Status.LoadBalancer.Ingress) > 0 {
			ipAddress := service.Status.LoadBalancer.Ingress[0].IP
			return ipAddress, nil
		}
	}
	return "", fmt.Errorf("timeout: unable to get ip address for the service %s in %s namespace", serviceName, namespace)
}

func (hc *Creater) getNodePortIPAddress(namespace string, serviceName string) (string, error) {
	for i := 0; i < 10; i++ {
		time.Sleep(10 * time.Second)
		service, err := util.GetService(hc.KubeClient, namespace, serviceName)
		if err != nil {
			return "", fmt.Errorf("unable to get service %s in %s namespace due to %s", serviceName, namespace, err.Error())
		}

		log.Debugf("Service: %v", service)

		if !strings.EqualFold(service.Spec.ClusterIP, "") {
			ipAddress := service.Spec.ClusterIP
			return ipAddress, nil
		}
	}
	return "", fmt.Errorf("timeout: unable to get ip address for the service %s in %s namespace", serviceName, namespace)
}
