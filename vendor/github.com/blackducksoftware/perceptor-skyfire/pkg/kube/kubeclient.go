/*
Copyright (C) 2018 Black Duck Software, Inc.

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

package kube

import (
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ClientInterface .....
type ClientInterface interface {
	Dump() (*Dump, error)
}

// KubeClient is an implementation of the Client interface for kubernetes
type KubeClient struct {
	clientset kubernetes.Clientset
}

// NewKubeClient .....
func NewKubeClient(config *KubeClientConfig) (*KubeClient, error) {
	if config != nil {
		return NewKubeClientFromOutsideCluster(config.MasterURL, config.KubeConfigPath)
	} else {
		return NewKubeClientFromInsideCluster()
	}
}

// NewKubeClientFromInsideCluster instantiates a KubeClient using configuration
// pulled from the cluster.
func NewKubeClientFromInsideCluster() (*KubeClient, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Annotate(err, "unable to build config from cluster")
	}
	return newKubeClientHelper(config)
}

// NewKubeClientFromOutsideCluster instantiates a KubeClient using a master URL and
// a path to a kubeconfig file.
func NewKubeClientFromOutsideCluster(masterURL string, kubeconfigPath string) (*KubeClient, error) {
	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		return nil, errors.Annotate(err, "unable to build config from flags")
	}

	return newKubeClientHelper(config)
}

func newKubeClientHelper(config *rest.Config) (*KubeClient, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Annotate(err, "unable to create kubernetes clientset")
	}

	return &KubeClient{clientset: *clientset}, nil
}

// Dump .....
func (client *KubeClient) Dump() (*Dump, error) {
	kubePods, err := client.GetAllPods()
	if err != nil {
		return nil, errors.Trace(err)
	}
	kubeMeta, err := client.GetMeta()
	if err != nil {
		return nil, errors.Trace(err)
	}
	return NewDump(kubeMeta, kubePods), nil
}

// GetAllPods .....
func (client *KubeClient) GetAllPods() ([]*Pod, error) {
	pods := []*Pod{}
	kubePods, err := client.clientset.CoreV1().Pods(v1.NamespaceAll).List(meta_v1.ListOptions{})
	if err != nil {
		return nil, errors.Trace(err)
	}
	for _, kubePod := range kubePods.Items {
		pods = append(pods, mapKubePod(&kubePod))
	}
	return pods, nil
}

/* DumpServices .....
func (client *KubeClient) GetAnnotations() {
	pods, x := client.GetAllPods()
	for _, pod := range pods {
		for k,v :=  range pod.BDAnnotations {
			log.Infof("annotation !!!  %v %v",k,v)
		}
	}
}*/

// DumpServices .....
func (client *KubeClient) DumpServices() (*ServiceDump, error) {
	// Get a Slice of Service items for all services
	kubeServices, err := client.GetAllServices()
	if err != nil {
		return nil, errors.Trace(err)
	}
	// Get meta data about the cluster
	kubeMeta, err := client.GetMeta()
	if err != nil {
		return nil, errors.Trace(err)
	}
	// Return the Services in a dump format
	return NewServiceDump(kubeMeta, kubeServices), nil
}

// GetAllServices .....
func (client *KubeClient) GetAllServices() ([]*Service, error) {
	// Empty Slice to store Service type items
	services := []*Service{}
	// Get a list of services from the KubeClient
	kubeServices, err := client.clientset.CoreV1().Services(v1.NamespaceAll).List(meta_v1.ListOptions{})
	if err != nil {
		return nil, errors.Trace(err)
	}
	// Append Service items to the Slice
	for _, kubeService := range kubeServices.Items {
		var ports []int32
		for _, port := range kubeService.Spec.Ports {
			ports = append(ports, port.Port)
		}
		services = append(services, NewService(kubeService.Name, kubeService.Namespace, ports))
	}
	return services, nil
}

// GetMeta .....
func (client *KubeClient) GetMeta() (*Meta, error) {
	nodeList, err := client.clientset.CoreV1().Nodes().List(meta_v1.ListOptions{})
	if err != nil {
		return nil, errors.Trace(err)
	}
	version, err := client.clientset.ServerVersion()
	if err != nil {
		return nil, errors.Trace(err)
	}
	meta := &Meta{
		BuildDate:    version.BuildDate,
		Compiler:     version.Compiler,
		GitCommit:    version.GitCommit,
		GitTreeState: version.GitTreeState,
		GitVersion:   version.GitVersion,
		GoVersion:    version.GoVersion,
		MajorVersion: version.Major,
		MinorVersion: version.Minor,
		Platform:     version.Platform,
		NodeCount:    len(nodeList.Items),
	}
	log.Infof("kube meta: %+v, from %+v", version, meta)
	return meta, nil
}
