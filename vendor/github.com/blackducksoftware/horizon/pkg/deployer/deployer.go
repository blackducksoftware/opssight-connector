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
	"bytes"
	"fmt"

	"github.com/blackducksoftware/horizon/pkg/api"
	utilserror "github.com/blackducksoftware/horizon/pkg/util/error"

	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	extensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	log "github.com/sirupsen/logrus"
)

var deployOrder = []api.ComponentType{
	api.NamespaceComponent,
	api.CRDComponent,
	api.ServiceAccountComponent,
	api.ClusterRoleComponent,
	api.ClusterRoleBindingComponent,
	api.RoleComponent,
	api.RoleBindingComponent,
	api.ConfigMapComponent,
	api.SecretComponent,
	api.PersistentVolumeClaimComponent,
	api.ReplicationControllerComponent,
	api.PodComponent,
	api.DeploymentComponent,
	api.ServiceComponent,
	api.JobComponent,
	api.HorizontalPodAutoscalerComponent,
	api.IngressComponent,
	api.StatefulSetComponent,
	api.DaemonSetComponent,
}

// Deployer handles deploying the components to a cluster
type Deployer struct {
	components  map[api.ComponentType][]api.DeployableComponentInterface
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

	d := createDeployer()
	d.client = client
	d.apiextensions = extensions
	return d, nil
}

// NewDeployerExporter creates a Deployer object that only supports exporting
func NewDeployerExporter() *Deployer {
	return createDeployer()
}

func createDeployer() *Deployer {
	return &Deployer{
		components:  make(map[api.ComponentType][]api.DeployableComponentInterface),
		controllers: make(map[string]api.DeployerControllerInterface),
	}
}

// AddController will add a custom controller that will be run after all
// components have been deployed.
func (d *Deployer) AddController(name string, c api.DeployerControllerInterface) {
	d.controllers[name] = c
}

// AddComponent will add a component to be deployed
func (d *Deployer) AddComponent(kind api.ComponentType, obj api.DeployableComponentInterface) {
	d.components[kind] = append(d.components[kind], obj)
}

func (d *Deployer) exporterOnly() bool {
	if d.client == nil {
		return true
	}

	if d.apiextensions == nil {
		return true
	}
	return false
}

// Run starts the deployer and deploys all components to the cluster
func (d *Deployer) Run() error {
	if d.exporterOnly() {
		return fmt.Errorf("deployer has no clients defined and can only be used to export")
	}

	allErrs := map[api.ComponentType][]error{}
	resources := d.getResources()
	for _, ct := range deployOrder {
		for _, c := range d.components[ct] {
			log.Infof("creating %s %s", ct, c.GetName())
			err := c.Deploy(resources)
			if err != nil {
				allErrs[ct] = append(allErrs[ct], err)
			}
		}
	}

	return utilserror.NewDeployErrors(allErrs)
}

func (d *Deployer) getResources() api.DeployerResources {
	return api.DeployerResources{
		KubeClient:           d.client,
		KubeExtensionsClient: d.apiextensions,
	}
}

// StartControllers will start all the configured controllers
func (d *Deployer) StartControllers(stopCh chan struct{}) map[string][]error {
	errs := make(map[string][]error)

	if d.exporterOnly() {
		errs["deployerCore"] = []error{fmt.Errorf("deployer has no clients defined and can only be used to export")}
		return errs
	}

	// Run the controllers if there are any configured
	resources := d.getResources()
	if len(d.controllers) > 0 {
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

// Undeploy will remove all components from the cluster
func (d *Deployer) Undeploy() error {
	var err error
	if d.exporterOnly() {
		return fmt.Errorf("deployer has no clients defined and can only be used to export")
	}

	allErrs := map[api.ComponentType][]error{}
	resources := d.getResources()
	for cnt := len(deployOrder) - 1; cnt >= 0; cnt-- {
		cType := deployOrder[cnt]
		for _, c := range d.components[cType] {
			log.Infof("deleting %s %s", cType, c.GetName())
			err = c.Undeploy(resources)
			if err != nil {
				allErrs[cType] = append(allErrs[cType], err)
			}
		}
	}

	return utilserror.NewDeployErrors(allErrs)
}

// Export returns api string objects for all types
func (d *Deployer) Export() map[string]string {
	ser := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme,
		scheme.Scheme)

	// append more yaml
	m := map[string]string{}
	appender := func(s string, obj api.DeployableComponentInterface) {
		buf := bytes.NewBufferString("")
		err := ser.Encode(obj, buf)
		if err != nil {
			panic(err)
		}
		m[s] = fmt.Sprintf("%v \n---", buf.String())
	}

	for _, ct := range deployOrder {
		for _, c := range d.components[ct] {
			appender(c.GetName(), c)
		}
	}

	return m
}
