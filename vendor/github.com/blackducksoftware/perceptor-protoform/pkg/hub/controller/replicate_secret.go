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

package controller

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	// hub_v1 "github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"

	hubclientset "github.com/blackducksoftware/perceptor-protoform/pkg/hub/client/clientset/versioned"
	"github.com/blackducksoftware/perceptor-protoform/pkg/util"

	log "github.com/sirupsen/logrus"
)

// SecretReplicator will have the configuration related to replicate the secrets
type SecretReplicator struct {
	client        *kubernetes.Clientset
	hubClient     *hubclientset.Clientset
	namespace     string
	controller    cache.Controller
	dependencyMap map[string][]string
}

// NewSecretReplicator creates a new secret replicator
func NewSecretReplicator(client *kubernetes.Clientset, hubClient *hubclientset.Clientset, namespace string, resyncPeriod time.Duration) *SecretReplicator {
	dependencyMap, _ := buildDependentKeys(hubClient, namespace)

	repl := SecretReplicator{
		client:        client,
		hubClient:     hubClient,
		namespace:     namespace,
		dependencyMap: dependencyMap,
	}

	_, controller := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().Secrets(v1.NamespaceAll).List(opts)
			},
			WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
				return client.CoreV1().Secrets(v1.NamespaceAll).Watch(opts)
			},
		},
		&v1.Secret{},
		resyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    repl.secretAdded,
			UpdateFunc: func(old interface{}, new interface{}) { repl.secretUpdated(old, new) },
		},
	)

	repl.controller = controller

	return &repl
}

// Run method will watch for secrets events
func (r *SecretReplicator) Run(stopCh <-chan struct{}) {
	log.Printf("running secret controller")
	go r.controller.Run(stopCh)
	// Wait until we're told to stop
	<-stopCh
}

func (r *SecretReplicator) secretAdded(obj interface{}) {
	secret := obj.(*v1.Secret)

	if strings.EqualFold(secret.Name, "blackduck-certificate") {
		hubList, err := util.ListHubs(r.hubClient, r.namespace)
		if err != nil {
			log.Errorf("unable to list the hubs due to %+v", err)
		}

		for _, hub := range hubList.Items {
			if strings.EqualFold(secret.Namespace, hub.Name) {
				if !strings.EqualFold(hub.Spec.CertificateName, "manual") {
					replicas := r.dependencyMap[hub.Spec.CertificateName]
					replicas = append(replicas, secret.Namespace)
					r.updateSecretData(secret, hub.Spec.CertificateName)
					r.dependencyMap[hub.Spec.CertificateName] = replicas
				} else {
					return
				}
			}
		}
	} else {
		return
	}
}

func (r *SecretReplicator) secretUpdated(oldObj interface{}, newobj interface{}) {
	secret := newobj.(*v1.Secret)

	if strings.EqualFold(secret.Name, "blackduck-certificate") {
		hubList, err := util.ListHubs(r.hubClient, r.namespace)
		if err != nil {
			log.Errorf("unable to list the hubs due to %+v", err)
		}

		for _, hub := range hubList.Items {
			if strings.EqualFold(secret.Namespace, hub.Name) {
				r.updateDependents(secret, r.dependencyMap[secret.Namespace])
			}
		}
	}
}

func (r *SecretReplicator) replicateSecret(sourceSecret *v1.Secret, updatedSecret *v1.Secret) error {

	if reflect.DeepEqual(sourceSecret.Data, updatedSecret.Data) {
		log.Infof("secret %s/%s is already up-to-date", sourceSecret.Namespace, sourceSecret.Name)
		return nil
	}

	secretCopy := sourceSecret.DeepCopy()

	if secretCopy.Data == nil {
		secretCopy.Data = make(map[string][]byte)
	}

	for key, value := range updatedSecret.Data {
		newValue := make([]byte, len(value))
		copy(newValue, value)
		secretCopy.Data[key] = newValue
	}

	log.Debugf("updating secret %s/%s", sourceSecret.Namespace, sourceSecret.Name)

	_, err := r.client.CoreV1().Secrets(sourceSecret.Namespace).Update(secretCopy)
	if err != nil {
		return err
	}

	log.Printf("updated secret %s/%s", sourceSecret.Namespace, sourceSecret.Name)

	return nil
}

func (r *SecretReplicator) deletePod(namespace string) error {
	log.Debugf("deleting pod %s/%s", namespace, "webserver")

	// Get all pods corresponding to the hub namespace
	pods, err := util.GetAllPodsForNamespace(r.client, namespace)
	if err != nil {
		return fmt.Errorf("unable to list the pods in namespace %s due to %+v", namespace, err)
	}

	webserverPod := util.FilterPodByNamePrefix(pods, "webserver")
	err = r.client.AppsV1().Deployments(namespace).Delete(webserverPod.Name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	log.Printf("deleted pod %s/%s", namespace, "webserver")
	return nil
}

func (r *SecretReplicator) updateSecretData(secret *v1.Secret, dependentSecretNamespace string) error {
	log.Printf("updating dependent secret %s/%s -> %s", secret.Namespace, secret.Name, dependentSecretNamespace)
	var err error
	var dependentSecret *v1.Secret
	if strings.EqualFold(dependentSecretNamespace, "default") {
		dependentSecret, err = util.GetSecret(r.client, r.namespace, secret.Name)
	} else {
		dependentSecret, err = util.GetSecret(r.client, dependentSecretNamespace, secret.Name)
	}
	if err != nil {
		log.Errorf("could not get dependent secret %s: %s", dependentSecretNamespace, err)
	}

	r.replicateSecret(secret, dependentSecret)

	return nil
}

func (r *SecretReplicator) updateDependents(secret *v1.Secret, dependents []string) error {
	for _, dependentKey := range dependents {
		log.Printf("updating dependent secret %s/%s -> %s", secret.Namespace, secret.Name, dependentKey)

		sourceSecret, err := util.GetSecret(r.client, dependentKey, secret.Name)
		if err != nil {
			log.Errorf("could not get dependent secret %s: %s", dependentKey, err)
		}

		r.replicateSecret(sourceSecret, secret)
		r.deletePod(secret.Namespace)
	}

	return nil
}

func buildDependentKeys(hubClient *hubclientset.Clientset, namespace string) (map[string][]string, error) {
	hubList, err := util.ListHubs(hubClient, namespace)
	if err != nil {
		log.Errorf("unable to list the hubs due to %+v", err)
		return nil, err
	}

	dependencyMap := make(map[string][]string)
	for _, hub := range hubList.Items {
		replicas := dependencyMap[hub.Spec.CertificateName]
		replicas = append(replicas, hub.Name)
		dependencyMap[hub.Spec.CertificateName] = replicas
	}
	log.Debugf("dependencyMap: %+v", dependencyMap)
	return dependencyMap, nil
}
