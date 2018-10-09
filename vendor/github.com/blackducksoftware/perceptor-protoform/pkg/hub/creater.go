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
	"math"
	"os"
	"strings"
	"time"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"

	"github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
	hubclientset "github.com/blackducksoftware/perceptor-protoform/pkg/hub/client/clientset/versioned"
	"github.com/blackducksoftware/perceptor-protoform/pkg/model"
	"github.com/blackducksoftware/perceptor-protoform/pkg/util"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	securityclient "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	log "github.com/sirupsen/logrus"
)

// Creater will store the configuration to create the Hub
type Creater struct {
	Config           *model.Config
	KubeConfig       *rest.Config
	KubeClient       *kubernetes.Clientset
	HubClient        *hubclientset.Clientset
	osSecurityClient *securityclient.SecurityV1Client
	routeClient      *routeclient.RouteV1Client
}

// NewCreater will instantiate the Creater
func NewCreater(config *model.Config, kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, hubClient *hubclientset.Clientset,
	osSecurityClient *securityclient.SecurityV1Client, routeClient *routeclient.RouteV1Client) *Creater {
	return &Creater{Config: config, KubeConfig: kubeConfig, KubeClient: kubeClient, HubClient: hubClient, osSecurityClient: osSecurityClient, routeClient: routeClient}
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

	// Delete a Cluster Role Binding
	err = util.DeleteClusterRoleBinding(hc.KubeClient, namespace)
	if err != nil {
		log.Errorf("unable to delete the cluster role binding for %+v", namespace)
	}
}

