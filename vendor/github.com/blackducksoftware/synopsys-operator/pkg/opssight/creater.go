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

package opssight

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	hub_v2 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	hubclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	securityclient "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Creater will store the configuration to create OpsSight
type Creater struct {
	config           *protoform.Config
	kubeConfig       *rest.Config
	kubeClient       *kubernetes.Clientset
	opssightClient   *opssightclientset.Clientset
	osSecurityClient *securityclient.SecurityV1Client
	routeClient      *routeclient.RouteV1Client
	hubClient        *hubclientset.Clientset
}

// NewCreater will instantiate the Creater
func NewCreater(config *protoform.Config, kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, opssightClient *opssightclientset.Clientset, osSecurityClient *securityclient.SecurityV1Client, routeClient *routeclient.RouteV1Client, hubClient *hubclientset.Clientset) *Creater {
	return &Creater{
		config:           config,
		kubeConfig:       kubeConfig,
		kubeClient:       kubeClient,
		opssightClient:   opssightClient,
		osSecurityClient: osSecurityClient,
		routeClient:      routeClient,
		hubClient:        hubClient,
	}
}

// DeleteOpsSight will delete the Black Duck OpsSight
func (ac *Creater) DeleteOpsSight(namespace string) error {
	log.Debugf("delete OpsSight details for %s", namespace)
	// Verify that the namespace exists
	_, err := util.GetNamespace(ac.kubeClient, namespace)
	if err != nil {
		return errors.Annotatef(err, "unable to find namespace %s", namespace)
	}

	rcs, err := util.GetAllReplicationControllersForNamespace(ac.kubeClient, namespace)
	if err != nil {
		return errors.Annotatef(err, "unable to find deployments in %s", namespace)
	}

	var downstream bool
	for _, rc := range rcs.Items {
		if strings.Contains(rc.Name, "opssight") {
			downstream = true
			break
		}
	}
	// Delete the namespace
	err = util.DeleteNamespace(ac.kubeClient, namespace)
	if err != nil {
		return errors.Annotatef(err, "unable to delete namespace %s", namespace)
	}

	for {
		// Verify whether the namespace was deleted
		ns, err := util.GetNamespace(ac.kubeClient, namespace)
		log.Infof("namespace: %v, status: %v", namespace, ns.Status)
		if err != nil {
			log.Infof("deleted the namespace %+v", namespace)
			break
		}
		time.Sleep(10 * time.Second)
	}

	// Delete a Cluster Role
	clusterRoles := []string{}
	if downstream {
		clusterRoles = []string{"opssight-pod-processor", "opssight-image-processor"}
	} else {
		clusterRoles = []string{"pod-perceiver", "image-perceiver"}
	}

	for _, clusterRole := range clusterRoles {
		err := util.DeleteClusterRole(ac.kubeClient, clusterRole)
		if err != nil {
			log.Errorf("unable to delete the cluster role for %+v", clusterRole)
		}
	}

	// Delete a Cluster Role Binding
	clusterRoleBindings := []string{}
	if downstream {
		clusterRoleBindings = []string{"opssight-pod-processor", "opssight-image-processor", "opssight-scanner"}
	} else {
		clusterRoleBindings = []string{"pod-perceiver", "image-perceiver", "perceptor-scanner"}
	}
	for _, clusterRoleBinding := range clusterRoleBindings {
		err := util.DeleteClusterRoleBinding(ac.kubeClient, clusterRoleBinding)
		if err != nil {
			log.Errorf("unable to delete the cluster role binding for %+v", clusterRoleBinding)
		}
	}

	return nil
}

