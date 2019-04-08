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
	"encoding/json"
	"fmt"
	"strings"

	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	hubclientset "github.com/blackducksoftware/synopsys-operator/pkg/blackduck/client/clientset/versioned"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/imdario/mergo"
	"github.com/juju/errors"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	securityclient "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// HandlerInterface contains the methods that are required
// ... not really sure why we have this type
type HandlerInterface interface {
	ObjectCreated(obj interface{})
	ObjectDeleted(obj string)
	ObjectUpdated(objOld, objNew interface{})
}

// State contains the state of the OpsSight
type State string

// DesiredState contains the desired state of the OpsSight
type DesiredState string

const (
	// Creating is used when OpsSight is installing or deploying
	Creating State = "Creating"
	// Running is used when OpsSight is running
	Running State = "Running"
	// Stopping is used when OpsSight is about to stip
	Stopping State = "Stopping"
	// Stopped is used when OpsSight to be stopped
	Stopped State = "Stopped"
	// Updating is used when OpsSight is about to updating
	Updating State = "Updating"
	// Error is used when OpsSight deployment errored out
	Error State = "Error"

	// Start is used when OpsSight deployment to be created or updated
	Start DesiredState = "Start"
	// Stop is used when OpsSight deployment to be stopped
	Stop DesiredState = "Stop"
)

// Handler will store the configuration that is required to initiantiate the informers callback
type Handler struct {
	Config           *protoform.Config
	KubeConfig       *rest.Config
	KubeClient       *kubernetes.Clientset
	OpsSightClient   *opssightclientset.Clientset
	Defaults         *opssightapi.OpsSightSpec
	Namespace        string
	OSSecurityClient *securityclient.SecurityV1Client
	RouteClient      *routeclient.RouteV1Client
	HubClient        *hubclientset.Clientset
}

// ObjectCreated will be called for create opssight events
func (h *Handler) ObjectCreated(obj interface{}) {
	if err := h.handleObjectCreated(obj); err != nil {
		log.Errorf("handle object created: %s", err.Error())
	}
}

func (h *Handler) handleObjectCreated(obj interface{}) error {
	recordEvent("objectCreated")
	// log.Debugf("objectCreated: %+v", obj)
	opssight, ok := obj.(*opssightapi.OpsSight)
	if !ok {
		recordError("unable to cast opssight object")
		return errors.Errorf("unable to cast opssight object")
	}

	if strings.EqualFold(opssight.Spec.DesiredState, string(Start)) && strings.EqualFold(opssight.Status.State, "") {
		log.Debugf("inside creation event of opssight %s", opssight.Spec.Namespace)
		newSpec := opssight.Spec
		defaultSpec := h.Defaults
		err := mergo.Merge(&newSpec, defaultSpec)
		if err != nil {
			recordError("unable to merge default and new objects")
			h.updateState(Error, err.Error(), opssight)
			return errors.Annotate(err, "unable to merge default and new objects")
		}

		bytes, err := json.Marshal(newSpec)
		log.Debugf("merged opssight details: %+v", newSpec)
		log.Debugf("opssight json (%+v): %s", err, string(bytes))

		opssight.Spec = newSpec
		opssight, err = h.updateState(Creating, "", opssight)
		if err != nil {
			recordError("unable to update state")
			return errors.Annotate(err, "unable to update state")
		}

		opssightCreator := NewCreater(h.Config, h.KubeConfig, h.KubeClient, h.OpsSightClient, h.OSSecurityClient, h.RouteClient, h.HubClient)
		err = opssightCreator.CreateOpsSight(opssight)
		if err != nil {
			recordError("unable to create opssight")
			h.updateState(Error, err.Error(), opssight)
			return errors.Annotatef(err, "create opssight %s", opssight.Name)
		}
		h.updateState(Running, "", opssight)
	} else {
		h.ObjectUpdated(nil, obj)
	}
	return nil
}

