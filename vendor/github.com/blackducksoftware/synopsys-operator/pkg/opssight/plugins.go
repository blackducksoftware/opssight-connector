/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownershia. The ASF licenses this file
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

// This is a controller that deletes the hub based on the delete threshold

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/blackducksoftware/horizon/pkg/api"
	hubv2 "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	opssightv1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1" //extensions "github.com/kubernetes/kubernetes/pkg/apis/extensions"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	hubclient "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

var logger *log.Entry

func init() {
	logger = log.WithField("subsystem", "opssight-plugins")
}

// DeleteHub ...
type DeleteHub struct {
	Config         *protoform.Config
	KubeClient     *kubernetes.Clientset
	OpsSightClient *opssightclientset.Clientset
	HubClient      *hubclient.Clientset
	OpsSightSpec   *v1.OpsSightSpec
}

// Run is a BLOCKING function which should be run by the framework .
func (d *DeleteHub) Run(resources api.ControllerResources, ch chan struct{}) error {
	hubCounts, err := d.getHubsCount()
	if err != nil {
		return errors.Trace(err)
	}
	// whether the max no of hub is reached?
	if d.OpsSightSpec.Blackduck.MaxCount == hubCounts {

	}

	return nil
}

func (d *DeleteHub) getHubsCount() (int, error) {
	hubs, err := util.ListHubs(d.HubClient, d.Config.Namespace)
	if err != nil {
		return 0, errors.Annotate(err, "unable to list hubs")
	}
	return len(hubs.Items), nil
}

// This is a controller that updates the configmap
// in perceptor periodically.
// It is assumed that the configmap in perceptor will
// roll over any time this is updated, and if not, that
// there is a problem in the orchestration environment.

// ConfigMapUpdater ...
type ConfigMapUpdater struct {
	config         *protoform.Config
	httpClient     *http.Client
	kubeClient     *kubernetes.Clientset
	hubClient      *hubclient.Clientset
	opssightClient *opssightclientset.Clientset
}

// NewConfigMapUpdater ...
func NewConfigMapUpdater(config *protoform.Config, kubeClient *kubernetes.Clientset, hubClient *hubclient.Clientset, opssightClient *opssightclientset.Clientset) *ConfigMapUpdater {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 5 * time.Second,
	}
	return &ConfigMapUpdater{
		config:         config,
		httpClient:     httpClient,
		kubeClient:     kubeClient,
		hubClient:      hubClient,
		opssightClient: opssightClient,
	}
}

func getKeys(dict map[string]string) []string {
	keys := make([]string, len(dict))
	i := 0
	for k := range dict {
		keys[i] = k
		i++
	}
	return keys
}

// sendHubs is one possible way to configure the perceptor hub family.
func sendHubs(kubeClient *kubernetes.Clientset, opsSightSpec *opssightv1.OpsSightSpec, hubs []string) error {
	configMapName := opsSightSpec.ConfigMapName
	logger.WithField("configMap", opsSightSpec.ConfigMapName).Info("send hubs: looking for config map")
	configMap, err := kubeClient.CoreV1().ConfigMaps(opsSightSpec.Namespace).Get(configMapName, metav1.GetOptions{})

	if err != nil {
		return errors.Annotatef(err, "unable to get configmap %s in %s", configMapName, opsSightSpec.Namespace)
	}

	cmKey := fmt.Sprintf("%s.json", configMapName)
	logger.WithField("lookingForKey", cmKey).WithField("foundKeys", getKeys(configMap.Data)).Infof("send hubs")

	var value MainOpssightConfigMap
	data := configMap.Data[cmKey]
	logger.Debugf("found config map data: %s", data)
	err = json.Unmarshal([]byte(data), &value)
	if err != nil {
		return errors.Trace(err)
	}

	value.Hub.Hosts = util.UniqueValues(append(hubs, value.Hub.Hosts...))

	jsonBytes, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return errors.Trace(err)
	}

	configMap.Data[cmKey] = string(jsonBytes)
	logger.WithFields(log.Fields{"configMap": configMapName, "namespace": opsSightSpec.Namespace}).Debugf("updated configmap to %+v", configMap)
	_, err = kubeClient.CoreV1().ConfigMaps(opsSightSpec.Namespace).Update(configMap)
	if err != nil {
		return errors.Annotatef(err, "unable to update configmap %s in %s", configMapName, opsSightSpec.Namespace)
	}
	return nil
}

