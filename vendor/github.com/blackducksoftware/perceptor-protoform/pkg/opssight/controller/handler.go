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
	"strings"

	opssight_v1 "github.com/blackducksoftware/perceptor-protoform/pkg/api/opssight/v1"
	"github.com/blackducksoftware/perceptor-protoform/pkg/model"
	"github.com/blackducksoftware/perceptor-protoform/pkg/opssight"
	opssightclientset "github.com/blackducksoftware/perceptor-protoform/pkg/opssight/client/clientset/versioned"
	"github.com/blackducksoftware/perceptor-protoform/pkg/util"
	"github.com/imdario/mergo"
	routeclient "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	securityclient "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Handler interface contains the methods that are required
type Handler interface {
	ObjectCreated(obj interface{})
	ObjectDeleted(obj string)
	ObjectUpdated(objOld, objNew interface{})
}

// OpsSightHandler will store the configuration that is required to initiantiate the informers callback
type OpsSightHandler struct {
	Config            *model.Config
	KubeConfig        *rest.Config
	Clientset         *kubernetes.Clientset
	OpsSightClientset *opssightclientset.Clientset
	Defaults          *opssight_v1.OpsSightSpec
	Namespace         string
	CmMutex           chan bool
	OSSecurityClient  *securityclient.SecurityV1Client
	RouteClient       *routeclient.RouteV1Client
}

// ObjectCreated will be called for create opssight events
func (h *OpsSightHandler) ObjectCreated(obj interface{}) {
	log.Debugf("objectCreated: %+v", obj)
	opssightv1 := obj.(*opssight_v1.OpsSight)
	if strings.EqualFold(opssightv1.Spec.State, "") {
		newSpec := opssightv1.Spec
		defaultSpec := h.Defaults
		err := mergo.Merge(&newSpec, defaultSpec)
		log.Debugf("merged opssight details for %s: %+v", newSpec)
		if err != nil {
			h.updateState("error", "error", err.Error(), opssightv1)
		} else {
			opssightv1.Spec = newSpec
			opssightv1, err := h.updateState("pending", "creating", "", opssightv1)

			if err == nil {
				opssightCreator := opssight.NewCreater(h.Config, h.KubeConfig, h.Clientset, h.OpsSightClientset, h.OSSecurityClient, h.RouteClient)

				err = opssightCreator.CreateOpsSight(&opssightv1.Spec)
				if err != nil {
					log.Errorf("unable to create opssight %s due to %s", opssightv1.Name, err.Error())
				}

				opssightv1, err1 := util.GetOpsSight(h.OpsSightClientset, opssightv1.Name, opssightv1.Name)
				if err1 != nil {
					log.Errorf("unable to get the opssight %s due to %+v", opssightv1.Name, err1)
				} else {
					if err != nil {
						h.updateState("error", "error", err.Error(), opssightv1)
					} else {
						h.updateState("running", "running", "", opssightv1)
					}
				}
			}
		}
	}
}

// ObjectDeleted will be called for delete opssight events
func (h *OpsSightHandler) ObjectDeleted(name string) {
	log.Debugf("objectDeleted: %+v", name)
	opssightCreator := opssight.NewCreater(h.Config, h.KubeConfig, h.Clientset, h.OpsSightClientset, h.OSSecurityClient, h.RouteClient)
	err := opssightCreator.DeleteOpsSight(name)
	if err != nil {
		log.Errorf("unable to delete opssight: %v", err)
	}
}

// ObjectUpdated will be called for update opssight events
func (h *OpsSightHandler) ObjectUpdated(objOld, objNew interface{}) {
	log.Debugf("objectUpdated: %+v", objNew)
}

func (h *OpsSightHandler) updateState(specState string, statusState string, errorMessage string, opssight *opssight_v1.OpsSight) (*opssight_v1.OpsSight, error) {
	opssight.Spec.State = specState
	opssight.Status.State = statusState
	opssight.Status.ErrorMessage = errorMessage
	opssight, err := h.updateOpsSightObject(opssight)
	if err != nil {
		log.Errorf("couldn't update the state of opssight object: %s", err.Error())
	}
	return opssight, err
}

func (h *OpsSightHandler) updateOpsSightObject(obj *opssight_v1.OpsSight) (*opssight_v1.OpsSight, error) {
	return h.OpsSightClientset.SynopsysV1().OpsSights(h.Namespace).Update(obj)
}