// ObjectDeleted will be called for delete opssight events
func (h *Handler) ObjectDeleted(name string) {
	recordEvent("objectDeleted")
	log.Debugf("objectDeleted: %+v", name)
	opssightCreator := NewCreater(h.Config, h.KubeConfig, h.KubeClient, h.OpsSightClient, h.OSSecurityClient, h.RouteClient, h.HubClient)
	err := opssightCreator.DeleteOpsSight(name)
	if err != nil {
		log.Errorf("unable to delete opssight: %v", err)
		recordError("unable to delete opssight")
	}
}

// ObjectUpdated will be called for update opssight events
func (h *Handler) ObjectUpdated(objOld, objNew interface{}) {
	recordEvent("objectUpdated")
	// log.Debugf("objectUpdated: %+v", objNew)
	opssight, ok := objNew.(*opssightapi.OpsSight)
	if !ok {
		recordError("unable to cast opssight object")
		log.Error("Unable to cast OpsSight object")
		return
	}

	switch strings.ToUpper(opssight.Spec.DesiredState) {
	case "STOP":
		opssightCreator := NewCreater(h.Config, h.KubeConfig, h.KubeClient, h.OpsSightClient, h.OSSecurityClient, h.RouteClient, h.HubClient)
		opssight, err := h.updateState(Stopping, "", opssight)
		if err != nil {
			recordError("unable to update state")
			log.Error(errors.Annotate(err, "unable to update stopping state"))
			return
		}
		err = opssightCreator.StopOpsSight(&opssight.Spec)
		if err != nil {
			recordError("unable to stop opssight")
			h.updateState(Error, err.Error(), opssight)
			log.Errorf("handle object stop: %s", err.Error())
			return
		}
		opssight, err = util.GetOpsSight(h.OpsSightClient, opssight.Spec.Namespace, opssight.Spec.Namespace)
		if err != nil {
			recordError("unable to get opssight")
			log.Error(errors.Annotate(err, "unable to get opssight"))
			return
		}
		_, err = h.updateState(Stopped, "", opssight)
		if err != nil {
			recordError("unable to update state")
			log.Error(errors.Annotate(err, "unable to update stopped state"))
			return
		}
	case "START":
		opssightCreator := NewCreater(h.Config, h.KubeConfig, h.KubeClient, h.OpsSightClient, h.OSSecurityClient, h.RouteClient, h.HubClient)
		opssight, err := h.updateState(Updating, "", opssight)
		if err != nil {
			recordError("unable to update state")
			log.Error(errors.Annotate(err, "unable to update updating state"))
			return
		}
		err = opssightCreator.UpdateOpsSight(opssight)
		if err != nil {
			recordError("unable to update opssight")
			h.updateState(Error, err.Error(), opssight)
			log.Errorf("handle object update: %s", err.Error())
			return
		}
		opssight, err = util.GetOpsSight(h.OpsSightClient, opssight.Spec.Namespace, opssight.Spec.Namespace)
		if err != nil {
			recordError("unable to get opssight")
			log.Error(errors.Annotate(err, "unable to get opssight"))
			return
		}
		_, err = h.updateState(Running, "", opssight)
		if err != nil {
			recordError("unable to update state")
			log.Error(errors.Annotate(err, "unable to update running state"))
			return
		}
	default:
		recordError("unable to find the desired state value")
		log.Errorf("unable to handle object update due to %+v", fmt.Errorf("desired state value is not expected"))
		return
	}
}

func (h *Handler) updateState(state State, errorMessage string, opssight *opssightapi.OpsSight) (*opssightapi.OpsSight, error) {
	opssight.Status.State = string(state)
	opssight.Status.ErrorMessage = errorMessage
	opssight, err := h.updateOpsSightObject(opssight)
	if err != nil {
		return nil, errors.Annotate(err, "unable to update the state of opssight object")
	}
	return opssight, nil
}

func (h *Handler) updateOpsSightObject(obj *opssightapi.OpsSight) (*opssightapi.OpsSight, error) {
	return h.OpsSightClient.SynopsysV1().OpsSights(h.Namespace).Update(obj)
}
