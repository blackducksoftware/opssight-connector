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
	"encoding/json"
	"fmt"
	"time"

	"github.com/blackducksoftware/horizon/pkg/components"
	hubclientset "github.com/blackducksoftware/perceptor-protoform/pkg/hub/client/clientset/versioned"
	hubinformerv1 "github.com/blackducksoftware/perceptor-protoform/pkg/hub/client/informers/externalversions/hub/v1"
	hubcontroller "github.com/blackducksoftware/perceptor-protoform/pkg/hub/controller"
	"github.com/blackducksoftware/perceptor-protoform/pkg/util"
	"github.com/juju/errors"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	//_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	horizon "github.com/blackducksoftware/horizon/pkg/deployer"

	"github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
	"github.com/blackducksoftware/perceptor-protoform/pkg/hub"
	"github.com/blackducksoftware/perceptor-protoform/pkg/hub/webservice"

	log "github.com/sirupsen/logrus"
)

// Controller defines the specification for the controller
type Controller struct {
	protoform *ProtoformConfig
}

// NewController will create a controller configuration
func NewController(config interface{}) (*Controller, error) {
	dependentConfig, ok := config.(*ProtoformConfig)
	if !ok {
		return nil, fmt.Errorf("failed to convert hub defaults: %v", config)
	}
	c := &Controller{protoform: dependentConfig}

	c.protoform.resyncPeriod = 0
	c.protoform.indexers = cache.Indexers{}

	return c, nil
}

// CreateClientSet will create the CRD client
func (c *Controller) CreateClientSet() error {
	hubClient, err := hubclientset.NewForConfig(c.protoform.KubeConfig)
	if err != nil {
		return errors.Trace(err)
	}
	c.protoform.customClientSet = hubClient
	return nil
}

// Deploy will deploy the CRD and other relevant components
func (c *Controller) Deploy() error {
	deployer, err := horizon.NewDeployer(c.protoform.KubeConfig)
	if err != nil {
		return err
	}

	// Hub CRD
	deployer.AddCustomDefinedResource(components.NewCustomResourceDefintion(horizonapi.CRDConfig{
		APIVersion: "apiextensions.k8s.io/v1beta1",
		Name:       "hubs.synopsys.com",
		Namespace:  c.protoform.Config.Namespace,
		Group:      "synopsys.com",
		CRDVersion: "v1",
		Kind:       "Hub",
		Plural:     "hubs",
		Singular:   "hub",
		Scope:      horizonapi.CRDClusterScoped,
	}))

	// Perceptor configMap
	hubFederatorConfig := components.NewConfigMap(horizonapi.ConfigMapConfig{Namespace: c.protoform.Config.Namespace, Name: "federator"})
	data := map[string]interface{}{
		"HubConfig": map[string]interface{}{
			"Port":                         c.protoform.Config.HubFederatorConfig.HubConfig.Port,
			"User":                         c.protoform.Config.HubFederatorConfig.HubConfig.User,
			"PasswordEnvVar":               c.protoform.Config.HubFederatorConfig.HubConfig.PasswordEnvVar,
			"ClientTimeoutMilliseconds":    c.protoform.Config.HubFederatorConfig.HubConfig.ClientTimeoutMilliseconds,
			"FetchAllProjectsPauseSeconds": c.protoform.Config.HubFederatorConfig.HubConfig.FetchAllProjectsPauseSeconds,
		},
		"Port":        c.protoform.Config.HubFederatorConfig.Port,
		"LogLevel":    c.protoform.Config.LogLevel,
		"UseMockMode": c.protoform.Config.HubFederatorConfig.UseMockMode,
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return errors.Trace(err)
	}
	hubFederatorConfig.AddData(map[string]string{"config.json": string(bytes)})
	deployer.AddConfigMap(hubFederatorConfig)

	// Perceptor service
	deployer.AddService(util.CreateService("federator", "federator", c.protoform.Config.Namespace, fmt.Sprint(c.protoform.Config.HubFederatorConfig.Port), fmt.Sprint(c.protoform.Config.HubFederatorConfig.Port), horizonapi.ClusterIPServiceTypeDefault))
	deployer.AddService(util.CreateService("federator-np", "federator", c.protoform.Config.Namespace, fmt.Sprint(c.protoform.Config.HubFederatorConfig.Port), fmt.Sprint(c.protoform.Config.HubFederatorConfig.Port), horizonapi.ClusterIPServiceTypeNodePort))
	deployer.AddService(util.CreateService("federator-lb", "federator", c.protoform.Config.Namespace, fmt.Sprint(c.protoform.Config.HubFederatorConfig.Port), fmt.Sprint(c.protoform.Config.HubFederatorConfig.Port), horizonapi.ClusterIPServiceTypeLoadBalancer))

	var hubPassword string
	for {
		blackduckSecret, err := util.GetSecret(c.protoform.KubeClientSet, c.protoform.Config.Namespace, "blackduck-secret")
		if err != nil {
			log.Infof("Aborting: You need to first create a 'blackduck-secret' in this namespace with HUB_PASSWORD and retry")
		} else {
			hubPassword = string(blackduckSecret.Data["HUB_PASSWORD"])
			break
		}
		time.Sleep(5 * time.Second)
	}

	// Hub federator deployment
	hubFederatorContainerConfig := &util.Container{
		ContainerConfig: &horizonapi.ContainerConfig{Name: "federator", Image: fmt.Sprintf("%s/%s/%s:%s", c.protoform.Config.HubFederatorConfig.Registry, c.protoform.Config.HubFederatorConfig.ImagePath, c.protoform.Config.HubFederatorConfig.ImageName, c.protoform.Config.HubFederatorConfig.ImageVersion),
			PullPolicy: horizonapi.PullAlways, Command: []string{"./federator"}, Args: []string{"/etc/federator/config.json"}},
		EnvConfigs:   []*horizonapi.EnvConfig{{Type: horizonapi.EnvVal, NameOrPrefix: c.protoform.Config.HubFederatorConfig.HubConfig.PasswordEnvVar, KeyOrVal: hubPassword}},
		VolumeMounts: []*horizonapi.VolumeMountConfig{{Name: "federator", MountPath: "/etc/federator", Propagation: horizonapi.MountPropagationNone}},
		PortConfig:   &horizonapi.PortConfig{ContainerPort: fmt.Sprint(c.protoform.Config.HubFederatorConfig.Port), Protocol: horizonapi.ProtocolTCP},
	}
	hubFederatorVolume := components.NewConfigMapVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "federator",
		MapOrSecretName: "federator",
		DefaultMode:     util.IntToInt32(420),
	})
	hubFederator := util.CreateDeploymentFromContainer(&horizonapi.DeploymentConfig{Namespace: c.protoform.Config.Namespace, Name: "federator", Replicas: util.IntToInt32(1)},
		[]*util.Container{hubFederatorContainerConfig}, []*components.Volume{hubFederatorVolume}, []*util.Container{}, []horizonapi.AffinityConfig{})
	deployer.AddDeployment(hubFederator)

	certificate, key := hub.CreateSelfSignedCert()

	certificateSecret := components.NewSecret(horizonapi.SecretConfig{Namespace: c.protoform.Config.Namespace, Name: "blackduck-certificate", Type: horizonapi.SecretTypeOpaque})
	certificateSecret.AddData(map[string][]byte{"WEBSERVER_CUSTOM_CERT_FILE": []byte(certificate), "WEBSERVER_CUSTOM_KEY_FILE": []byte(key)})

	deployer.AddSecret(certificateSecret)

	err = deployer.Run()
	if err != nil {
		log.Errorf("unable to create the hub federator resources due to %+v", err)
	}

	time.Sleep(5 * time.Second)

	return err
}

