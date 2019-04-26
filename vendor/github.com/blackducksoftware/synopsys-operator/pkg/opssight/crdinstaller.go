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
	"reflect"
	"strings"
	"time"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	hubclient "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	opssightinformer "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/informers/externalversions/opssight/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	securityclient "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// CRDInstaller defines the specification
type CRDInstaller struct {
	config *Config
}

// NewCRDInstaller will create a controller configuration
func NewCRDInstaller(config interface{}) (*CRDInstaller, error) {
	dependentConfig, ok := config.(*Config)
	if !ok {
		return nil, errors.Errorf("failed to convert opssight defaults: %v", config)
	}
	c := &CRDInstaller{config: dependentConfig}

	log.Debugf("resync period: %d", c.config.Config.ResyncIntervalInSeconds)
	c.config.resyncPeriod = time.Duration(c.config.Config.ResyncIntervalInSeconds) * time.Second
	c.config.indexers = cache.Indexers{}

	return c, nil
}

// CreateClientSet will create the CRD client
func (c *CRDInstaller) CreateClientSet() error {
	opssightClient, err := opssightclientset.NewForConfig(c.config.KubeConfig)
	if err != nil {
		return errors.Annotate(err, "Unable to create OpsSight informer client")
	}
	c.config.customClientSet = opssightClient
	return nil
}

// Deploy will deploy the CRD
func (c *CRDInstaller) Deploy() error {
	deployer, err := horizon.NewDeployer(c.config.KubeConfig)
	if err != nil {
		return errors.Trace(err)
	}

	// OpsSight CRD
	deployer.AddCustomDefinedResource(components.NewCustomResourceDefintion(horizonapi.CRDConfig{
		APIVersion: "apiextensions.k8s.io/v1beta1",
		Name:       "opssights.synopsys.com",
		Namespace:  c.config.Config.Namespace,
		Group:      "synopsys.com",
		CRDVersion: "v1",
		Kind:       "OpsSight",
		Plural:     "opssights",
		Singular:   "opssight",
		Scope:      horizonapi.CRDClusterScoped,
	}))

	err = deployer.Run()
	if err != nil {
		log.Errorf("unable to create the opssight CRD due to %+v", err)
		// return errors.Trace(err)
	}

	time.Sleep(5 * time.Second)

	// Any new, pluggable maintainance stuff should go in here...
	hubClientset, err := hubclient.NewForConfig(c.config.KubeConfig)
	if err != nil {
		return errors.Trace(err)
	}
	configMapEditor := NewUpdater(c.config.Config, c.config.KubeClientSet, hubClientset, c.config.customClientSet)
	configMapEditor.Run(c.config.StopCh)

	return nil
}

// PostDeploy will initialize before deploying the CRD
func (c *CRDInstaller) PostDeploy() {
}

// CreateInformer will create a informer for the CRD
func (c *CRDInstaller) CreateInformer() {
	c.config.informer = opssightinformer.NewOpsSightInformer(
		c.config.customClientSet,
		c.config.Config.Namespace,
		c.config.resyncPeriod,
		c.config.indexers,
	)
}

// CreateQueue will create a queue to process the CRD
func (c *CRDInstaller) CreateQueue() {
	// create a new queue so that when the informer gets a resource that is either
	// a result of listing or watching, we can add an idenfitying key to the queue
	// so that it can be handled in the handler
	c.config.queue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
}

// AddInformerEventHandler will add the event handlers for the informers
func (c *CRDInstaller) AddInformerEventHandler() {
	log.Debugf("adding informer event handler for opssight")
	c.config.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// convert the resource object into a key (in this case
			// we are just doing it in the format of 'namespace/name')
			key, err := cache.MetaNamespaceKeyFunc(obj)
			log.Infof("add opssight: %s", key)
			if err == nil {
				// add the key to the queue for the handler to get
				c.config.queue.Add(key)
			} else {
				log.Errorf("unable to add OpsSight: %v", err)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			old := oldObj.(*opssightapi.OpsSight)
			new := newObj.(*opssightapi.OpsSight)
			if strings.EqualFold(old.Status.State, string(Running)) || !reflect.DeepEqual(old.Spec, new.Spec) || !reflect.DeepEqual(old.Status.InternalHosts, new.Status.InternalHosts) {
				key, err := cache.MetaNamespaceKeyFunc(newObj)
				log.Infof("update opssight: %s", key)
				if err == nil {
					c.config.queue.Add(key)
				} else {
					log.Errorf("unable to update OpsSight: %v", err)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			// DeletionHandlingMetaNamespaceKeyFunc is a helper function that allows
			// us to check the DeletedFinalStateUnknown existence in the event that
			// a resource was deleted but it is still contained in the index
			//
			// this then in turn calls MetaNamespaceKeyFunc
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			log.Infof("delete opssight: %s: %+v", key, obj)

			if err == nil {
				c.config.queue.Add(key)
			} else {
				log.Errorf("unable to delete OpsSight: %v", err)
			}
		},
	})
}

// CreateHandler will create a CRD handler
func (c *CRDInstaller) CreateHandler() {
	osClient, err := securityclient.NewForConfig(c.config.KubeConfig)
	if err != nil {
		osClient = nil
	} else {
		_, err := util.GetOpenShiftSecurityConstraint(osClient, "privileged")
		if err != nil && strings.Contains(err.Error(), "could not find the requested resource") && strings.Contains(err.Error(), "openshift.io") {
			log.Debugf("Ignoring scc privileged for kubernetes cluster")
			osClient = nil
		}
	}

	routeClient := util.GetRouteClient(c.config.KubeConfig)

	hubClient, err := hubclient.NewForConfig(c.config.KubeConfig)
	if err != nil {
		log.Errorf("unable to create the hub client for opssight: %+v", err)
		return
	}

	c.config.handler = &Handler{
		Config:           c.config.Config,
		KubeConfig:       c.config.KubeConfig,
		KubeClient:       c.config.KubeClientSet,
		OpsSightClient:   c.config.customClientSet,
		Namespace:        c.config.Config.Namespace,
		OSSecurityClient: osClient,
		RouteClient:      routeClient,
		Defaults:         c.config.Defaults.(*opssightapi.OpsSightSpec),
		HubClient:        hubClient,
	}
}

// CreateController will create a CRD controller
func (c *CRDInstaller) CreateController() {
	c.config.controller = NewController(
		&Controller{
			Logger:            log.NewEntry(log.New()),
			Clientset:         c.config.KubeClientSet,
			Queue:             c.config.queue,
			Informer:          c.config.informer,
			Handler:           c.config.handler,
			OpsSightClientset: c.config.customClientSet,
			Namespace:         c.config.Config.Namespace,
		})
}

// Run will run the CRD controller
func (c *CRDInstaller) Run() {
	go c.config.controller.Run(c.config.Threadiness, c.config.StopCh)
}

// PostRun will run post CRD controller execution
func (c *CRDInstaller) PostRun() {
}
