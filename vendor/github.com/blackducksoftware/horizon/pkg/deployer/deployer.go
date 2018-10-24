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

package deployer

import (
	"fmt"

	"github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/horizon/pkg/util"
	utilserror "github.com/blackducksoftware/horizon/pkg/util/error"

	"github.com/koki/short/converter/converters"
	shorttypes "github.com/koki/short/types"

	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	extensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	log "github.com/sirupsen/logrus"
)

// Deployer handles deploying the components to a cluster
type Deployer struct {
	replicationControllers map[string]*shorttypes.ReplicationController
	pods                   map[string]*shorttypes.Pod
	configMaps             map[string]*shorttypes.ConfigMap
	secrets                map[string]*shorttypes.Secret
	services               map[string]*shorttypes.Service
	serviceAccounts        map[string]*shorttypes.ServiceAccount
	deployments            map[string]*shorttypes.Deployment
	clusterRoles           map[string]*shorttypes.ClusterRole
	clusterRoleBindings    map[string]*shorttypes.ClusterRoleBinding
	crds                   map[string]*shorttypes.CustomResourceDefinition
	namespaces             map[string]*shorttypes.Namespace
	pvcs                   map[string]*shorttypes.PersistentVolumeClaim

	controllers map[string]api.DeployerControllerInterface

	client        *kubernetes.Clientset
	apiextensions *extensionsclient.Clientset
}

// NewDeployer creates a Deployer object
func NewDeployer(kubeconfig *rest.Config) (*Deployer, error) {
	// creates the client
	client, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error creating the kubernetes client: %v", err)
	}

	// creates the extensions client
	extensions, err := extensionsclient.NewForConfig(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error creating the kubernetes api extensions client: %v", err)
	}

	d := Deployer{
		client:                 client,
		apiextensions:          extensions,
		replicationControllers: make(map[string]*shorttypes.ReplicationController),
		pods:                make(map[string]*shorttypes.Pod),
		configMaps:          make(map[string]*shorttypes.ConfigMap),
		secrets:             make(map[string]*shorttypes.Secret),
		services:            make(map[string]*shorttypes.Service),
		serviceAccounts:     make(map[string]*shorttypes.ServiceAccount),
		deployments:         make(map[string]*shorttypes.Deployment),
		clusterRoles:        make(map[string]*shorttypes.ClusterRole),
		clusterRoleBindings: make(map[string]*shorttypes.ClusterRoleBinding),
		crds:                make(map[string]*shorttypes.CustomResourceDefinition),
		namespaces:          make(map[string]*shorttypes.Namespace),
		controllers:         make(map[string]api.DeployerControllerInterface),
		pvcs:                make(map[string]*shorttypes.PersistentVolumeClaim),
	}
	return &d, nil
}

// AddController will add a custom controller that will be run after all
// components have been deployed.
func (d *Deployer) AddController(name string, c api.DeployerControllerInterface) {
	d.controllers[name] = c
}

// AddConfigMap will add the provided config map to the config maps
// that will be deployed
func (d *Deployer) AddConfigMap(obj *components.ConfigMap) {
	d.configMaps[obj.GetName()] = obj.GetObj()
}

// AddDeployment will add the provided deployment to the deployments
// that will be deployed
func (d *Deployer) AddDeployment(obj *components.Deployment) {
	d.deployments[obj.GetName()] = obj.GetObj()
}

// AddService will add the provided service to the services
// that will be deployed
func (d *Deployer) AddService(obj *components.Service) {
	d.services[obj.GetName()] = obj.GetObj()
}

// AddSecret will add the provided secret to the secrets
// that will be deployed
func (d *Deployer) AddSecret(obj *components.Secret) {
	d.secrets[obj.GetName()] = obj.GetObj()
}

// AddClusterRole will add the provided cluster role to the
// cluster roles that will be deployed
func (d *Deployer) AddClusterRole(obj *components.ClusterRole) {
	d.clusterRoles[obj.GetName()] = obj.GetObj()
}

// AddClusterRoleBinding will add the provided cluster role binding
// to the cluster role bindings that will be deployed
func (d *Deployer) AddClusterRoleBinding(obj *components.ClusterRoleBinding) {
	d.clusterRoleBindings[obj.GetName()] = obj.GetObj()
}

// AddCustomDefinedResource will add the provided custom defined resource
// to the custom defined resources that will be deployed
func (d *Deployer) AddCustomDefinedResource(obj *components.CustomResourceDefinition) {
	d.crds[obj.GetName()] = obj.GetObj()
}

// AddReplicationController will add the provided replication controller
// to the replication controllers that will be deployed
func (d *Deployer) AddReplicationController(obj *components.ReplicationController) {
	d.replicationControllers[obj.GetName()] = obj.GetObj()
}

