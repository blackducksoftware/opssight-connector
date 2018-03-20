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

package controller

import (
	"fmt"
	"time"

	"github.com/blackducksoftware/perceivers/pkg/annotations"
	"github.com/blackducksoftware/perceivers/pkg/communicator"
	"github.com/blackducksoftware/perceivers/pod/pkg/mapper"
	"github.com/blackducksoftware/perceivers/pod/pkg/metrics"

	perceptorapi "github.com/blackducksoftware/perceptor/pkg/api"

	"k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"

	"k8s.io/client-go/kubernetes"
	v1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	log "github.com/sirupsen/logrus"
)

// PodController handles watching pods and sending them to perceptor
type PodController struct {
	client        kubernetes.Interface
	podController cache.Controller
	podIndexer    cache.Indexer
	podLister     v1lister.PodLister
	podURL        string

	syncHandler func(string) error
	queue       workqueue.RateLimitingInterface
}

// NewPodController creates a new PodController object
func NewPodController(kubeClient kubernetes.Interface, perceptorURL string) *PodController {
	pc := PodController{
		client: kubeClient,
		queue:  workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Pods"),
		podURL: fmt.Sprintf("%s/%s", perceptorURL, perceptorapi.PodPath),
	}
	pc.podIndexer, pc.podController = cache.NewIndexerInformer(
		&cache.ListWatch{
			ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
				return pc.client.CoreV1().Pods(metav1.NamespaceAll).List(opts)
			},
			WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
				return pc.client.CoreV1().Pods(metav1.NamespaceAll).Watch(opts)
			},
		},
		&v1.Pod{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: pc.enqueueJob,
			UpdateFunc: func(oldObj, newObj interface{}) {
				old, ok1 := oldObj.(*v1.Pod)
				new, ok2 := newObj.(*v1.Pod)
				if ok1 && ok2 && pc.needsUpdate(old, new) {
					pc.enqueueJob(new)
				}
			},
			DeleteFunc: pc.enqueueJob,
		},
		cache.Indexers{},
	)
	pc.podLister = v1lister.NewPodLister(pc.podIndexer)
	pc.syncHandler = pc.processPod

	return &pc
}

// Run starts a controller that watches pods and sends them to perceptor
func (pc *PodController) Run(threadiness int, stopCh <-chan struct{}) {
	log.Infof("starting pod controller")

	defer pc.queue.ShutDown()

	go pc.podController.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, pc.podController.HasSynced) {
		return
	}

	// Start up your worker threads based on threadiness.  Some controllers have multiple kinds of workers
	for i := 0; i < threadiness; i++ {
		// runWorker will loop until "something bad" happens.  The .Until will then rekick the worker
		// after one second
		go wait.Until(pc.runWorker, time.Second, stopCh)
	}

	// Wait until we're told to stop
	<-stopCh
}

func (pc *PodController) enqueueJob(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err == nil {
		pc.queue.Add(key)
	} else {
		metrics.RecordError("controller", "unable to get key")
	}
}

func (pc *PodController) needsUpdate(oldObj *v1.Pod, newObj *v1.Pod) bool {
	return !annotations.MapContainsBlackDuckEntries(oldObj.GetLabels(), newObj.GetLabels()) ||
		!annotations.MapContainsBlackDuckEntries(oldObj.GetAnnotations(), newObj.GetAnnotations())
}

func (pc *PodController) runWorker() {
	// Hot loop until we're told to stop.  processNextWorkItem will automatically wait until there's work
	// available, so we don't worry about secondary waits
	for pc.processNextWorkItem() {
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false when it's time to quit.
func (pc *PodController) processNextWorkItem() bool {
	// Pull the next work item from queue.  It should be a key we use to lookup something in a cache
	keyObj, quit := pc.queue.Get()
	if quit {
		return false
	}
	// You always have to indicate to the queue that you've completed a piece of work
	defer pc.queue.Done(keyObj)

	key := keyObj.(string)

	// Do your work on the key.  This method will contains your "do stuff" logic
	err := pc.syncHandler(key)
	if err == nil {
		// if you had no error, tell the queue to stop tracking history for your key.  This will
		// reset things like failure counts for per-item rate limiting
		pc.queue.Forget(key)
		return true
	}

	metrics.RecordError("controller", "unable to sync handler")

	// There was a failure so be sure to report it.  This method allows for pluggable error handling
	// which can be used for things like cluster-monitoring
	utilruntime.HandleError(fmt.Errorf("%v failed with : %v", key, err))

	// Since we failed, we should requeue the item to work on later.  This method will add a backoff
	// to avoid hotlooping on particular items (they're probably still not going to work right away)
	// and overall controller protection (everything I've done is broken, this controller needs to
	// calm down or it can starve other useful work) cases.
	pc.queue.AddRateLimited(key)

	return true
}

func (pc *PodController) processPod(key string) error {
	log.Infof("processing pod %s", key)

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		metrics.RecordError("controller", "error getting name of pod")
		return fmt.Errorf("error getting name of pod %q to get pod from informer: %v", key, err)
	}

	// Get the pod
	getPodStart := time.Now()
	pod, err := pc.podLister.Pods(ns).Get(name)
	metrics.RecordDuration("get pod -- pod controller", time.Now().Sub(getPodStart))
	if err != nil {
		metrics.RecordError("controller", "unable to get pod")
	}
	if errors.IsNotFound(err) {
		// Pod doesn't exist (anymore), so this is a delete event
		return communicator.SendPerceptorDeleteEvent(pc.podURL, name)
	} else if err != nil {
		return fmt.Errorf("error getting pod %s from informer: %v", name, err)
	}

	// Convert the pod from kubernetes to perceptor format and send to
	// the perceptor
	podInfo, err := mapper.NewPerceptorPodFromKubePod(pod)
	if err != nil {
		metrics.RecordError("controller", "unable to convert kube pod to perceptor pod")
		return fmt.Errorf("error converting pod to perceptor pod: %v", err)
	}
	return communicator.SendPerceptorAddEvent(pc.podURL, podInfo)
}
