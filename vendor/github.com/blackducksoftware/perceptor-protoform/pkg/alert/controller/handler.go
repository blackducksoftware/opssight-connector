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

	"github.com/blackducksoftware/perceptor-protoform/pkg/alert"
	alertclientset "github.com/blackducksoftware/perceptor-protoform/pkg/alert/client/clientset/versioned"
	alert_v1 "github.com/blackducksoftware/perceptor-protoform/pkg/api/alert/v1"
	"github.com/imdario/mergo"
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

// AlertHandler will store the configuration that is required to initiantiate the informers callback
type AlertHandler struct {
	Config         *rest.Config
	Clientset      *kubernetes.Clientset
	AlertClientset *alertclientset.Clientset
	Defaults       *alert_v1.AlertSpec
	Namespace      string
	CmMutex        chan bool
}

// ObjectCreated will be called for create alert events
func (h *AlertHandler) ObjectCreated(obj interface{}) {
	log.Debugf("objectCreated: %+v", obj)
	alertv1 := obj.(*alert_v1.Alert)
	if strings.EqualFold(alertv1.Spec.State, "") {
		// merge with default values
		newSpec := alertv1.Spec
		alertDefaultSpec := h.Defaults
		err := mergo.Merge(&newSpec, alertDefaultSpec)
		log.Debugf("merged alert details %+v", newSpec)
		if err != nil {
			log.Errorf("unable to merge the alert structs for %s due to %+v", alertv1.Name, err)
			//Set spec/state  and status/state to started
			h.updateState("error", "error", fmt.Sprintf("unable to merge the alert structs for %s due to %+v", alertv1.Name, err), alertv1)
		} else {
			alertv1.Spec = newSpec
			// update status
			alertv1, err := h.updateState("pending", "creating", "", alertv1)

			if err == nil {
				alertCreator := alert.NewCreater(h.Config, h.Clientset, h.AlertClientset)

				// create alert instance
				err = alertCreator.CreateAlert(&alertv1.Spec)

				if err != nil {
					h.updateState("error", "error", fmt.Sprintf("%+v", err), alertv1)
				} else {
					h.updateState("running", "running", "", alertv1)
				}
			}
		}
	}
}

// ObjectDeleted will be called for delete alert events
func (h *AlertHandler) ObjectDeleted(name string) {
	log.Debugf("objectDeleted: %+v", name)
	alertCreator := alert.NewCreater(h.Config, h.Clientset, h.AlertClientset)
	alertCreator.DeleteAlert(name)
}

// ObjectUpdated will be called for update alert events
func (h *AlertHandler) ObjectUpdated(objOld, objNew interface{}) {
	log.Debugf("objectUpdated: %+v", objNew)
}

func (h *AlertHandler) updateState(specState string, statusState string, errorMessage string, alert *alert_v1.Alert) (*alert_v1.Alert, error) {
	alert.Spec.State = specState
	alert.Status.State = statusState
	alert.Status.ErrorMessage = errorMessage
	alert, err := h.updateAlertObject(alert)
	if err != nil {
		log.Errorf("couldn't update the state of alert object: %s", err.Error())
	}
	return alert, err
}

func (h *AlertHandler) updateAlertObject(obj *alert_v1.Alert) (*alert_v1.Alert, error) {
	return h.AlertClientset.SynopsysV1().Alerts(h.Namespace).Update(obj)
}