// PostDeploy will call after deploying the CRD
func (c *Controller) PostDeploy() {
	hc := hub.NewCreater(c.protoform.Config, c.protoform.KubeConfig, c.protoform.KubeClientSet, c.protoform.customClientSet)
	webservice.SetupHTTPServer(hc, c.protoform.Config.Namespace)
}

// CreateInformer will create a informer for the CRD
func (c *Controller) CreateInformer() {
	c.protoform.infomer = hubinformerv1.NewHubInformer(
		c.protoform.customClientSet,
		c.protoform.Config.Namespace,
		c.protoform.resyncPeriod,
		c.protoform.indexers,
	)
}

// CreateQueue will create a queue to process the CRD
func (c *Controller) CreateQueue() {
	// create a new queue so that when the informer gets a resource that is either
	// a result of listing or watching, we can add an idenfitying key to the queue
	// so that it can be handled in the handler
	c.protoform.queue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
}

// AddInformerEventHandler will add the event handlers for the informers
func (c *Controller) AddInformerEventHandler() {
	c.protoform.infomer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// convert the resource object into a key (in this case
			// we are just doing it in the format of 'namespace/name')
			key, err := cache.MetaNamespaceKeyFunc(obj)
			log.Infof("add hub: %s", key)
			if err == nil {
				// add the key to the queue for the handler to get
				c.protoform.queue.Add(key)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			log.Infof("update hub: %s", key)
			if err == nil {
				c.protoform.queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// DeletionHandlingMetaNamsespaceKeyFunc is a helper function that allows
			// us to check the DeletedFinalStateUnknown existence in the event that
			// a resource was deleted but it is still contained in the index
			//
			// this then in turn calls MetaNamespaceKeyFunc
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			log.Infof("delete hub: %s: %+v", key, obj)

			if err == nil {
				c.protoform.queue.Add(key)
			}
		},
	})
}

// CreateHandler will create a CRD handler
func (c *Controller) CreateHandler() {
	c.protoform.handler = &hubcontroller.HubHandler{
		Config:           c.protoform.Config,
		KubeConfig:       c.protoform.KubeConfig,
		Clientset:        c.protoform.KubeClientSet,
		HubClientset:     c.protoform.customClientSet,
		Namespace:        c.protoform.Config.Namespace,
		FederatorBaseURL: fmt.Sprintf("http://federator:%d", c.protoform.Config.HubFederatorConfig.Port),
		CmMutex:          make(chan bool, 1),
		Defaults:         c.protoform.Defaults.(*v1.HubSpec),
	}
}

// CreateController will create a CRD controller
func (c *Controller) CreateController() {
	c.protoform.controller = hubcontroller.NewController(
		&hubcontroller.Controller{
			Logger:   log.NewEntry(log.New()),
			Queue:    c.protoform.queue,
			Informer: c.protoform.infomer,
			Handler:  c.protoform.handler,
		})
}

// Run will run the CRD controller
func (c *Controller) Run() {
	go c.protoform.controller.Run(c.protoform.Threadiness, c.protoform.StopCh)
}

// PostRun will run post CRD controller execution
func (c *Controller) PostRun() {
	secretReplicator := hubcontroller.NewSecretReplicator(c.protoform.KubeClientSet, c.protoform.customClientSet, c.protoform.Config.Namespace, c.protoform.resyncPeriod)
	go secretReplicator.Run(c.protoform.StopCh)
}
