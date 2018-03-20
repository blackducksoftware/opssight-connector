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
	"strings"
	"time"

	perceptorapi "github.com/blackducksoftware/perceptor/pkg/api"

	"github.com/blackducksoftware/perceivers/image/pkg/mapper"
	"github.com/blackducksoftware/perceivers/image/pkg/metrics"
	"github.com/blackducksoftware/perceivers/pkg/communicator"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	imageapi "github.com/openshift/api/image/v1"
	imageclient "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	imagelister "github.com/openshift/client-go/image/listers/image/v1"
)

type OSImageStreamController struct {
	client            *imageclient.ImageV1Client
	imageController   cache.Controller
	indexer           cache.Indexer
	imageStreamLister imagelister.ImageStreamLister
	imageURL          string

	syncHandler func(obj *imageapi.ImageStream) error
	queue       workqueue.RateLimitingInterface
}

func NewOSImageStreamController(oic *imageclient.ImageV1Client, perceptorURL string) *OSImageStreamController {
	osisc := OSImageStreamController{
		client:   oic,
		queue:    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ImageStreams"),
		imageURL: fmt.Sprintf("%s/%s", perceptorURL, perceptorapi.ImagePath),
	}
	osisc.indexer, osisc.imageController = cache.NewIndexerInformer(
		&cache.ListWatch{
			ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
				return osisc.client.ImageStreams(metav1.NamespaceAll).List(opts)
			},
			WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
				return osisc.client.ImageStreams(metav1.NamespaceAll).Watch(opts)
			},
		},
		&imageapi.ImageStream{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				osisc.queue.Add(obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				old, ok1 := oldObj.(*imageapi.ImageStream)
				new, ok2 := newObj.(*imageapi.ImageStream)
				if ok1 && ok2 && osisc.needsUpdate(old, new) {
					osisc.queue.Add(newObj)
				}
			},
			DeleteFunc: func(obj interface{}) {
				osisc.queue.Add(obj)
			},
		},
		cache.Indexers{},
	)
	osisc.imageStreamLister = imagelister.NewImageStreamLister(osisc.indexer)
	osisc.syncHandler = osisc.processImageStream

	return &osisc
}

func (osisc *OSImageStreamController) Run(threadiness int, stopCh <-chan struct{}) {
	defer osisc.queue.ShutDown()

	go osisc.imageController.Run(stopCh)

	// start up your worker threads based on threadiness.  Some controllers have multiple kinds of workers
	for i := 0; i < threadiness; i++ {
		// runWorker will loop until "something bad" happens.  The .Until will then rekick the worker
		// after one second
		go wait.Until(osisc.runWorker, time.Second, stopCh)
	}

	// wait until we're told to stop
	<-stopCh
}

func (osisc *OSImageStreamController) needsUpdate(oldObj *imageapi.ImageStream, newObj *imageapi.ImageStream) bool {
	return true
}

func (osisc *OSImageStreamController) runWorker() {
	// hot loop until we're told to stop.  processNextWorkItem will automatically wait until there's work
	// available, so we don't worry about secondary waits
	for osisc.processNextWorkItem() {
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false when it's time to quit.
func (osisc *OSImageStreamController) processNextWorkItem() bool {
	// pull the next work item from queue.  It should be a key we use to lookup something in a cache
	keyObj, quit := osisc.queue.Get()
	if quit {
		return false
	}
	// you always have to indicate to the queue that you've completed a piece of work
	defer osisc.queue.Done(keyObj)

	key, ok := keyObj.(*imageapi.ImageStream)
	if !ok {
		metrics.RecordError("imagestream_controller", "key wasn't an imagestream")
		utilruntime.HandleError(fmt.Errorf("key wasn't an imagestream"))
		return true
	}

	// do your work on the key.  This method will contains your "do stuff" logic
	err := osisc.syncHandler(key)
	if err == nil {
		// if you had no error, tell the queue to stop tracking history for your key.  This will
		// reset things like failure counts for per-item rate limiting
		osisc.queue.Forget(key)
		return true
	}

	metrics.RecordError("imagestream_controller", "unable to sync handler")

	// there was a failure so be sure to report it.  This method allows for pluggable error handling
	// which can be used for things like cluster-monitoring
	utilruntime.HandleError(fmt.Errorf("%v failed with : %v", key, err))

	// since we failed, we should requeue the item to work on later.  This method will add a backoff
	// to avoid hotlooping on particular items (they're probably still not going to work right away)
	// and overall controller protection (everything I've done is broken, this controller needs to
	// calm down or it can starve other useful work) cases.
	osisc.queue.AddRateLimited(key)

	return true
}

func (osisc *OSImageStreamController) processImageStream(obj *imageapi.ImageStream) error {
	errList := []string{}
	// Get an updated version of this imagestream if it exists
	getImageStream := time.Now()
	is, err := osisc.imageStreamLister.ImageStreams(metav1.NamespaceAll).Get(obj.GetName())
	metrics.RecordDuration("get image stream", time.Now().Sub(getImageStream))
	if err != nil {
		metrics.RecordError("imagestream_controller", "unable to get updated version of image stream")
	}
	if errors.IsNotFound(err) {
		// ImageStream doesn't exist (anymore), so this is a delete event
		images, err := osisc.getImagesFromImageStream(obj)
		metrics.RecordError("imagestream_controller", "unable to get images from image stream #1")
		if err != nil {
			return err
		}
		for _, image := range images {
			err = communicator.SendPerceptorDeleteEvent(osisc.imageURL, image.Name)
			metrics.RecordHttpStats(osisc.imageURL, err == nil)
			if err != nil {
				metrics.RecordError("imagestream_controller", "unable to send delete event")
				errList = append(errList, err.Error())
			}
		}
		return fmt.Errorf(strings.Join(errList, ","))
	} else if err != nil {
		is = obj
	}

	images, err := osisc.getImagesFromImageStream(is)
	if err != nil {
		metrics.RecordError("imagestream_controller", "unable to get images from image stream #2")
		return err
	}
	for _, image := range images {
		err = communicator.SendPerceptorAddEvent(osisc.imageURL, image)
		metrics.RecordHttpStats(osisc.imageURL, err == nil)
		if err != nil {
			metrics.RecordError("imagestream_controller", "unable to send add event")
			errList = append(errList, err.Error())
		}
	}
	return fmt.Errorf(strings.Join(errList, ","))
}

func (osisc *OSImageStreamController) getImagesFromImageStream(stream *imageapi.ImageStream) ([]*perceptorapi.Image, error) {
	tags := stream.Status.Tags
	if tags == nil {
		metrics.RecordError("imagestream_controller", "image stream has no tags")
		return nil, fmt.Errorf("image stream %s has no tags", stream.GetName())
	}

	digest := stream.Spec.DockerImageRepository
	images := []*perceptorapi.Image{}
	for _, events := range tags {
		ref := events.Items[0].Image
		getImagesStart := time.Now()
		image, err := osisc.client.Images().Get(ref, metav1.GetOptions{})
		metrics.RecordDuration("get images from image stream", time.Now().Sub(getImagesStart))
		if err != nil {
			metrics.RecordError("imagestream_controller", "error getting image")
			return nil, fmt.Errorf("error getting image %s@%s: %v", digest, ref, err)
		}

		// We ignore errors here so we don't lose all the images
		imageInfo, _ := mapper.NewPerceptorImageFromOSImage(image)
		images = append(images, imageInfo)
	}
	return images, nil
}
