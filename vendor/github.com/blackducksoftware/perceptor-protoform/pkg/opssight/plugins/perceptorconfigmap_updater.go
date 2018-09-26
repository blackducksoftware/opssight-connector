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

package plugins

// This is a controller that updates the configmap
// in perceptor periodically.
// It is assumed that the configmap in perceptor will
// roll over any time this is updated, and if not, that
// there is a problem in the orchestration environment.

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/blackducksoftware/horizon/pkg/api"
	hubv1 "github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
	hubclient "github.com/blackducksoftware/perceptor-protoform/pkg/hub/client/clientset/versioned"
	"github.com/blackducksoftware/perceptor-protoform/pkg/model"
	opssightclientset "github.com/blackducksoftware/perceptor-protoform/pkg/opssight/client/clientset/versioned"
	"github.com/blackducksoftware/perceptor-protoform/pkg/util"
	v1 "k8s.io/api/core/v1"

	//extensions "github.com/kubernetes/kubernetes/pkg/apis/extensions"

	log "github.com/sirupsen/logrus"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type hubConfig struct {
	Hosts                     []string
	User                      string
	PasswordEnvVar            string
	ClientTimeoutMilliseconds *int
	Port                      *int
	ConcurrentScanLimit       *int
	TotalScanLimit            *int
}

type timings struct {
	CheckForStalledScansPauseHours *int
	StalledScanClientTimeoutHours  *int
	ModelMetricsPauseSeconds       *int
	UnknownImagePauseMilliseconds  *int
}

type perceptorConfig struct {
	Hub         *hubConfig
	Timings     *timings
	UseMockMode bool
	Port        *int
	LogLevel    string
}

// PerceptorConfigMap ...
type PerceptorConfigMap struct {
	Config         *model.Config
	KubeConfig     *rest.Config
	OpsSightClient *opssightclientset.Clientset
	Namespace      string
}

// sendHubs is one possible way to configure the perceptor hub family.
// TODO replace w/ configmap mutation if we want to.
func sendHubs(kubeClient *kubernetes.Clientset, namespace string, hubs []string) error {
	configmapList, err := kubeClient.CoreV1().ConfigMaps(namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	var configMap *v1.ConfigMap
	for _, cm := range configmapList.Items {
		if cm.Name == "perceptor" {
			configMap = &cm
			break
		}
	}

	if configMap == nil {
		return fmt.Errorf("unable to find configmap perceptor in %s", namespace)
	}

	var value perceptorConfig
	err = json.Unmarshal([]byte(configMap.Data["perceptor.yaml"]), &value)
	if err != nil {
		return err
	}

	value.Hub.Hosts = hubs

	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	configMap.Data["perceptor.yaml"] = string(jsonBytes)
	log.Debugf("updated configmap in %s is %+v", namespace, configMap)
	_, err = kubeClient.CoreV1().ConfigMaps(namespace).Update(configMap)
	if err != nil {
		return err
	}
	return nil
}

// Run is a BLOCKING function which should be run by the framework .
func (p *PerceptorConfigMap) Run(resources api.ControllerResources, ch chan struct{}) error {
	hubClient, err := hubclient.NewForConfig(p.KubeConfig)
	if err != nil {
		log.Panicf("unable to create the hub client due to %+v", err)
	}

	syncFunc := func() {
		p.updateAllHubs(hubClient, resources.KubeClient)
	}

	syncFunc()

	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return hubClient.SynopsysV1().Hubs(p.Config.Namespace).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return hubClient.SynopsysV1().Hubs(p.Config.Namespace).Watch(options)
		},
	}
	_, ctrl := cache.NewInformer(lw,
		&hubv1.Hub{},
		2*time.Second,
		cache.ResourceEventHandlerFuncs{
			// TODO kinda dumb, we just do a complete re-list of all hubs,
			// every time an event happens... But thats all we need to do, so its good enough.
			DeleteFunc: func(obj interface{}) {
				log.Debugf("Hub deleted ! %v ", obj)
				syncFunc()
			},

			AddFunc: func(obj interface{}) {
				log.Debugf("Hub added ! %v ", obj)
				syncFunc()
			},
		},
	)
	log.Infof("Starting controller for hub<->perceptor updates... this blocks, so running in a go func.")

	// make sure this is called from a go func.
	// This blocks!
	go ctrl.Run(ch)

	return nil
}

// updateAllHubs will list all hubs in the cluster, and send them to opssight as scan targets.
// TODO there may be hubs which we dont want opssight to use.  Not sure how to deal with that yet.
func (p *PerceptorConfigMap) updateAllHubs(hubClient *hubclient.Clientset, kubeClient *kubernetes.Clientset) error {
	allHubNamespaces := func() []string {
		allHubNamespaces := []string{}

		hubsList, _ := util.ListHubs(hubClient, p.Config.Namespace)
		hubs := hubsList.Items
		for _, hub := range hubs {
			if strings.EqualFold(hub.Spec.HubType, "worker") {
				hubURL := fmt.Sprintf("webserver.%s.svc", hub.Name)
				status := p.verifyHub(hubClient, hubURL, hub.Name)
				if status {
					allHubNamespaces = append(allHubNamespaces, hubURL)
				}
				log.Infof("Hub config map controller, namespace is %s", hub.Name)
			}
		}
		return allHubNamespaces
	}()

	log.Debugf("allHubNamespaces: %+v", allHubNamespaces)
	// for opssight 3.0, only support one opssight
	opssight, err := util.GetOpsSight(p.OpsSightClient, p.Namespace, p.Namespace)
	if err != nil {
		log.Errorf("unable to get opssight in %s due to %+v", p.Namespace, err)
		return err
	}

	// TODO, replace w/ configmap mutat ?
	// curl perceptor w/ the latest hub list
	err = sendHubs(kubeClient, opssight.Name, allHubNamespaces)
	if err != nil {
		log.Errorf("unable to send hubs due to %+v", err)
		return err
	}

	return nil
}

func (p *PerceptorConfigMap) verifyHub(hubClient *hubclient.Clientset, hubURL string, name string) bool {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 5 * time.Second,
	}

	for i := 0; i < 60; i++ {
		resp, err := client.Get(fmt.Sprintf("https://%s:443/api/current-version", hubURL))
		if err != nil {
			log.Debugf("unable to talk with the hub %s", hubURL)
			time.Sleep(10 * time.Second)
			_, err := util.GetHub(hubClient, p.Config.Namespace, name)
			if err != nil {
				return false
			}
			continue
		}

		_, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("unable to read the response from hub %s due to %+v", hubURL, err)
			return false
		}
		defer resp.Body.Close()
		log.Debugf("hub response status for %s is %v", hubURL, resp.Status)

		if resp.StatusCode == 200 {
			return true
		}
		time.Sleep(10 * time.Second)
	}
	return false
}
