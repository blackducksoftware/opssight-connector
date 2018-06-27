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

package basicskyfire

// import (
// 	"fmt"
//
// 	log "github.com/sirupsen/logrus"
//
// 	"k8s.io/api/core/v1"
// 	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/client-go/kubernetes"
// 	rest "k8s.io/client-go/rest"
//
// 	"k8s.io/client-go/tools/clientcmd"
// )
//
// type KubeClientConfig struct {
// 	MasterURL      string
// 	KubeConfigPath string
// }
//
// // KubeClient is an implementation of the Client interface for kubernetes
// type KubeClient struct {
// 	clientset kubernetes.Clientset
// }
//
// func NewKubeClient(config *KubeClientConfig) (*KubeClient, error) {
// 	if config != nil {
// 		return NewKubeClientFromOutsideCluster(config.MasterURL, config.KubeConfigPath)
// 	} else {
// 		return NewKubeClientFromInsideCluster()
// 	}
// }
//
// // NewKubeClientFromInsideCluster instantiates a KubeClient using configuration
// // pulled from the cluster.
// func NewKubeClientFromInsideCluster() (*KubeClient, error) {
// 	config, err := rest.InClusterConfig()
// 	if err != nil {
// 		log.Errorf("unable to build config from cluster: %s", err.Error())
// 		return nil, err
// 	}
// 	return newKubeClientHelper(config)
// }
//
// // NewKubeClientFromOutsideCluster instantiates a KubeClient using a master URL and
// // a path to a kubeconfig file.
// func NewKubeClientFromOutsideCluster(masterURL string, kubeconfigPath string) (*KubeClient, error) {
// 	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
// 	if err != nil {
// 		log.Errorf("unable to build config from flags: %s", err.Error())
// 		return nil, err
// 	}
//
// 	return newKubeClientHelper(config)
// }
//
// func newKubeClientHelper(config *rest.Config) (*KubeClient, error) {
// 	clientset, err := kubernetes.NewForConfig(config)
// 	if err != nil {
// 		log.Errorf("unable to create kubernetes clientset: %s", err.Error())
// 		return nil, err
// 	}
//
// 	return &KubeClient{clientset: *clientset}, nil
// }
//
// func (client *KubeClient) GetAllPods() ([]*Pod, error) {
// 	pods := []*Pod{}
// 	kubePods, err := client.clientset.CoreV1().Pods(v1.NamespaceAll).List(meta_v1.ListOptions{})
// 	if err != nil {
// 		return nil, err
// 	}
// 	for _, kubePod := range kubePods.Items {
// 		pods = append(pods, mapKubePod(&kubePod))
// 	}
// 	return pods, nil
// }
//
// func (client *KubeClient) GetMeta() (*Meta, error) {
// 	nodeList, err := client.clientset.CoreV1().Nodes().List(meta_v1.ListOptions{})
// 	if err != nil {
// 		return nil, err
// 	}
// 	version, err := client.clientset.ServerVersion()
// 	if err != nil {
// 		return nil, err
// 	}
// 	meta := &Meta{
// 		BuildDate:    version.BuildDate,
// 		Compiler:     version.Compiler,
// 		GitCommit:    version.GitCommit,
// 		GitTreeState: version.GitTreeState,
// 		GitVersion:   version.GitVersion,
// 		GoVersion:    version.GoVersion,
// 		MajorVersion: version.Major,
// 		MinorVersion: version.Minor,
// 		Platform:     version.Platform,
// 		NodeCount:    len(nodeList.Items),
// 	}
// 	log.Infof("kube meta: %+v, from %+v", version, meta)
// 	return meta, nil
// }
//
// type Pod struct {
// 	Name        string
// 	UID         string
// 	Namespace   string
// 	Containers  []*Container
// 	Annotations map[string]string
// 	Labels      map[string]string
// }
//
// func mapKubePod(kubePod *v1.Pod) *Pod {
// 	containers := []*Container{}
// 	for _, newCont := range kubePod.Status.ContainerStatuses {
// 		newImage := &Image{newCont.Image, newCont.ImageID}
// 		addedCont := &Container{newImage, newCont.Name}
// 		containers = append(containers, addedCont)
// 	}
// 	return &Pod{
// 		kubePod.Name,
// 		string(kubePod.UID),
// 		kubePod.Namespace,
// 		containers,
// 		kubePod.Annotations,
// 		kubePod.Labels}
// }
//
// func (pod *Pod) QualifiedName() string {
// 	return fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
// }
//
// type Container struct {
// 	Image *Image
// 	Name  string
// }
//
// type Image struct {
// 	Image   string
// 	ImageID string
// }
//
// type Meta struct {
// 	MajorVersion string
// 	MinorVersion string
// 	GitVersion   string
// 	GitCommit    string
// 	GitTreeState string
// 	BuildDate    string
// 	GoVersion    string
// 	Compiler     string
// 	Platform     string
// 	NodeCount    int
// }
