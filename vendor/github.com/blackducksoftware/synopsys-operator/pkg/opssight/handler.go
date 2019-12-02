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
	"fmt"
	"reflect"
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
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	// Running is used when OpsSight is running
	Running State = "Running"
	// Stopped is used when OpsSight to be stopped
	Stopped State = "Stopped"
	// Error is used when OpsSight deployment errored out
	Error State = "Error"

	// Start is used when OpsSight deployment to be created or updated
	Start DesiredState = ""
	// Stop is used when OpsSight deployment to be stopped
	Stop DesiredState = "Stop"
)

// Handler will store the configuration that is required to initiantiate the informers callback
type Handler struct {
	Config                  *protoform.Config
	KubeConfig              *rest.Config
	KubeClient              *kubernetes.Clientset
	OpsSightClient          *opssightclientset.Clientset
	IsBlackDuckClusterScope bool
	Defaults                *opssightapi.OpsSightSpec
	Namespace               string
	OSSecurityClient        *securityclient.SecurityV1Client
	RouteClient             *routeclient.RouteV1Client
	HubClient               *hubclientset.Clientset
}

// ObjectCreated will be called for create opssight events
func (h *Handler) ObjectCreated(obj interface{}) {
	if err := h.handleObjectCreated(obj); err != nil {
		log.Errorf("handle opssight: %s", err.Error())
	}
}

func (h *Handler) handleObjectCreated(obj interface{}) error {
	h.ObjectUpdated(nil, obj)
	return nil
}

// ObjectDeleted will be called for delete opssight events
func (h *Handler) ObjectDeleted(name string) {
	log.Debugf("objectDeleted: %+v", name)

	// if cluster scope, then check whether the OpsSight CRD exist. If not exist, then don't delete the instance
	if h.Config.IsClusterScoped {
		apiClientset, err := clientset.NewForConfig(h.KubeConfig)
		crd, err := apiClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(util.OpsSightCRDName, metav1.GetOptions{})
		if err != nil || crd.DeletionTimestamp != nil {
			// We do not delete the OpsSight instance if the CRD doesn't exist or that it is in the process of being deleted
			log.Warnf("Ignoring request to delete %s because the %s CRD doesn't exist or is being deleted", name, util.OpsSightCRDName)
			return
		}
	}

	opssightCreator := NewCreater(h.Config, h.KubeConfig, h.KubeClient, h.OpsSightClient, h.OSSecurityClient, h.RouteClient, h.HubClient, h.IsBlackDuckClusterScope)
	err := opssightCreator.DeleteOpsSight(name)
	if err != nil {
		log.Errorf("unable to delete opssight: %v", err)
	}
}

// ObjectUpdated will be called for update opssight events
func (h *Handler) ObjectUpdated(objOld, objNew interface{}) {
	// log.Debugf("objectUpdated: %+v", objNew)
	opssight, ok := objNew.(*opssightapi.OpsSight)
	if !ok {
		log.Error("Unable to cast OpsSight object")
		return
	}

	var err error
	if _, ok = opssight.Annotations["synopsys.com/created.by"]; !ok {
		opssight.Annotations = util.InitAnnotations(opssight.Annotations)
		opssight.Annotations["synopsys.com/created.by"] = h.Config.Version
		opssight, err = util.UpdateOpsSight(h.OpsSightClient, h.Config.CrdNamespace, opssight)
		if err != nil {
			log.Errorf("couldn't update the annotation for %s OpsSight instance in %s namespace due to %+v", opssight.Name, opssight.Spec.Namespace, err)
			return
		}
	}

	newSpec := opssight.Spec
	defaultSpec := h.Defaults
	err = mergo.Merge(&newSpec, defaultSpec)
	if err != nil {
		h.updateState(Error, err.Error(), opssight)
		log.Errorf("unable to merge default and new objects due to %+v", err)
		return
	}

	opssight.Spec = newSpec

	switch strings.ToUpper(opssight.Spec.DesiredState) {
	case "STOP":
		opssightCreator := NewCreater(h.Config, h.KubeConfig, h.KubeClient, h.OpsSightClient, h.OSSecurityClient, h.RouteClient, h.HubClient, h.IsBlackDuckClusterScope)
		err = opssightCreator.StopOpsSight(&opssight.Spec)
		if err != nil {
			h.updateState(Error, err.Error(), opssight)
			log.Errorf("handle object stop: %s", err.Error())
			return
		}

		_, err = h.updateState(Stopped, "", opssight)
		if err != nil {
			log.Error(errors.Annotate(err, "unable to update stopped state"))
			return
		}
	case "":
		opssightCreator := NewCreater(h.Config, h.KubeConfig, h.KubeClient, h.OpsSightClient, h.OSSecurityClient, h.RouteClient, h.HubClient, h.IsBlackDuckClusterScope)
		err = opssightCreator.UpdateOpsSight(opssight)
		if err != nil {
			h.updateState(Error, err.Error(), opssight)
			log.Errorf("handle object update: %s", err.Error())
			return
		}

		if !strings.EqualFold(opssight.Status.State, string(Running)) {
			_, err = h.updateState(Running, "", opssight)
			if err != nil {
				log.Error(errors.Annotate(err, "unable to update running state"))
				return
			}
		}
	default:
		log.Errorf("unable to handle object update due to %+v", fmt.Errorf("desired state value is not expected"))
		return
	}
}

func (h *Handler) updateState(state State, errorMessage string, opssight *opssightapi.OpsSight) (*opssightapi.OpsSight, error) {
	newOpssight, err := util.GetOpsSight(h.OpsSightClient, h.Config.CrdNamespace, opssight.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Annotate(err, "unable to get the opssight to update the state of opssight object")
	}

	if !reflect.DeepEqual(newOpssight.Status.State, string(state)) || !reflect.DeepEqual(newOpssight.Status.ErrorMessage, errorMessage) {
		newOpssight.Spec = opssight.Spec
		newOpssight.Status.State = string(state)
		newOpssight.Status.ErrorMessage = errorMessage
		newOpssight, err = h.updateOpsSightObject(newOpssight)
		if err != nil {
			return nil, errors.Annotate(err, "unable to update the state of opssight object")
		}
	}
	return newOpssight, nil
}

func (h *Handler) updateOpsSightObject(obj *opssightapi.OpsSight) (*opssightapi.OpsSight, error) {
	return util.UpdateOpsSight(h.OpsSightClient, h.Config.CrdNamespace, obj)
}
