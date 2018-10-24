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
	"fmt"
	"time"

	"github.com/blackducksoftware/horizon/pkg/components"
	alertclientset "github.com/blackducksoftware/perceptor-protoform/pkg/alert/client/clientset/versioned"
	alertinformerv1 "github.com/blackducksoftware/perceptor-protoform/pkg/alert/client/informers/externalversions/alert/v1"
	alertcontroller "github.com/blackducksoftware/perceptor-protoform/pkg/alert/controller"
	"github.com/juju/errors"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	//_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"

	"github.com/blackducksoftware/perceptor-protoform/pkg/api/alert/v1"
	log "github.com/sirupsen/logrus"
)

// Controller defines the specification for the controller
type Controller struct {
	config *Config
}

// NewController will create a controller configuration
func NewController(config interface{}) (*Controller, error) {
	dependentConfig, ok := config.(*Config)
	if !ok {
		return nil, fmt.Errorf("failed to convert alert defaults: %v", config)
	}
	c := &Controller{config: dependentConfig}

	c.config.resyncPeriod = 0
	c.config.indexers = cache.Indexers{}

	return c, nil
}

// CreateClientSet will create the CRD client
func (c *Controller) CreateClientSet() error {
	alertClient, err := alertclientset.NewForConfig(c.config.KubeConfig)
	if err != nil {
		return errors.Trace(err)
	}
	c.config.customClientSet = alertClient
	return nil
}

// Deploy will deploy the CRD
func (c *Controller) Deploy() error {
	deployer, err := horizon.NewDeployer(c.config.KubeConfig)
	if err != nil {
		return err
	}

	// Hub CRD
	deployer.AddCustomDefinedResource(components.NewCustomResourceDefintion(horizonapi.CRDConfig{
		APIVersion: "apiextensions.k8s.io/v1beta1",
		Name:       "alerts.synopsys.com",
		Namespace:  c.config.Config.Namespace,
		Group:      "synopsys.com",
		CRDVersion: "v1",
		Kind:       "Alert",
		Plural:     "alerts",
		Singular:   "alert",
		Scope:      horizonapi.CRDClusterScoped,
	}))

	err = deployer.Run()
	if err != nil {
		log.Errorf("unable to create the alert CRD due to %+v", err)
	}

	time.Sleep(5 * time.Second)

	return err
}

// PostDeploy will initialize before deploying the CRD
func (c *Controller) PostDeploy() {
}

// CreateInformer will create a informer for the CRD
func (c *Controller) CreateInformer() {
	c.config.infomer = alertinformerv1.NewAlertInformer(
		c.config.customClientSet,
		c.config.Config.Namespace,
		c.config.resyncPeriod,
		c.config.indexers,
	)
}

// CreateQueue will create a queue to process the CRD
func (c *Controller) CreateQueue() {
	// create a new queue so that when the informer gets a resource that is either
	// a result of listing or watching, we can add an idenfitying key to the queue
	// so that it can be handled in the handler
	c.config.queue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
}

// AddInformerEventHandler will add the event handlers for the informers
func (c *Controller) AddInformerEventHandler() {
	c.config.infomer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// convert the resource object into a key (in this case
			// we are just doing it in the format of 'namespace/name')
			key, err := cache.MetaNamespaceKeyFunc(obj)
			log.Infof("add alert: %s", key)
			if err == nil {
				// add the key to the queue for the handler to get
				c.config.queue.Add(key)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			log.Infof("update alert: %s", key)
			if err == nil {
				c.config.queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// DeletionHandlingMetaNamsespaceKeyFunc is a helper function that allows
			// us to check the DeletedFinalStateUnknown existence in the event that
			// a resource was deleted but it is still contained in the index
			//
			// this then in turn calls MetaNamespaceKeyFunc
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			log.Infof("delete alert: %s: %+v", key, obj)

			if err == nil {
				c.config.queue.Add(key)
			}
		},
	})
}

// CreateHandler will create a CRD handler
func (c *Controller) CreateHandler() {
	c.config.handler = &alertcontroller.AlertHandler{
		Config:         c.config.KubeConfig,
		Clientset:      c.config.KubeClientSet,
		AlertClientset: c.config.customClientSet,
		Namespace:      c.config.Config.Namespace,
		CmMutex:        make(chan bool, 1),
		Defaults:       c.config.Defaults.(*v1.AlertSpec),
	}
}

// CreateController will create a CRD controller
func (c *Controller) CreateController() {
	c.config.controller = alertcontroller.NewController(
		&alertcontroller.Controller{
			Logger:         log.NewEntry(log.New()),
			Clientset:      c.config.KubeClientSet,
			Queue:          c.config.queue,
			Informer:       c.config.infomer,
			Handler:        c.config.handler,
			AlertClientset: c.config.customClientSet,
			Namespace:      c.config.Config.Namespace,
		})
}

// Run will run the CRD controller
func (c *Controller) Run() {
	go c.config.controller.Run(c.config.Threadiness, c.config.StopCh)
}

// PostRun will run post CRD controller execution
func (c *Controller) PostRun() {
}
