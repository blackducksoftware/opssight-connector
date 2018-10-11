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
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	hub_v1 "github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
	"github.com/blackducksoftware/perceptor-protoform/pkg/hub"
	hubclientset "github.com/blackducksoftware/perceptor-protoform/pkg/hub/client/clientset/versioned"
	"github.com/blackducksoftware/perceptor-protoform/pkg/model"
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

// HubHandler will store the configuration that is required to initiantiate the informers callback
type HubHandler struct {
	Config           *model.Config
	KubeConfig       *rest.Config
	Clientset        *kubernetes.Clientset
	HubClientset     *hubclientset.Clientset
	Defaults         *hub_v1.HubSpec
	Namespace        string
	FederatorBaseURL string
	CmMutex          chan bool
	OSSecurityClient *securityclient.SecurityV1Client
	RouteClient      *routeclient.RouteV1Client
}

// APISetHubsRequest to set the Hub urls for Perceptor
type APISetHubsRequest struct {
	HubURLs []string
}

// ObjectCreated will be called for create hub events
func (h *HubHandler) ObjectCreated(obj interface{}) {
	log.Debugf("ObjectCreated: %+v", obj)
	hubv1 := obj.(*hub_v1.Hub)
	if strings.EqualFold(hubv1.Spec.State, "") {
		newSpec := hubv1.Spec
		hubDefaultSpec := h.Defaults
		err := mergo.Merge(&newSpec, hubDefaultSpec)
		log.Debugf("merged hub details %+v", newSpec)
		if err != nil {
			log.Errorf("unable to merge the hub structs for %s due to %+v", hubv1.Name, err)
			h.updateState("error", "error", fmt.Sprintf("%+v", err), hubv1)
		} else {
			hubv1.Spec = newSpec
			// Update status
			hubv1, err := h.updateState("pending", "creating", "", hubv1)

			if err == nil {
				hubCreator := hub.NewCreater(h.Config, h.KubeConfig, h.Clientset, h.HubClientset, h.OSSecurityClient, h.RouteClient)
				ip, pvc, updateError, err := hubCreator.CreateHub(&hubv1.Spec)
				if err != nil {
					log.Errorf("unable to create hub for %s due to %+v", hubv1.Name, err)
				}
				hubv1.Status.IP = ip
				hubv1.Status.PVCVolumeName = pvc
				if updateError {
					h.updateState("error", "error", fmt.Sprintf("%+v", err), hubv1)
				} else {
					h.updateState("running", "running", fmt.Sprintf("%+v", err), hubv1)
				}
				h.callHubFederator()
			}
		}
	}

	log.Infof("Done w/ install, starting post-install nanny monitors...")

}

// ObjectDeleted will be called for delete hub events
func (h *HubHandler) ObjectDeleted(name string) {
	log.Debugf("ObjectDeleted: %+v", name)

	hubCreator := hub.NewCreater(h.Config, h.KubeConfig, h.Clientset, h.HubClientset, h.OSSecurityClient, h.RouteClient)
	hubCreator.DeleteHub(name)
	h.callHubFederator()

	//Set spec/state  and status/state to started
	// obj.Spec.State = "deleted"
	// obj.Status.State = "deleted"
	// obj, err = h.updateHubObject(obj)
	// if err != nil {
	// 	log.Errorf("Couldn't update Hub object: %s", err.Error())
	// }
}

// ObjectUpdated will be called for update hub events
func (h *HubHandler) ObjectUpdated(objOld, objNew interface{}) {
	//if strings.Compare(objOld.Spec.State, objNew.Spec.State) != 0 {
	//	log.Infof("%s - Changing state [%s] -> [%s] | Current: [%s]", objNew.Name, objOld.Spec.State, objNew.Spec.State, objNew.Status.State )
	//	// TO DO
	//	objNew.Status.State = objNew.Spec.State
	//	h.hubClientset.SynopsysV1().Hubs(objNew.Namespace).Update(objNew)
	//}
}

func (h *HubHandler) updateState(specState string, statusState string, errorMessage string, hub *hub_v1.Hub) (*hub_v1.Hub, error) {
	hub.Spec.State = specState
	hub.Status.State = statusState
	hub.Status.ErrorMessage = errorMessage
	hub, err := h.updateHubObject(hub)
	if err != nil {
		log.Errorf("couldn't update the state of hub object: %s", err.Error())
	}
	return hub, err
}

func (h *HubHandler) updateHubObject(obj *hub_v1.Hub) (*hub_v1.Hub, error) {
	return h.HubClientset.SynopsysV1().Hubs(h.Namespace).Update(obj)
}

func (h *HubHandler) callHubFederator() {
	// IMPORTANT ! This will block.
	h.CmMutex <- true
	defer func() {
		<-h.CmMutex
	}()
	hubUrls, err := h.getHubUrls()
	log.Debugf("hubUrls: %+v", hubUrls)
	if err != nil {
		log.Errorf("unable to get the hub urls due to %+v", err)
		return
	}
	err = h.addHubFederatorEvents(fmt.Sprintf("%s/sethubs", h.FederatorBaseURL), hubUrls)
	if err != nil {
		log.Errorf("unable to update the hub urls in perceptor due to %+v", err)
		return
	}
}

// HubNamespaces will list the hub namespaces
func (h *HubHandler) getHubUrls() (*APISetHubsRequest, error) {
	// 1. get Hub CDR list from default ns
	hubList, err := util.ListHubs(h.HubClientset, h.Namespace)
	if err != nil {
		return &APISetHubsRequest{}, err
	}

	// 2. extract the namespaces
	hubURLs := []string{}
	for _, hub := range hubList.Items {
		if len(hub.Spec.Namespace) > 0 && strings.EqualFold(hub.Spec.State, "running") {
			hubURL := fmt.Sprintf("webserver.%s.svc", hub.Spec.Namespace)
			status := h.verifyHub(hubURL, hub.Spec.Namespace)
			if status {
				hubURLs = append(hubURLs, hubURL)
			}
		}
	}

	return &APISetHubsRequest{HubURLs: hubURLs}, nil
}

func (h *HubHandler) verifyHub(hubURL string, name string) bool {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 5 * time.Second,
	}

	for i := 0; i < 60; i++ {
		resp, err := client.Get(fmt.Sprintf("https://%s:443/api/current-version", hubURL))
		if err != nil {
			log.Debugf("unable to talk with the hub %s", hubURL)
			time.Sleep(10 * time.Second)
			_, err := util.GetHub(h.HubClientset, h.Namespace, name)
			if err != nil {
				return false
			}
			continue
		}

		_, err = ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		log.Debugf("hub response status for %s is %v", hubURL, resp.Status)

		if resp.StatusCode == 200 {
			return true
		}
		time.Sleep(10 * time.Second)
	}
	return false
}

func (h *HubHandler) addHubFederatorEvents(dest string, obj interface{}) error {
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("unable to serialize %v: %v", obj, err)
	}
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, dest, bytes.NewBuffer(jsonBytes))
	log.Debugf("hub req: %+v", req)
	if err != nil {
		return fmt.Errorf("unable to create the request due to %v", err)
	}
	resp, err := client.Do(req)
	log.Debugf("hub resp: %+v", resp)
	if err != nil {
		return fmt.Errorf("unable to POST to %s: %v", dest, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("http POST request to %s failed with status code %d", dest, resp.StatusCode)
	}
	return nil
}
