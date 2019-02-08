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
	"strings"

	opssight_v1 "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
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

// Handler will store the configuration that is required to initiantiate the informers callback
type Handler struct {
	Config            *protoform.Config
	KubeConfig        *rest.Config
	Clientset         *kubernetes.Clientset
	OpsSightClientset *opssightclientset.Clientset
	Defaults          *opssight_v1.OpsSightSpec
	Namespace         string
	OSSecurityClient  *securityclient.SecurityV1Client
	RouteClient       *routeclient.RouteV1Client
	HubClient         *hubclientset.Clientset
}

// ObjectCreated will be called for create opssight events
func (h *Handler) ObjectCreated(obj interface{}) {
	if err := h.handleObjectCreated(obj); err != nil {
		log.Errorf("unable to handle object created: %s", err.Error())
	}
}

func (h *Handler) handleObjectCreated(obj interface{}) error {
	recordEvent("objectCreated")
	log.Debugf("objectCreated: %+v", obj)
	opssightv1, ok := obj.(*opssight_v1.OpsSight)
	if !ok {
		return errors.Errorf("unable to cast")
	}
	if !strings.EqualFold(opssightv1.Spec.State, "") {
		return nil // ??? why nil?
	}
	newSpec := opssightv1.Spec
	defaultSpec := h.Defaults
	err := mergo.Merge(&newSpec, defaultSpec)
	if err != nil {
		recordError("unable to merge objects")
		h.updateState("error", "error", err.Error(), opssightv1)
		return errors.Annotate(err, "unable to merge objects")
	}
	bytes, err := json.Marshal(newSpec)
	log.Debugf("merged opssight details: %+v", newSpec)
	log.Debugf("opssight json (%+v): %s", err, string(bytes))

	opssightv1.Spec = newSpec
	opssightv1, err = h.updateState("pending", "creating", "", opssightv1)
	if err != nil {
		recordError("unable to update state")
		return errors.Annotate(err, "unable to update state")
	}
	opssightCreator := NewCreater(h.Config, h.KubeConfig, h.Clientset, h.OpsSightClientset, h.OSSecurityClient, h.RouteClient, h.HubClient)

	err = opssightCreator.CreateOpsSight(&opssightv1.Spec)
	if err != nil {
		recordError("unable to create opssight")
		h.updateState("error", "error", err.Error(), opssightv1)
		return errors.Annotatef(err, "unable to create opssight %s", opssightv1.Name)
	}

	opssightv1, err = util.GetOpsSight(h.OpsSightClientset, opssightv1.Name, opssightv1.Name)
	if err != nil {
		recordError("unable to get opssight")
		return errors.Annotatef(err, "unable to get the opssight %s", opssightv1.Name)
	}

	h.updateState("running", "running", "", opssightv1)
	return nil
}

// ObjectDeleted will be called for delete opssight events
func (h *Handler) ObjectDeleted(name string) {
	recordEvent("objectDeleted")
	log.Debugf("objectDeleted: %+v", name)
	opssightCreator := NewCreater(h.Config, h.KubeConfig, h.Clientset, h.OpsSightClientset, h.OSSecurityClient, h.RouteClient, h.HubClient)
	err := opssightCreator.DeleteOpsSight(name)
	if err != nil {
		log.Errorf("unable to delete opssight: %v", err)
		recordError("unable to delete opssight")
	}
}

// ObjectUpdated will be called for update opssight events
func (h *Handler) ObjectUpdated(objOld, objNew interface{}) {
	recordEvent("objectUpdated")
	log.Debugf("objectUpdated: %+v", objNew)
}

func (h *Handler) updateState(specState string, statusState string, errorMessage string, opssight *opssight_v1.OpsSight) (*opssight_v1.OpsSight, error) {
	opssight.Spec.State = specState
	opssight.Status.State = statusState
	opssight.Status.ErrorMessage = errorMessage
	opssight, err := h.updateOpsSightObject(opssight)
	return opssight, errors.Annotate(err, "unable to update the state of opssight object")
}

func (h *Handler) updateOpsSightObject(obj *opssight_v1.OpsSight) (*opssight_v1.OpsSight, error) {
	return h.OpsSightClientset.SynopsysV1().OpsSights(h.Namespace).Update(obj)
}
