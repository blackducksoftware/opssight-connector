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
	"strings"

	"github.com/blackducksoftware/synopsys-operator/pkg/api"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	hubclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/crdupdater"
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
	config                  *protoform.Config
	kubeConfig              *rest.Config
	kubeClient              *kubernetes.Clientset
	opssightClient          *opssightclientset.Clientset
	osSecurityClient        *securityclient.SecurityV1Client
	routeClient             *routeclient.RouteV1Client
	hubClient               *hubclientset.Clientset
	isBlackDuckClusterScope bool
}

// NewCreater will instantiate the Creater
func NewCreater(config *protoform.Config, kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, opssightClient *opssightclientset.Clientset, osSecurityClient *securityclient.SecurityV1Client, routeClient *routeclient.RouteV1Client, hubClient *hubclientset.Clientset, isBlackDuckClusterScope bool) *Creater {
	return &Creater{
		config:                  config,
		kubeConfig:              kubeConfig,
		kubeClient:              kubeClient,
		opssightClient:          opssightClient,
		osSecurityClient:        osSecurityClient,
		routeClient:             routeClient,
		hubClient:               hubClient,
		isBlackDuckClusterScope: isBlackDuckClusterScope,
	}
}

// DeleteOpsSight will delete the OpsSight
func (ac *Creater) DeleteOpsSight(name string) error {
	log.Infof("deleting a %s OpsSight instance", name)
	values := strings.SplitN(name, "/", 2)
	var namespace string
	if len(values) == 0 {
		return fmt.Errorf("invalid name to delete the OpsSight instance")
	} else if len(values) == 1 {
		name = values[0]
		namespace = values[0]
		ns, err := util.ListNamespaces(ac.kubeClient, fmt.Sprintf("synopsys.com/%s.%s", util.OpsSightName, name))
		if err != nil {
			log.Errorf("unable to list %s OpsSight instance namespaces %s due to %+v", name, namespace, err)
		}
		if len(ns.Items) > 0 {
			namespace = ns.Items[0].Name
		} else {
			return fmt.Errorf("unable to find %s OpsSight instance namespace", name)
		}
	} else {
		name = values[1]
		namespace = values[0]
	}

	// delete the OpsSight instance
	commonConfig := crdupdater.NewCRUDComponents(ac.kubeConfig, ac.kubeClient, ac.config.DryRun, false, namespace, "",
		&api.ComponentList{}, fmt.Sprintf("app=%s,name=%s", util.OpsSightName, name), true)
	_, crudErrors := commonConfig.CRUDComponents()
	if len(crudErrors) > 0 {
		return fmt.Errorf("unable to delete the %s OpsSight instance in %s namespace due to %+v", name, namespace, crudErrors)
	}

	// delete namespace and if other apps are running, remove the Synopsys app label from the namespace
	var delErr error
	// if cluster scope, if no other instance running in Synopsys Operator namespace, delete the namespace or delete the Synopsys labels in the namespace
	if ac.config.IsClusterScoped {
		delErr = util.DeleteResourceNamespace(ac.kubeClient, util.OpsSightName, namespace, name, false)
	} else {
		// if namespace scope, delete the label from the namespace
		_, delErr = util.CheckAndUpdateNamespace(ac.kubeClient, util.OpsSightName, namespace, name, "", true)
	}
	if delErr != nil {
		return delErr
	}

	return nil
}

// CreateOpsSight will create the Black Duck OpsSight
func (ac *Creater) CreateOpsSight(opssight *opssightapi.OpsSight) error {
	// log.Debugf("create OpsSight details for %s: %+v", opssight.Namespace, opssight)
	opssightSpec := &opssight.Spec
	// get the registry auth credentials for default OpenShift internal docker registries
	if !ac.config.DryRun {
		ac.addRegistryAuth(opssightSpec)
	}

	spec := NewSpecConfig(ac.config, ac.kubeClient, ac.opssightClient, ac.hubClient, opssight, ac.config.DryRun, ac.isBlackDuckClusterScope)

	components, err := spec.GetComponents()
	if err != nil {
		return errors.Annotatef(err, "unable to get opssight components for %s", opssight.Spec.Namespace)
	}

	if !ac.config.DryRun {
		// call the CRUD updater to create or update opssight
		commonConfig := crdupdater.NewCRUDComponents(ac.kubeConfig, ac.kubeClient, ac.config.DryRun, false, opssightSpec.Namespace, "2.2.4",
			components, fmt.Sprintf("app=%s,name=%s", util.OpsSightName, opssight.Name), true)
		_, errs := commonConfig.CRUDComponents()

		if len(errs) > 0 {
			return fmt.Errorf("update components errors: %+v", errs)
		}

		// if OpenShift, add a privileged role to scanner account
		err = ac.postDeploy(spec, opssightSpec.Namespace)
		if err != nil {
			return errors.Annotatef(err, "post deploy")
		}

		err = ac.deployHub(opssightSpec)
		if err != nil {
			return errors.Annotatef(err, "deploy hub")
		}
	}

	return nil
}