// AddNamespace will add the provided namespace to the
// namespaces that will be deployed
func (d *Deployer) AddNamespace(obj *components.Namespace) {
	d.namespaces[obj.GetName()] = obj.GetObj()
}

// AddServiceAccount will add the provided service account to the
// service accounts that will be deployed
func (d *Deployer) AddServiceAccount(obj *components.ServiceAccount) {
	d.serviceAccounts[obj.GetName()] = obj.GetObj()
}

// AddPod will add the provided pod to the pods that will be deployed
func (d *Deployer) AddPod(obj *components.Pod) {
	d.pods[obj.GetName()] = obj.GetObj()
}

// AddPVC will add the provided persistent volume claim to the
// persistent volume claims that will be deployed
func (d *Deployer) AddPVC(obj *components.PersistentVolumeClaim) {
	d.pvcs[obj.GetName()] = obj.GetObj()
}

// Run starts the deployer and deploys all components to the cluster
func (d *Deployer) Run() error {
	allErrs := map[util.ComponentType][]error{}

	err := d.deployNamespaces()
	if len(err) > 0 {
		allErrs[util.NamespaceComponent] = err
	}

	err = d.deployCRDs()
	if len(err) > 0 {
		allErrs[util.CRDComponent] = err
	}

	err = d.deployServiceAccounts()
	if len(err) > 0 {
		allErrs[util.ServiceAccountComponent] = err
	}

	errMap := d.deployRBAC()
	if len(errMap) > 0 {
		for k, v := range errMap {
			allErrs[k] = v
		}
	}

	err = d.deployConfigMaps()
	if len(err) > 0 {
		allErrs[util.ConfigMapComponent] = err
	}

	err = d.deploySecrets()
	if len(err) > 0 {
		allErrs[util.SecretComponent] = err
	}

	err = d.deployPVCs()
	if len(err) > 0 {
		allErrs[util.PersistentVolumeClaimComponent] = err
	}

	err = d.deployReplicationControllers()
	if len(err) > 0 {
		allErrs[util.ReplicationControllerComponent] = err
	}

	err = d.deployPods()
	if len(err) > 0 {
		allErrs[util.PodComponent] = err
	}

	err = d.deployDeployments()
	if len(err) > 0 {
		allErrs[util.DeploymentComponent] = err
	}

	err = d.deployServices()
	if len(err) > 0 {
		allErrs[util.ServiceComponent] = err
	}

	return utilserror.NewDeployErrors(allErrs)
}

// StartControllers will start all the configured controllers
func (d *Deployer) StartControllers(stopCh chan struct{}) map[string][]error {
	errs := make(map[string][]error)

	// Run the controllers if there are any configured
	if len(d.controllers) > 0 {
		resources := api.ControllerResources{
			KubeClient:           d.client,
			KubeExtensionsClient: d.apiextensions,
		}
		errCh := make(chan map[string]error)
		for n, c := range d.controllers {
			go func(name string, controller api.DeployerControllerInterface) {
				err := controller.Run(resources, stopCh)
				if err != nil {
					errCh <- map[string]error{name: err}
				}
			}(n, c)
		}

	controllerRun:
		for {
			select {
			case e := <-errCh:
				for k, v := range e {
					errs[k] = append(errs[k], v)
				}
			case <-stopCh:
				break controllerRun
			}
		}
	}
	return errs
}