// GetDefaultPasswords returns admin,user,postgres passwords for db maintainance tasks.  Should only be used during
// initialization, or for 'babysitting' ephemeral hub instances (which might have postgres restarts)
// MAKE SURE YOU SEND THE NAMESPACE OF THE SECRET SOURCE (operator), NOT OF THE new hub  THAT YOUR TRYING TO CREATE !
func GetDefaultPasswords(kubeClient *kubernetes.Clientset, nsOfSecretHolder string) (adminPassword string, userPassword string, postgresPassword string, err error) {
	blackduckSecret, err := util.GetSecret(kubeClient, nsOfSecretHolder, "blackduck-secret")
	if err != nil {
		log.Infof("warning: You need to first create a 'blackduck-secret' in this namespace with ADMIN_PASSWORD, USER_PASSWORD, POSTGRES_PASSWORD")
		return "", "", "", err
	}
	adminPassword = string(blackduckSecret.Data["ADMIN_PASSWORD"])
	userPassword = string(blackduckSecret.Data["USER_PASSWORD"])
	postgresPassword = string(blackduckSecret.Data["POSTGRES_PASSWORD"])

	// default named return
	return adminPassword, userPassword, postgresPassword, err
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

	for dbInitTry := 0; dbInitTry < math.MaxInt32; dbInitTry++ {
		// get the secret from the default operator namespace, then copy it into the hub namespace.
		adminPassword, userPassword, postgresPassword, err = GetDefaultPasswords(hc.KubeClient, hc.Config.Namespace)
		if err == nil {
			break
		} else {
			log.Infof("wasn't able to init database, sleeping 5 seconds.  try = %v", dbInitTry)
			time.Sleep(5 * time.Second)
		}
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

	// Validate all pods are in running state
	err = util.ValidatePodsAreRunningInNamespace(hc.KubeClient, createHub.Namespace)
	if err != nil {
		return "", "", true, err
	}

	if strings.EqualFold(createHub.DbPrototype, "empty") {
		err := InitDatabase(createHub, adminPassword, userPassword, postgresPassword)
		if err != nil {
			log.Errorf("%v: error: %+v", createHub.Namespace, err)
			return "", "", true, fmt.Errorf("%v: error: %+v", createHub.Namespace, err)
		}
	}

	err = hc.addAnyUIDToServiceAccount(createHub)
	if err != nil {
		log.Error(err)
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

	// Validate all pods are in running state
	err = util.ValidatePodsAreRunningInNamespace(hc.KubeClient, createHub.Namespace)
	if err != nil {
		return "", "", true, err
	}

	// Filter the registration pod to auto register the hub using the registration key from the environment variable
	registrationPod, err := util.FilterPodByNamePrefixInNamespace(hc.KubeClient, createHub.Namespace, "registration")
	log.Debugf("registration pod: %+v", registrationPod)
	if err != nil {
		return "", "", true, err
	}
	registrationKey := os.Getenv("REGISTRATION_KEY")
	// log.Debugf("registration key: %s", registrationKey)

	if registrationPod != nil && !strings.EqualFold(registrationKey, "") {
		for i := 0; i < 20; i++ {
			// Create the exec into kubernetes pod request
			req := util.CreateExecContainerRequest(hc.KubeClient, registrationPod)
			// Exec into the kubernetes pod and execute the commands
			if strings.HasPrefix(createHub.HubVersion, "4.") {
				err = hc.execContainer(req, []string{fmt.Sprintf(`curl -k -X POST "https://127.0.0.1:8443/registration/HubRegistration?registrationid=%s&action=activate"`, registrationKey)})
			} else {
				err = hc.execContainer(req, []string{fmt.Sprintf(`curl -k -X POST "https://127.0.0.1:8443/registration/HubRegistration?registrationid=%s&action=activate" -k --cert /opt/blackduck/hub/hub-registration/security/blackduck_system.crt --key /opt/blackduck/hub/hub-registration/security/blackduck_system.key`, registrationKey)})
			}

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

	// OpenShift routes
	ipAddress := ""
	if hc.routeClient != nil {
		route, err := util.CreateOpenShiftRoutes(hc.routeClient, createHub.Namespace, createHub.Namespace, "Service", "webserver")
		if err != nil {
			return "", pvcVolumeName, false, err
		}
		log.Debugf("openshift route host: %s", route.Spec.Host)
		ipAddress = route.Spec.Host
	}

	if strings.EqualFold(ipAddress, "") {
		ipAddress, err = hc.getLoadBalancerIPAddress(createHub.Namespace, "webserver-lb")
		if err != nil {
			ipAddress, err = hc.getNodePortIPAddress(createHub.Namespace, "webserver-np")
			if err != nil {
				return "", pvcVolumeName, false, err
			}
		}
	}
	log.Infof("hub Ip address: %s", ipAddress)

	go func() {
		var checks int32
		for {
			log.Debugf("%v: Waiting 3 minutes before running repair check.", createHub.Namespace)
			time.Sleep(time.Duration(3) * time.Minute) // i.e. hacky.  TODO make configurable.
			log.Debugf("%v: running postgres schema repair check # %v...", createHub.Namespace, checks)
			// name == namespace (before the namespace is set, it might be empty, but name wont be)
			hostName := fmt.Sprintf("postgres.%s.svc.cluster.local", createHub.Namespace)
			adminPassword, userPassword, postgresPassword, err := GetDefaultPasswords(hc.KubeClient, hc.Config.Namespace)

			dbNeedsInitBecause := ""

			log.Debugf("%v : Checking connection now...", createHub.Namespace)
			db, err := OpenDatabaseConnection(hostName, "bds_hub", "postgres", postgresPassword, "postgres")
			log.Debugf("%v : Done checking [ error status == %v ] ...", createHub.Namespace, err)
			if err != nil {
				dbNeedsInitBecause = "couldnt connect !"
			} else {
				_, err := db.Query("SELECT * FROM USER")
				if err != nil {
					dbNeedsInitBecause = "couldnt select!"
				}
			}
			db.Close()

			if dbNeedsInitBecause != "" {
				log.Warnf("%v: database needs init because (%v), ::: %v ", createHub.Namespace, dbNeedsInitBecause, err)
				err := InitDatabase(createHub, adminPassword, userPassword, postgresPassword)
				if err != nil {
					log.Errorf("%v: error: %+v", createHub.Namespace, err)
				}
			} else {
				log.Debugf("%v Database connection and USER table query  succeeded, not fixing ", createHub.Namespace)
			}
			checks++
		}
	}()

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
