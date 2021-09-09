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

package protoform

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth" //for auths
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// GetKubeConfig  will return the kube config
func GetKubeConfig(kubeconfigpath string, insecureSkipTLSVerify bool) (*rest.Config, error) {
	log.Debugf("Getting Kube Rest Config")
	var err error
	var kubeConfig *rest.Config
	// creates the in-cluster config
	kubeConfig, err = rest.InClusterConfig()
	if err != nil {
		log.Debugf("using native config due to %+v", err)
		kubeConfig, err = GetKubeClientFromOutsideCluster(kubeconfigpath, insecureSkipTLSVerify)
		if err != nil {
			return nil, err
		}
	}

	return kubeConfig, nil
}

// homeDir determines the user's home directory path
func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

// GetKubeClientSet will return the kube clientset
func GetKubeClientSet(kubeConfig *rest.Config) (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(kubeConfig)
}

// GetKubeClientFromOutsideCluster returns the rest config of outside cluster
func GetKubeClientFromOutsideCluster(kubeconfigpath string, insecureSkipTLSVerify bool) (*rest.Config, error) {
	// Determine Config Paths
	if home := homeDir(); len(kubeconfigpath) == 0 && home != "" {
		kubeconfigpath = filepath.Join(home, ".kube", "config")
	}

	kubeConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{
			ExplicitPath: kubeconfigpath,
		},
		&clientcmd.ConfigOverrides{
			ClusterInfo: clientcmdapi.Cluster{
				Server:                "",
				InsecureSkipTLSVerify: insecureSkipTLSVerify,
			},
		}).ClientConfig()
	if err != nil {
		return nil, err
	}
	return kubeConfig, nil
}
