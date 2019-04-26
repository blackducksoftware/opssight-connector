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
	"fmt"
	"math"
	"strings"
	"time"

	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
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

// This is a controller that updates the secret in perceptor periodically.
// It is assumed that the secret in perceptor will roll over any time this is updated, and
// if not, that there is a problem in the orchestration environment.

// Updater stores the opssight updater configuration
type Updater struct {
	config         *protoform.Config
	kubeClient     *kubernetes.Clientset
	hubClient      *hubclient.Clientset
	opssightClient *opssightclientset.Clientset
}

// NewUpdater returns the opssight updater configuration
func NewUpdater(config *protoform.Config, kubeClient *kubernetes.Clientset, hubClient *hubclient.Clientset, opssightClient *opssightclientset.Clientset) *Updater {
	return &Updater{
		config:         config,
		kubeClient:     kubeClient,
		hubClient:      hubClient,
		opssightClient: opssightClient,
	}
}

// Run watches for Black Duck and OpsSight events and update the internal Black Duck hosts in Perceptor secret and
// then patch the corresponding replication controller
func (p *Updater) Run(ch <-chan struct{}) {
	logger.Infof("Starting controller for hub<->perceptor updates... this blocks, so running in a go func.")

	syncFunc := func() {
		err := p.updateAllHubs()
		if len(err) > 0 {
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
		&blackduckapi.Blackduck{},
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
				running := p.isBlackDuckRunning(obj)
				if !running {
					syncFunc()
				}
			},
		},
	)

	// make sure this is called from a go func -- it blocks!
	go hubController.Run(ch)
	// go opssightController.Run(ch)
}

// isBlackDuckRunning return whether the Black Duck instance is in running state
func (p *Updater) isBlackDuckRunning(obj interface{}) bool {
	blackduck, _ := obj.(*blackduckapi.Blackduck)
	if strings.EqualFold(blackduck.Status.State, "Running") {
		return true
	}
	return false
}

// isOpsSightRunning return whether the OpsSight is in running state
func (p *Updater) isOpsSightRunning(obj interface{}) bool {
	opssight, _ := obj.(*opssightapi.OpsSight)
	if strings.EqualFold(opssight.Status.State, "Running") {
		return true
	}
	return false
}

// updateAllHubs will update the Black Duck instances in opssight resources
func (p *Updater) updateAllHubs() []error {
	opssights, err := util.GetOpsSights(p.opssightClient)
	if err != nil {
		return []error{errors.Annotate(err, "unable to get opssights")}
	}

	if len(opssights.Items) == 0 {
		return nil
	}

	errList := []error{}
	for _, opssight := range opssights.Items {
		err = p.updateOpsSight(&opssight)
		if err != nil {
			errList = append(errList, errors.Annotate(err, "unable to update perceptor"))
		}
	}
	return errList
}

// updateOpsSight will update the opssight resource with latest Black Duck instances
func (p *Updater) updateOpsSight(obj interface{}) error {
	opssight, ok := obj.(*opssightapi.OpsSight)
	if !ok {
		return errors.Errorf("unable to cast object")
	}
	var err error
	for j := 0; j < 20; j++ {
		if strings.EqualFold(opssight.Status.State, "running") {
			break
		}
		logger.Debugf("waiting for opssight %s to be up.....", opssight.Name)
		time.Sleep(10 * time.Second)

		opssight, err = util.GetOpsSight(p.opssightClient, p.config.Namespace, opssight.Name)
		if err != nil {
			return fmt.Errorf("unable to get opssight %s due to %+v", opssight.Name, err)
		}
	}
	err = p.update(opssight)
	return err
}

// update will list all Black Ducks in the cluster, and send them to opssight as scan targets.
func (p *Updater) update(opssight *opssightapi.OpsSight) error {
	hubType := opssight.Spec.Blackduck.BlackduckSpec.Type
	allHubs := p.getAllHubs(hubType)

	err := p.updateOpsSightCRD(&opssight.Spec, allHubs)
	if err != nil {
		return errors.Annotate(err, "unable to update opssight CRD")
	}
	return nil
}

