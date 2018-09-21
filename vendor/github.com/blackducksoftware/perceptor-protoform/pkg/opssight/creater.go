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
	"strconv"
	"strings"
	"time"

	"github.com/blackducksoftware/perceptor-protoform/pkg/api/opssight/v1"
	"github.com/blackducksoftware/perceptor-protoform/pkg/model"
	opssightclientset "github.com/blackducksoftware/perceptor-protoform/pkg/opssight/client/clientset/versioned"
	"github.com/blackducksoftware/perceptor-protoform/pkg/opssight/plugins"
	"github.com/blackducksoftware/perceptor-protoform/pkg/util"
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
	config           *model.Config
	kubeConfig       *rest.Config
	kubeClient       *kubernetes.Clientset
	opssightClient   *opssightclientset.Clientset
	osSecurityClient *securityclient.SecurityV1Client
	routeClient      *routeclient.RouteV1Client
}

// NewCreater will instantiate the Creater
func NewCreater(config *model.Config, kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, opssightClient *opssightclientset.Clientset, osSecurityClient *securityclient.SecurityV1Client, routeClient *routeclient.RouteV1Client) *Creater {
	return &Creater{config: config, kubeConfig: kubeConfig, kubeClient: kubeClient, opssightClient: opssightClient, osSecurityClient: osSecurityClient, routeClient: routeClient}
}

// DeleteOpsSight will delete the Black Duck OpsSight
func (ac *Creater) DeleteOpsSight(namespace string) error {
	log.Debugf("Delete OpsSight details for %s", namespace)
	// Verify that the namespace exists
	_, err := util.GetNamespace(ac.kubeClient, namespace)
	if err != nil {
		return errors.Annotatef(err, "Unable to find namespace %s", namespace)
	}
	// Delete the namespace
	err = util.DeleteNamespace(ac.kubeClient, namespace)
	if err != nil {
		return errors.Annotatef(err, "Unable to delete namespace %s", namespace)
	}

	for {
		// Verify whether the namespace was deleted
		ns, err := util.GetNamespace(ac.kubeClient, namespace)
		log.Infof("Namespace: %v, status: %v", namespace, ns.Status)
		time.Sleep(10 * time.Second)
		if err != nil {
			log.Infof("Deleted the namespace %+v", namespace)
			break
		}
	}
	return nil
}

// CreateOpsSight will create the Black Duck OpsSight
func (ac *Creater) CreateOpsSight(createOpsSight *v1.OpsSightSpec) error {
	log.Debugf("Create OpsSight details for %s: %+v", createOpsSight.Namespace, createOpsSight)

	// get the registry auth credentials for default OpenShift internal docker registries
	ac.addRegistryAuth(createOpsSight)

	opssight := NewSpecConfig(createOpsSight)

	components, err := opssight.GetComponents()
	if err != nil {
		return errors.Annotatef(err, "unable to get opssight components for %s", createOpsSight.Namespace)
	}
	deployer, err := util.NewDeployer(ac.kubeConfig)
	if err != nil {
		return errors.Annotatef(err, "unable to get deployer object for %s", createOpsSight.Namespace)
	}
	// Note: controllers that need to continually run to update your app
	// should be added in PreDeploy().
	deployer.PreDeploy(components, createOpsSight.Namespace)

	// Any new, pluggable maintainance stuff should go in here...
	deployer.AddController("perceptor_configmap_controller", &plugins.PerceptorConfigMap{Config: ac.config, KubeConfig: ac.kubeConfig, OpsSightClient: ac.opssightClient, Namespace: createOpsSight.Namespace})

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

	return nil
}

func (ac *Creater) addRegistryAuth(opsSightSpec *v1.OpsSightSpec) {
	// if OpenShift, get the registry auth informations
	if ac.osSecurityClient != nil {
		var internalRegistries []string
		route, err := ac.routeClient.Routes("default").Get("docker-registry", metav1.GetOptions{})
		if err != nil {
			log.Errorf("unable to get docker-registry router in default namespace due to %+v", err)
		} else {
			internalRegistries = append(internalRegistries, route.Spec.Host)
			internalRegistries = append(internalRegistries, fmt.Sprintf("%s:443", route.Spec.Host))
		}

		registrySvc, err := ac.kubeClient.CoreV1().Services("default").Get("docker-registry", metav1.GetOptions{})
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
				opsSightSpec.InternalRegistries = append(opsSightSpec.InternalRegistries, registryAuth)
			}
		}
	}
}

func (ac *Creater) postDeploy(opssight *SpecConfig, namespace string) error {
	if ac.osSecurityClient != nil {
		// Need to add the perceptor-scanner service account to the privelged scc
		scc, err := ac.osSecurityClient.SecurityContextConstraints().Get("privileged", metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get scc privileged: %v", err)
		}

		var scannerAccount string
		s := opssight.ScannerServiceAccount()
		scannerAccount = fmt.Sprintf("system:serviceaccount:%s:%s", namespace, s.GetName())

		// Only add the service account if it isn't already in the list of users for the privileged scc
		exists := false
		for _, u := range scc.Users {
			if strings.Compare(u, scannerAccount) == 0 {
				exists = true
				break
			}
		}

		if !exists {
			scc.Users = append(scc.Users, scannerAccount)

			_, err = ac.osSecurityClient.SecurityContextConstraints().Update(scc)
			if err != nil {
				return fmt.Errorf("failed to update scc privileged: %v", err)
			}
		}
	}

	return nil
}
