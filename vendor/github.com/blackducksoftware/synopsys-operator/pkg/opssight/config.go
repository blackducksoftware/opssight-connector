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
	"time"

	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// Config defines the specification for the controller
type Config struct {
	Config          *protoform.Config
	KubeConfig      *rest.Config
	KubeClientSet   *kubernetes.Clientset
	Defaults        interface{}
	resyncPeriod    time.Duration
	indexers        cache.Indexers
	informer        cache.SharedIndexInformer
	queue           workqueue.RateLimitingInterface
	handler         *Handler
	controller      *Controller
	customClientSet *opssightclientset.Clientset
	Threadiness     int
	StopCh          <-chan struct{}
}
