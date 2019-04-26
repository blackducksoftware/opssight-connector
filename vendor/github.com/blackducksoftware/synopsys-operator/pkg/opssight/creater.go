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
	rbacv1 "k8s.io/api/rbac/v1"
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

	// get all replication controller for the namespace
	rcs, err := util.ListReplicationControllers(ac.kubeClient, namespace, "")
	if err != nil {
		return errors.Annotatef(err, "unable to list the replication controller in %s", namespace)
	}

	// get only opssight related replication controller for the namespace
	opssightRCs, err := util.ListReplicationControllers(ac.kubeClient, namespace, "app=opssight")
	if err != nil {
		return errors.Annotatef(err, "unable to list the opssight's replication controller in %s", namespace)
	}

	// if both the length same, then delete the namespace, if different, delete only the replication controller
	if len(rcs.Items) == len(opssightRCs.Items) {
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
	} else {
		// delete the replication controller
		for _, opssightRC := range opssightRCs.Items {
			err = util.DeleteReplicationController(ac.kubeClient, namespace, opssightRC.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete the %s replication controller in %s namespace", opssightRC.GetName(), namespace)
			}
		}
	}

	clusterRoleBindings, err := util.ListClusterRoleBindings(ac.kubeClient, "app=opssight")

	for _, clusterRoleBinding := range clusterRoleBindings.Items {
		if len(clusterRoleBinding.Subjects) == 1 {
			if !strings.EqualFold(clusterRoleBinding.RoleRef.Name, "synopsys-operator-admin") {
				log.Debugf("deleting cluster role %s", clusterRoleBinding.RoleRef.Name)
				err = util.DeleteClusterRole(ac.kubeClient, clusterRoleBinding.RoleRef.Name)
				if err != nil {
					log.Errorf("unable to delete the cluster role for %+v", clusterRoleBinding.RoleRef.Name)
				}
			}

			log.Debugf("deleting cluster role binding %s", clusterRoleBinding.GetName())
			err = util.DeleteClusterRoleBinding(ac.kubeClient, clusterRoleBinding.GetName())
			if err != nil {
				log.Errorf("unable to delete the cluster role binding for %+v", clusterRoleBinding.GetName())
			}
		} else {
			log.Debugf("updating cluster role binding %s", clusterRoleBinding.GetName())
			clusterRoleBinding.Subjects = removeSubjects(clusterRoleBinding.Subjects, namespace)
			_, err = util.UpdateClusterRoleBinding(ac.kubeClient, &clusterRoleBinding)
			if err != nil {
				log.Errorf("unable to update the cluster role binding for %+v", clusterRoleBinding.GetName())
			}
		}
	}

	return nil
}

func removeSubjects(subjects []rbacv1.Subject, namespace string) []rbacv1.Subject {
	newSubjects := []rbacv1.Subject{}
	for _, subject := range subjects {
		if !strings.EqualFold(subject.Namespace, namespace) {
			newSubjects = append(newSubjects, subject)
		}
	}
	return newSubjects
}

// CreateOpsSight will create the Black Duck OpsSight
func (ac *Creater) CreateOpsSight(opssight *opssightapi.OpsSight) error {
	log.Debugf("create OpsSight details for %s: %+v", opssight.Namespace, opssight)
	opssightSpec := &opssight.Spec
	// get the registry auth credentials for default OpenShift internal docker registries
	if !ac.config.DryRun {
		ac.addRegistryAuth(opssightSpec)
	}

	spec := NewSpecConfig(ac.config, ac.kubeClient, ac.opssightClient, ac.hubClient, opssight, ac.config.DryRun)

	components, err := spec.GetComponents()
	if err != nil {
		return errors.Annotatef(err, "unable to get opssight components for %s", opssight.Spec.Namespace)
	}

	commonConfig := crdupdater.NewCRUDComponents(ac.kubeConfig, ac.kubeClient, ac.config.DryRun, opssightSpec.Namespace, components, "app=opssight")
	errs := commonConfig.CRUDComponents()

	if len(errs) > 0 {
		return fmt.Errorf("update components errors: %+v", errs)
	}

	if !ac.config.DryRun {
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
	return err
}

// UpdateOpsSight will update the Black Duck OpsSight
func (ac *Creater) UpdateOpsSight(opssight *opssightapi.OpsSight) error {
	newConfigMapConfig := NewSpecConfig(ac.config, ac.kubeClient, ac.opssightClient, ac.hubClient, opssight, ac.config.DryRun)

	opssightSpec := &opssight.Spec
	// get new components build from the latest updates
	components, err := newConfigMapConfig.GetComponents()
	if err != nil {
		return errors.Annotatef(err, "unable to get opssight components for %s", opssightSpec.Namespace)
	}

	commonConfig := crdupdater.NewCRUDComponents(ac.kubeConfig, ac.kubeClient, ac.config.DryRun, opssightSpec.Namespace, components, "app=opssight")
	errors := commonConfig.CRUDComponents()

	if len(errors) > 0 {
		return fmt.Errorf("unable to update components due to %+v", errors)
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

func (ac *Creater) addRegistryAuth(opsSightSpec *opssightapi.OpsSightSpec) {
	// if OpenShift, get the registry auth informations
	if ac.routeClient == nil {
		return
	}

	internalRegistries := []*string{}
	route, err := util.GetOpenShiftRoutes(ac.routeClient, "default", "docker-registry")
	if err != nil {
		log.Errorf("unable to get docker-registry router in default namespace due to %+v", err)
	} else {
		internalRegistries = append(internalRegistries, &route.Spec.Host)
		routeHostPort := fmt.Sprintf("%s:443", route.Spec.Host)
		internalRegistries = append(internalRegistries, &routeHostPort)
	}

	registrySvc, err := util.GetService(ac.kubeClient, "default", "docker-registry")
	if err != nil {
		log.Errorf("unable to get docker-registry service in default namespace due to %+v", err)
	} else {
		if !strings.EqualFold(registrySvc.Spec.ClusterIP, "") {
			for _, port := range registrySvc.Spec.Ports {
				clusterIPSvc := fmt.Sprintf("%s:%s", registrySvc.Spec.ClusterIP, strconv.Itoa(int(port.Port)))
				internalRegistries = append(internalRegistries, &clusterIPSvc)
				clusterIPSvcPort := fmt.Sprintf("%s:%s", "docker-registry.default.svc", strconv.Itoa(int(port.Port)))
				internalRegistries = append(internalRegistries, &clusterIPSvcPort)
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
	if ac.osSecurityClient != nil {
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
		_, err = util.CreateHub(ac.hubClient, name, createHub)
		if err != nil {
			log.Errorf("hub[%d]: unable to create the hub due to %+v", i, err)
			hubErrs[name] = fmt.Errorf("unable to create the hub due to %+v", err)
		}
	}

	return util.NewMapErrors(hubErrs)
}