// StopOpsSight will stop the Black Duck OpsSight
func (ac *Creater) StopOpsSight(opssight *opssightapi.OpsSightSpec) error {
	rcl, err := util.ListReplicationControllers(ac.kubeClient, opssight.Namespace, "app=opssight")
	for _, rc := range rcl.Items {
		if util.Int32ToInt(rc.Spec.Replicas) > 0 {
			_, err := util.PatchReplicationControllerForReplicas(ac.kubeClient, &rc, util.IntToInt32(0))
			if err != nil {
				return fmt.Errorf("unable to patch %s replication controller with replicas %d in %s namespace because %+v", rc.Name, 0, opssight.Namespace, err)
			}
		}
	}

	dpl, err := util.ListDeployments(ac.kubeClient, opssight.Namespace, "app=opssight")
	for _, dp := range dpl.Items {
		if util.Int32ToInt(dp.Spec.Replicas) > 0 {
			_, err := util.PatchDeploymentForReplicas(ac.kubeClient, &dp, util.IntToInt32(0))
			if err != nil {
				return fmt.Errorf("unable to patch %s deployment with replicas %d in %s namespace because %+v", dp.Name, 0, opssight.Namespace, err)
			}
		}
	}
	return err
}

// UpdateOpsSight will update the Black Duck OpsSight
func (ac *Creater) UpdateOpsSight(opssight *opssightapi.OpsSight) error {
	return ac.CreateOpsSight(opssight)
}

func (ac *Creater) addRegistryAuth(opsSightSpec *opssightapi.OpsSightSpec) {
	// if OpenShift, get the registry auth informations
	if ac.routeClient == nil {
		return
	}

	internalRegistries := []*string{}

	// Adding default image registry routes
	routes := map[string]string{"default": "docker-registry", "openshift-image-registry": "image-registry"}
	for namespace, name := range routes {
		route, err := util.GetRoute(ac.routeClient, namespace, name)
		if err != nil {
			continue
		}
		internalRegistries = append(internalRegistries, &route.Spec.Host)
		routeHostPort := fmt.Sprintf("%s:443", route.Spec.Host)
		internalRegistries = append(internalRegistries, &routeHostPort)
	}

	// Adding default OpenShift internal Docker/image registry service
	labelSelectors := []string{"docker-registry=default", "router in (router,router-default)"}
	for _, labelSelector := range labelSelectors {
		registrySvcs, err := util.ListServices(ac.kubeClient, "", labelSelector)
		if err != nil {
			continue
		}
		for _, registrySvc := range registrySvcs.Items {
			if !strings.EqualFold(registrySvc.Spec.ClusterIP, "") {
				for _, port := range registrySvc.Spec.Ports {
					clusterIPSvc := fmt.Sprintf("%s:%d", registrySvc.Spec.ClusterIP, port.Port)
					internalRegistries = append(internalRegistries, &clusterIPSvc)
					clusterIPSvcPort := fmt.Sprintf("%s.%s.svc:%d", registrySvc.Name, registrySvc.Namespace, port.Port)
					internalRegistries = append(internalRegistries, &clusterIPSvcPort)
				}
			}
		}
	}

	file, err := util.ReadFromFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		log.Errorf("unable to read the service account token file due to %+v", err)
	} else {
		for _, internalRegistry := range internalRegistries {
			opsSightSpec.ScannerPod.ImageFacade.InternalRegistries = append(opsSightSpec.ScannerPod.ImageFacade.InternalRegistries, &opssightapi.RegistryAuth{URL: *internalRegistry, User: "admin", Password: string(file)})
		}
	}
}

func (ac *Creater) postDeploy(spec *SpecConfig, namespace string) error {
	// Need to add the perceptor-scanner service account to the privileged scc
	if ac.config.IsOpenshift {
		scannerServiceAccount := spec.ScannerServiceAccount()
		perceiverServiceAccount := spec.PodPerceiverServiceAccount()
		serviceAccounts := []string{fmt.Sprintf("system:serviceaccount:%s:%s", namespace, perceiverServiceAccount.GetName())}
		if !strings.EqualFold(spec.opssight.Spec.ScannerPod.ImageFacade.ImagePullerType, "skopeo") {
			serviceAccounts = append(serviceAccounts, fmt.Sprintf("system:serviceaccount:%s:%s", namespace, scannerServiceAccount.GetName()))
		}
		return util.UpdateOpenShiftSecurityConstraint(ac.osSecurityClient, serviceAccounts, "privileged")
	}
	return nil
}

func (ac *Creater) deployHub(createOpsSight *opssightapi.OpsSightSpec) error {
	if createOpsSight.Blackduck.InitialCount > createOpsSight.Blackduck.MaxCount {
		createOpsSight.Blackduck.InitialCount = createOpsSight.Blackduck.MaxCount
	}

	hubErrs := map[string]error{}
	for i := 0; i < createOpsSight.Blackduck.InitialCount; i++ {
		name := fmt.Sprintf("%s-%v", createOpsSight.Namespace, i)

		_, err := util.GetNamespace(ac.kubeClient, name)
		if err == nil {
			continue
		}

		ns, err := util.CreateNamespace(ac.kubeClient, name)
		log.Debugf("created namespace: %+v", ns)
		if err != nil {
			log.Errorf("hub[%d]: unable to create the namespace due to %+v", i, err)
			hubErrs[name] = fmt.Errorf("unable to create the namespace due to %+v", err)
		}

		hubSpec := createOpsSight.Blackduck.BlackduckSpec
		hubSpec.Namespace = name
		createHub := &blackduckapi.Blackduck{ObjectMeta: metav1.ObjectMeta{Name: name}, Spec: *hubSpec}
		log.Debugf("hub[%d]: %+v", i, createHub)
		_, err = util.CreateBlackduck(ac.hubClient, name, createHub)
		if err != nil {
			log.Errorf("hub[%d]: unable to create the hub due to %+v", i, err)
			hubErrs[name] = fmt.Errorf("unable to create the hub due to %+v", err)
		}
	}

	return util.NewMapErrors(hubErrs)
}