func (d *Deployer) deployCRDs() []error {
	errs := []error{}

	for name, crdObj := range d.crds {
		wrapper := &shorttypes.CRDWrapper{CRD: *crdObj}
		crd, err := converters.Convert_Koki_CRD_to_Kube(wrapper)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		log.Infof("Creating custom defined resource %s", name)
		_, err = d.apiextensions.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (d *Deployer) deployServiceAccounts() []error {
	errs := []error{}

	for name, saObj := range d.serviceAccounts {
		wrapper := &shorttypes.ServiceAccountWrapper{ServiceAccount: *saObj}
		sa, err := converters.Convert_Koki_ServiceAccount_to_Kube_ServiceAccount(wrapper)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		log.Infof("Creating service account %s", name)
		_, err = d.client.Core().ServiceAccounts(sa.Namespace).Create(sa)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (d *Deployer) deployRBAC() map[util.ComponentType][]error {
	errs := map[util.ComponentType][]error{}

	for name, crObj := range d.clusterRoles {
		wrapper := &shorttypes.ClusterRoleWrapper{ClusterRole: *crObj}
		cr, err := converters.Convert_Koki_ClusterRole_to_Kube(wrapper)
		if err != nil {
			errs[util.ClusterRoleComponent] = append(errs[util.ClusterRoleComponent], err)
			continue
		}
		log.Infof("Creating cluster role %s", name)
		_, err = d.client.Rbac().ClusterRoles().Create(cr)
		if err != nil {
			errs[util.ClusterRoleComponent] = append(errs[util.ClusterRoleComponent], err)
		}
	}

	for name, crbObj := range d.clusterRoleBindings {
		wrapper := &shorttypes.ClusterRoleBindingWrapper{ClusterRoleBinding: *crbObj}
		crb, err := converters.Convert_Koki_ClusterRoleBinding_to_Kube(wrapper)
		if err != nil {
			errs[util.ClusterRoleBindingComponent] = append(errs[util.ClusterRoleComponent], err)
			continue
		}
		log.Infof("Creating cluster role binding %s", name)
		_, err = d.client.Rbac().ClusterRoleBindings().Create(crb)
		if err != nil {
			errs[util.ClusterRoleBindingComponent] = append(errs[util.ClusterRoleComponent], err)
		}
	}
	return errs
}

func (d *Deployer) deployConfigMaps() []error {
	errs := []error{}

	for name, cmObj := range d.configMaps {
		wrapper := &shorttypes.ConfigMapWrapper{ConfigMap: *cmObj}
		cm, err := converters.Convert_Koki_ConfigMap_to_Kube_v1_ConfigMap(wrapper)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		log.Infof("Creating config map %s", name)
		_, err = d.client.Core().ConfigMaps(cm.Namespace).Create(cm)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (d *Deployer) deploySecrets() []error {
	errs := []error{}

	for name, secretObj := range d.secrets {
		wrapper := &shorttypes.SecretWrapper{Secret: *secretObj}
		secret, err := converters.Convert_Koki_Secret_to_Kube_v1_Secret(wrapper)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		log.Infof("Creating secret %s", name)
		_, err = d.client.Core().Secrets(secret.Namespace).Create(secret)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (d *Deployer) deployReplicationControllers() []error {
	errs := []error{}

	for name, rcObj := range d.replicationControllers {
		wrapper := &shorttypes.ReplicationControllerWrapper{ReplicationController: *rcObj}
		rc, err := converters.Convert_Koki_ReplicationController_to_Kube_v1_ReplicationController(wrapper)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		log.Infof("Creating replication controller %s", name)
		_, err = d.client.Core().ReplicationControllers(rc.Namespace).Create(rc)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (d *Deployer) deployPods() []error {
	errs := []error{}

	for name, pObj := range d.pods {
		wrapper := &shorttypes.PodWrapper{Pod: *pObj}
		pod, err := converters.Convert_Koki_Pod_to_Kube_v1_Pod(wrapper)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		log.Infof("Creating pod %s", name)
		_, err = d.client.Core().Pods(pod.Namespace).Create(pod)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (d *Deployer) deployDeployments() []error {
	errs := []error{}

	for name, dObj := range d.deployments {
		wrapper := &shorttypes.DeploymentWrapper{Deployment: *dObj}
		deploy, err := converters.Convert_Koki_Deployment_to_Kube_apps_v1beta2_Deployment(wrapper)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		log.Infof("Creating deployment %s", name)
		_, err = d.client.AppsV1beta2().Deployments(deploy.Namespace).Create(deploy)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (d *Deployer) deployServices() []error {
	errs := []error{}

	for name, svcObj := range d.services {
		sWrapper := &shorttypes.ServiceWrapper{Service: *svcObj}
		svc, err := converters.Convert_Koki_Service_To_Kube_v1_Service(sWrapper)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		log.Infof("Creating service %s", name)
		_, err = d.client.Core().Services(svc.Namespace).Create(svc)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (d *Deployer) deployNamespaces() []error {
	errs := []error{}

	for name, nsObj := range d.namespaces {
		wrapper := &shorttypes.NamespaceWrapper{Namespace: *nsObj}
		ns, err := converters.Convert_Koki_Namespace_to_Kube_Namespace(wrapper)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		log.Infof("Creating namespace %s", name)
		_, err = d.client.Core().Namespaces().Create(ns)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (d *Deployer) deployPVCs() []error {
	errs := []error{}

	for name, pvcObj := range d.pvcs {
		wrapper := &shorttypes.PersistentVolumeClaimWrapper{PersistentVolumeClaim: *pvcObj}
		pvc, err := converters.Convert_Koki_PVC_to_Kube_PVC(wrapper)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		log.Infof("Creating persistent volume claim %s", name)
		p, ok := pvc.(*v1.PersistentVolumeClaim)
		if !ok {
			errs = append(errs, fmt.Errorf("failed to convert persistent volume claim %s", name))
			continue
		}
		_, err = d.client.Core().PersistentVolumeClaims(p.Namespace).Create(p)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (d *Deployer) Undeploy() error {
	allErrs := map[util.ComponentType][]error{}

	err := d.undeployServices()
	if len(err) > 0 {
		allErrs[util.ServiceComponent] = err
	}

	err = d.undeployDeployments()
	if len(err) > 0 {
		allErrs[util.DeploymentComponent] = err
	}

	err = d.undeployPods()
	if len(err) > 0 {
		allErrs[util.PodComponent] = err
	}

	err = d.undeployReplicationControllers()
	if len(err) > 0 {
		allErrs[util.ReplicationControllerComponent] = err
	}

	err = d.undeploySecrets()
	if len(err) > 0 {
		allErrs[util.SecretComponent] = err
	}

	err = d.undeployConfigMaps()
	if len(err) > 0 {
		allErrs[util.ConfigMapComponent] = err
	}

	errMap := d.undeployRBAC()
	if len(errMap) > 0 {
		for k, v := range errMap {
			allErrs[k] = v
		}
	}

	err = d.undeployServiceAccounts()
	if len(err) > 0 {
		allErrs[util.ServiceAccountComponent] = err
	}

	err = d.undeployCRDs()
	if len(err) > 0 {
		allErrs[util.CRDComponent] = err
	}

	err = d.undeployNamespaces()
	if len(err) > 0 {
		allErrs[util.NamespaceComponent] = err
	}

	return utilserror.NewDeployErrors(allErrs)
}

func (d *Deployer) undeployCRDs() []error {
	errs := []error{}

	for name := range d.crds {
		log.Infof("Deleting custom defined resource %s", name)
		err := d.apiextensions.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(name, &meta_v1.DeleteOptions{})
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (d *Deployer) undeployServiceAccounts() []error {
	errs := []error{}

	for name, saObj := range d.serviceAccounts {
		log.Infof("Deleting service account %s", name)
		err := d.client.Core().ServiceAccounts(saObj.Namespace).Delete(name, &meta_v1.DeleteOptions{})
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (d *Deployer) undeployRBAC() map[util.ComponentType][]error {
	errs := map[util.ComponentType][]error{}

	for name := range d.clusterRoles {
		log.Infof("Deleting cluster role %s", name)
		err := d.client.Rbac().ClusterRoles().Delete(name, &meta_v1.DeleteOptions{})
		if err != nil {
			errs[util.ClusterRoleComponent] = append(errs[util.ClusterRoleComponent], err)
		}
	}

	for name := range d.clusterRoleBindings {
		log.Infof("Deleting cluster role binding %s", name)
		err := d.client.Rbac().ClusterRoleBindings().Delete(name,  &meta_v1.DeleteOptions{})
		if err != nil {
			errs[util.ClusterRoleBindingComponent] = append(errs[util.ClusterRoleComponent], err)
		}
	}
	return errs
}

func (d *Deployer) undeployConfigMaps() []error {
	errs := []error{}

	for name, cmObj := range d.configMaps {
		log.Infof("Deleting config map %s", name)
		err := d.client.Core().ConfigMaps(cmObj.Namespace).Delete(name, &meta_v1.DeleteOptions{})
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (d *Deployer) undeploySecrets() []error {
	errs := []error{}

	for name, secretObj := range d.secrets {
		log.Infof("Deleting secret %s", name)
		err := d.client.Core().Secrets(secretObj.Namespace).Delete(name, &meta_v1.DeleteOptions{})
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (d *Deployer) undeployReplicationControllers() []error {
	errs := []error{}

	for name, rcObj := range d.replicationControllers {
		log.Infof("Deleting replication controller %s", name)
		err := d.client.Core().ReplicationControllers(rcObj.Namespace).Delete(name, &meta_v1.DeleteOptions{})
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (d *Deployer) undeployPods() []error {
	errs := []error{}

	for name, pObj := range d.pods {
		log.Infof("Deleting pod %s", name)
		err := d.client.Core().Pods(pObj.Namespace).Delete(name, &meta_v1.DeleteOptions{})
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (d *Deployer) undeployDeployments() []error {
	errs := []error{}

	for name, dObj := range d.deployments {
		log.Infof("Deleting deployment %s", name)
		propagationPolicy := meta_v1.DeletePropagationBackground
		err := d.client.AppsV1beta2().Deployments(dObj.Namespace).Delete(name, &meta_v1.DeleteOptions{
			PropagationPolicy: &propagationPolicy,
		})
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (d *Deployer) undeployServices() []error {
	errs := []error{}

	for name, svcObj := range d.services {
		log.Infof("Deleting service %s", name)
		err := d.client.Core().Services(svcObj.Namespace).Delete(name, &meta_v1.DeleteOptions{})
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (d *Deployer) undeployNamespaces() []error {
	errs := []error{}

	for name := range d.namespaces {
		log.Infof("Deleting namespace %s", name)
		err := d.client.Core().Namespaces().Delete(name, &meta_v1.DeleteOptions{})
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}