// getAllHubs get only the internal Black Duck instances from the cluster
func (p *Updater) getAllHubs(hubType string) []*opssightapi.Host {
	hosts := []*opssightapi.Host{}
	hubsList, err := util.ListHubs(p.hubClient, p.config.Namespace)
	if err != nil {
		log.Errorf("unable to list blackducks due to %+v", err)
	}
	blackduckPassword := p.getDefaultPassword()
	for _, hub := range hubsList.Items {
		if strings.EqualFold(hub.Spec.Type, hubType) {
			var concurrentScanLimit int
			switch strings.ToUpper(hub.Spec.Size) {
			case "MEDIUM":
				concurrentScanLimit = 3
			case "LARGE":
				concurrentScanLimit = 4
			case "X-LARGE":
				concurrentScanLimit = 6
			default:
				concurrentScanLimit = 2
			}
			host := &opssightapi.Host{Domain: fmt.Sprintf("webserver.%s.svc", hub.Name), ConcurrentScanLimit: concurrentScanLimit, Scheme: "https", User: "sysadmin", Port: 443, Password: blackduckPassword}
			hosts = append(hosts, host)
		}
	}
	log.Debugf("total no of Black Duck's for type %s is %d", hubType, len(hosts))
	return hosts
}

// getDefaultPassword get the default password for the hub
func (p *Updater) getDefaultPassword() string {
	var hubPassword string
	var err error
	for dbInitTry := 0; dbInitTry < math.MaxInt32; dbInitTry++ {
		// get the secret from the default operator namespace, then copy it into the hub namespace.
		hubPassword, err = GetDefaultPasswords(p.kubeClient, p.config.Namespace)
		if err == nil {
			break
		} else {
			log.Infof("wasn't able to get hub password, sleeping 5 seconds.  try = %v", dbInitTry)
			time.Sleep(5 * time.Second)
		}
	}
	return hubPassword
}

// updateOpsSightCRD will update the opssight CRD
func (p *Updater) updateOpsSightCRD(opsSightSpec *opssightapi.OpsSightSpec, hubs []*opssightapi.Host) error {
	opssightName := opsSightSpec.Namespace
	logger.WithField("opssight", opssightName).Info("update opssight: looking for opssight")
	opssight, err := p.opssightClient.SynopsysV1().OpsSights(p.config.Namespace).Get(opssightName, metav1.GetOptions{})
	if err != nil {
		return errors.Annotatef(err, "unable to get opssight %s in %s namespace", opssightName, opsSightSpec.Namespace)
	}

	opssight.Status.InternalHosts = p.appendBlackDuckHosts(opssight.Status.InternalHosts, hubs)

	_, err = p.opssightClient.SynopsysV1().OpsSights(p.config.Namespace).Update(opssight)
	if err != nil {
		return errors.Annotatef(err, "unable to update opssight %s in %s", opssightName, opsSightSpec.Namespace)
	}
	return nil
}

// appendBlackDuckHosts will append the hosts of external and internal Black Duck
func (p *Updater) appendBlackDuckHosts(existingBlackDucks []*opssightapi.Host, internalBlackDucks []*opssightapi.Host) []*opssightapi.Host {
	finalBlackDucks := []*opssightapi.Host{}
	// remove the deleted Black Duck from the final Black Duck list
	for _, existingBlackDuck := range existingBlackDucks {
		isExist := false
		for _, internalBlackDuck := range internalBlackDucks {
			if strings.EqualFold(internalBlackDuck.Domain, existingBlackDuck.Domain) {
				isExist = true
				break
			}
		}
		if isExist {
			finalBlackDucks = append(finalBlackDucks, existingBlackDuck)
		}
	}

	// add the new Black Duck to the final Black Duck list
	for _, internalBlackDuck := range internalBlackDucks {
		isExist := false
		for _, finalBlackDuck := range finalBlackDucks {
			if strings.EqualFold(internalBlackDuck.Domain, finalBlackDuck.Domain) {
				isExist = true
				break
			}
		}
		if !isExist {
			finalBlackDucks = append(finalBlackDucks, internalBlackDuck)
		}
	}
	return finalBlackDucks
}

// appendBlackDuckSecrets will append the secrets of external and internal Black Duck
func (p *Updater) appendBlackDuckSecrets(existingBlackDucks map[string]*opssightapi.Host, internalBlackDucks []*opssightapi.Host) map[string]*opssightapi.Host {
	// remove the deleted Black Duck from the Black Duck secret
	for _, existingBlackDuck := range existingBlackDucks {
		isExist := false
		for _, internalBlackDuck := range internalBlackDucks {
			if strings.EqualFold(internalBlackDuck.Domain, existingBlackDuck.Domain) {
				isExist = true
				break
			}
		}
		if !isExist {
			delete(existingBlackDucks, existingBlackDuck.Domain)
		}
	}

	// add the new Black Duck to the Black Duck secret
	for _, internalBlackDuck := range internalBlackDucks {
		if _, ok := existingBlackDucks[internalBlackDuck.Domain]; !ok {
			existingBlackDucks[internalBlackDuck.Domain] = internalBlackDuck
		}
	}
	return existingBlackDucks
}