// CreateOpsSight will create the Black Duck OpsSight
func (ac *Creater) CreateOpsSight(createOpsSight *v1.OpsSightSpec) error {
	log.Debugf("create OpsSight details for %s: %+v", createOpsSight.Namespace, createOpsSight)

	// get the registry auth credentials for default OpenShift internal docker registries
	if !ac.config.DryRun {
		ac.addRegistryAuth(createOpsSight)
	}

	opssight := NewSpecConfig(createOpsSight)

	components, err := opssight.GetComponents()
	if err != nil {
		return errors.Annotatef(err, "unable to get opssight components for %s", createOpsSight.Namespace)
	}

	// setting up hub password in perceptor secret
	if !ac.config.DryRun {
		var hubPassword string
		var err error
		for dbInitTry := 0; dbInitTry < math.MaxInt32; dbInitTry++ {
			// get the secret from the default operator namespace, then copy it into the hub namespace.
			hubPassword, err = GetDefaultPasswords(ac.kubeClient, ac.config.Namespace)
			if err == nil {
				break
			} else {
				log.Infof("wasn't able to get hub password, sleeping 5 seconds.  try = %v", dbInitTry)
				time.Sleep(5 * time.Second)
			}
		}

		for _, secret := range components.Secrets {
			if strings.EqualFold(secret.GetName(), createOpsSight.SecretName) {
				secret.AddData(map[string][]byte{"HubUserPassword": []byte(hubPassword)})
				break
			}
		}
	}

	deployer, err := util.NewDeployer(ac.kubeConfig)
	if err != nil {
		return errors.Annotatef(err, "unable to get deployer object for %s", createOpsSight.Namespace)
	}
	// Note: controllers that need to continually run to update your app
	// should be added in PreDeploy().
	deployer.PreDeploy(components, createOpsSight.Namespace)

	if !ac.config.DryRun {
		err = deployer.Run()
		if err != nil {
			log.Errorf("unable to deploy opssight %s due to %+v", createOpsSight.Namespace, err)
		}
		deployer.StartControllers()
		// if OpenShift, add a privileged role to scanner account
		err = ac.postDeploy(opssight, createOpsSight.Namespace)
		if err != nil {
			return errors.Trace(err)
		}

		err = ac.deployHub(createOpsSight)
		if err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}

// GetDefaultPasswords returns admin,user,postgres passwords for db maintainance tasks.  Should only be used during
// initialization, or for 'babysitting' ephemeral hub instances (which might have postgres restarts)
// MAKE SURE YOU SEND THE NAMESPACE OF THE SECRET SOURCE (operator), NOT OF THE new hub  THAT YOUR TRYING TO CREATE !
func GetDefaultPasswords(kubeClient *kubernetes.Clientset, nsOfSecretHolder string) (hubPassword string, err error) {
	blackduckSecret, err := util.GetSecret(kubeClient, nsOfSecretHolder, "blackduck-secret")
	if err != nil {
		return "", errors.Annotate(err, "You need to first create a 'blackduck-secret' in this namespace with HUB_PASSWORD")
	}
	hubPassword = string(blackduckSecret.Data["HUB_PASSWORD"])

	// default named return
	return hubPassword, nil
}

func (ac *Creater) addRegistryAuth(opsSightSpec *v1.OpsSightSpec) {
	// if OpenShift, get the registry auth informations
	if ac.routeClient == nil {
		return
	}

	internalRegistries := []string{}
	route, err := util.GetOpenShiftRoutes(ac.routeClient, "default", "docker-registry")
	if err != nil {
		log.Errorf("unable to get docker-registry router in default namespace due to %+v", err)
	} else {
		internalRegistries = append(internalRegistries, route.Spec.Host)
		internalRegistries = append(internalRegistries, fmt.Sprintf("%s:443", route.Spec.Host))
	}

	registrySvc, err := util.GetService(ac.kubeClient, "default", "docker-registry")
	if err != nil {
		log.Errorf("unable to get docker-registry service in default namespace due to %+v", err)
	} else {
		if !strings.EqualFold(registrySvc.Spec.ClusterIP, "") {
			for _, port := range registrySvc.Spec.Ports {
				internalRegistries = append(internalRegistries, fmt.Sprintf("%s:%s", registrySvc.Spec.ClusterIP, strconv.Itoa(int(port.Port))))
				internalRegistries = append(internalRegistries, fmt.Sprintf("%s:%s", "docker-registry.default.svc", strconv.Itoa(int(port.Port))))
			}
		}
	}

	file, err := util.ReadFromFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		log.Errorf("unable to read the service account token file due to %+v", err)
	} else {
		for _, internalRegistry := range internalRegistries {
			registryAuth := v1.RegistryAuth{URL: internalRegistry, User: "admin", Password: string(file)}
			opsSightSpec.ScannerPod.ImageFacade.InternalRegistries = append(opsSightSpec.ScannerPod.ImageFacade.InternalRegistries, registryAuth)
		}
	}
}

func (ac *Creater) postDeploy(opssight *SpecConfig, namespace string) error {
	// Need to add the perceptor-scanner service account to the privileged scc
	if ac.osSecurityClient != nil {
		scannerServiceAccount := opssight.ScannerServiceAccount()
		perceiverServiceAccount := opssight.PodPerceiverServiceAccount()
		serviceAccounts := []string{fmt.Sprintf("system:serviceaccount:%s:%s", namespace, perceiverServiceAccount.GetName())}
		if !strings.EqualFold(opssight.config.ScannerPod.ImageFacade.ImagePullerType, "skopeo") {
			serviceAccounts = append(serviceAccounts, fmt.Sprintf("system:serviceaccount:%s:%s", namespace, scannerServiceAccount.GetName()))
		}
		return util.UpdateOpenShiftSecurityConstraint(ac.osSecurityClient, serviceAccounts, "privileged")
	}
	return nil
}

func (ac *Creater) deployHub(createOpsSight *v1.OpsSightSpec) error {
	if createOpsSight.Blackduck.InitialCount > createOpsSight.Blackduck.MaxCount {
		createOpsSight.Blackduck.InitialCount = createOpsSight.Blackduck.MaxCount
	}

	hubErrs := map[string]error{}
	for i := 0; i < createOpsSight.Blackduck.InitialCount; i++ {
		name := fmt.Sprintf("%s-%v", createOpsSight.Namespace, i)

		ns, err := util.CreateNamespace(ac.kubeClient, name)
		log.Debugf("created namespace: %+v", ns)
		if err != nil {
			log.Errorf("hub[%d]: unable to create the namespace due to %+v", i, err)
			hubErrs[name] = fmt.Errorf("unable to create the namespace due to %+v", err)
		}

		hubSpec := createOpsSight.Blackduck.BlackduckSpec
		hubSpec.Namespace = name
		createHub := &hub_v2.Blackduck{ObjectMeta: metav1.ObjectMeta{Name: name}, Spec: *hubSpec}
		log.Debugf("hub[%d]: %+v", i, createHub)
		_, err = util.CreateHub(ac.hubClient, name, createHub)
		if err != nil {
			log.Errorf("hub[%d]: unable to create the hub due to %+v", i, err)
			hubErrs[name] = fmt.Errorf("unable to create the hub due to %+v", err)
		}
	}

	return util.NewMapErrors(hubErrs)
}