// Run ...
func (p *ConfigMapUpdater) Run(ch <-chan struct{}) {
	logger.Infof("Starting controller for hub<->perceptor updates... this blocks, so running in a go func.")

	syncFunc := func() {
		err := p.updateAllHubs()
		if err != nil {
			logger.Errorf("unable to update hubs because %+v", err)
		}
	}

	syncFunc()

	hubListWatch := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return p.hubClient.SynopsysV1().Blackducks(p.config.Namespace).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return p.hubClient.SynopsysV1().Blackducks(p.config.Namespace).Watch(options)
		},
	}
	_, hubController := cache.NewInformer(hubListWatch,
		&hubv2.Blackduck{},
		2*time.Second,
		cache.ResourceEventHandlerFuncs{
			// TODO kinda dumb, we just do a complete re-list of all hubs,
			// every time an event happens... But thats all we need to do, so its good enough.
			DeleteFunc: func(obj interface{}) {
				logger.Debugf("configmap updater hub deleted event ! %v ", obj)
				syncFunc()
			},

			AddFunc: func(obj interface{}) {
				logger.Debugf("configmap updater hub added event! %v ", obj)
				syncFunc()
			},
		},
	)

	opssightListWatch := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return p.opssightClient.SynopsysV1().OpsSights(p.config.Namespace).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return p.opssightClient.SynopsysV1().OpsSights(p.config.Namespace).Watch(options)
		},
	}
	_, opssightController := cache.NewInformer(opssightListWatch,
		&opssightv1.OpsSight{},
		2*time.Second,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				logger.Debugf("configmap updater opssight added event! %v ", obj)
				err := p.updateOpsSight(obj)
				if err != nil {
					logger.Errorf("unable to update opssight because %+v", err)
				}
			},
		},
	)

	// make sure this is called from a go func -- it blocks!
	go hubController.Run(ch)
	go opssightController.Run(ch)
}

func (p *ConfigMapUpdater) getAllHubs(hubType string) []string {
	allHubNamespaces := []string{}
	hubsList, _ := util.ListHubs(p.hubClient, p.config.Namespace)
	hubs := hubsList.Items
	for _, hub := range hubs {
		if strings.EqualFold(hub.Spec.Type, hubType) {
			hubURL := fmt.Sprintf("webserver.%s.svc", hub.Name)
			allHubNamespaces = append(allHubNamespaces, hubURL)
			logger.Infof("Blackduck config map controller, namespace is %s", hub.Name)
		}
	}

	logger.Debugf("allHubNamespaces: %+v", allHubNamespaces)
	return allHubNamespaces
}

// updateAllHubs will list all hubs in the cluster, and send them to opssight as scan targets.
// TODO there may be hubs which we dont want opssight to use.  Not sure how to deal with that yet.
func (p *ConfigMapUpdater) updateAllHubs() []error {
	// for opssight 3.0, only support one opssight
	opssights, err := util.GetOpsSights(p.opssightClient)
	if err != nil {
		return []error{errors.Annotate(err, "unable to get opssights")}
	}

	if len(opssights.Items) == 0 {
		return nil
	}
	// if len(opssights.Items) > 1 {
	// 	return errors.Errorf("cowardly refusing to update OpsSights: found %d", len(opssights.Items))
	// }

	errList := []error{}
	for _, o := range opssights.Items {
		hubType := o.Spec.Blackduck.BlackduckSpec.Type
		allHubNamespaces := p.getAllHubs(hubType)

		for _, o := range opssights.Items {
			err = sendHubs(p.kubeClient, &o.Spec, allHubNamespaces)
			if err != nil {
				errList = append(errList, errors.Annotate(err, "unable to send hubs"))
			}
		}
	}
	return errList
}

// updateOpsSight will list all hubs in the cluster, and send them to opssight as scan targets.
// TODO there may be hubs which we dont want opssight to use.  Not sure how to deal with that yet.
func (p *ConfigMapUpdater) updateOpsSight(obj interface{}) error {
	opssight, ok := obj.(*opssightv1.OpsSight)
	if !ok {
		return errors.Errorf("unable to cast object")
	}
	var err error
	for j := 0; j < 20; j++ {
		opssight, err = util.GetOpsSight(p.opssightClient, p.config.Namespace, opssight.Name)
		if err != nil {
			return fmt.Errorf("unable to get opssight %s due to %+v", opssight.Name, err)
		}

		if strings.EqualFold(opssight.Status.State, "running") {
			break
		}
		logger.Debugf("waiting for opssight %s to be up.....", opssight.Name)
		time.Sleep(10 * time.Second)
	}

	hubType := opssight.Spec.Blackduck.BlackduckSpec.Type
	allHubNamespaces := p.getAllHubs(hubType)

	err = sendHubs(p.kubeClient, &opssight.Spec, allHubNamespaces)
	if err != nil {
		return errors.Annotate(err, "unable to send hubs")
	}

	return nil
}
