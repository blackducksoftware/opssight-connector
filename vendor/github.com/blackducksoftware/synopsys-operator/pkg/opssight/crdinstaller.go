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

	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	hubclient "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	opssightinformer "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/informers/externalversions/opssight/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	securityclient "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// CRDInstaller defines the specification
type CRDInstaller struct {
	config         *protoform.Config
	kubeConfig     *rest.Config
	kubeClient     *kubernetes.Clientset
	defaults       interface{}
	resyncPeriod   time.Duration
	indexers       cache.Indexers
	informer       cache.SharedIndexInformer
	queue          workqueue.RateLimitingInterface
	handler        *Handler
	controller     *Controller
	opssightclient *opssightclientset.Clientset
	stopCh         <-chan struct{}
}

// NewCRDInstaller will create a controller configuration
func NewCRDInstaller(config *protoform.Config, kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, defaults interface{}, stopCh <-chan struct{}) *CRDInstaller {
	crdInstaller := &CRDInstaller{config: config, kubeConfig: kubeConfig, kubeClient: kubeClient, defaults: defaults, stopCh: stopCh}
	log.Debugf("resync period: %d", config.ResyncIntervalInSeconds)
	crdInstaller.resyncPeriod = time.Duration(config.ResyncIntervalInSeconds) * time.Second
	crdInstaller.indexers = cache.Indexers{}
	return crdInstaller
}

// CreateClientSet will create the CRD client
func (c *CRDInstaller) CreateClientSet() error {
	opssightClient, err := opssightclientset.NewForConfig(c.kubeConfig)
	if err != nil {
		return errors.Annotate(err, "Unable to create OpsSight informer client")
	}
	c.opssightclient = opssightClient
	return nil
}

// Deploy will deploy the CRD
func (c *CRDInstaller) Deploy() error {
	// Any new, pluggable maintainance stuff should go in here...
	blackDuckClient, err := hubclient.NewForConfig(c.kubeConfig)
	if err != nil {
		return errors.Trace(err)
	}
	crdUpdater := NewUpdater(c.config, c.kubeClient, blackDuckClient, c.opssightclient)
	go crdUpdater.Run(c.stopCh)
	return nil
}

// PostDeploy will initialize before deploying the CRD
func (c *CRDInstaller) PostDeploy() {
}

// CreateInformer will create a informer for the CRD
func (c *CRDInstaller) CreateInformer() {
	c.informer = opssightinformer.NewOpsSightInformer(
		c.opssightclient,
		c.config.CrdNamespace,
		c.resyncPeriod,
		c.indexers,
	)
}

// CreateQueue will create a queue to process the CRD
func (c *CRDInstaller) CreateQueue() {
	// create a new queue so that when the informer gets a resource that is either
	// a result of listing or watching, we can add an idenfitying key to the queue
	// so that it can be handled in the handler
	c.queue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
}

// AddInformerEventHandler will add the event handlers for the informers
func (c *CRDInstaller) AddInformerEventHandler() {
	c.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// convert the resource object into a key (in this case
			// we are just doing it in the format of 'namespace/name')
			key, err := cache.MetaNamespaceKeyFunc(obj)
			log.Infof("add opssight: %s", key)
			if err == nil {
				// add the key to the queue for the handler to get
				c.queue.Add(key)
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
					c.queue.Add(key)
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
				c.queue.Add(key)
			} else {
				log.Errorf("unable to delete OpsSight: %v", err)
			}
		},
	})
}

// CreateHandler will create a CRD handler
func (c *CRDInstaller) CreateHandler() {
	hubClient, err := hubclient.NewForConfig(c.kubeConfig)
	if err != nil {
		log.Errorf("unable to create the hub client for opssight: %+v", err)
		return
	}

	c.handler = &Handler{
		Config:         c.config,
		KubeConfig:     c.kubeConfig,
		KubeClient:     c.kubeClient,
		OpsSightClient: c.opssightclient,
		Namespace:      c.config.Namespace,
		Defaults:       c.defaults.(*opssightapi.OpsSightSpec),
		HubClient:      hubClient,
	}

	if util.IsOpenshift(c.kubeClient) {
		c.handler.RouteClient = util.GetRouteClient(c.kubeConfig, c.kubeClient, c.config.Namespace)
		if c.handler.OSSecurityClient, err = securityclient.NewForConfig(c.kubeConfig); err != nil {
			log.Errorf("error in creating the OpenShift security client due to %+v", err)
			return
		}
	}
}

// CreateController will create a CRD controller
func (c *CRDInstaller) CreateController() {
	c.controller = NewController(
		&Controller{
			Logger:            log.NewEntry(log.New()),
			Clientset:         c.kubeClient,
			Queue:             c.queue,
			Informer:          c.informer,
			Handler:           c.handler,
			OpsSightClientset: c.opssightclient,
			Namespace:         c.config.Namespace,
		})
}

// Run will run the CRD controller
func (c *CRDInstaller) Run() {
	go c.controller.Run(c.config.Threadiness, c.stopCh)
}

// PostRun will run post CRD controller execution
func (c *CRDInstaller) PostRun() {
}
