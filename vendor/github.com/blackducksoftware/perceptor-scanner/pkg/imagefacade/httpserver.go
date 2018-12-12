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

package imagefacade

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	api "github.com/blackducksoftware/perceptor-scanner/pkg/api"
	common "github.com/blackducksoftware/perceptor-scanner/pkg/common"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// HTTPResponder ...
type HTTPResponder interface {
	PullImage(*common.Image) error
	GetImage(*common.Image) common.ImageStatus
	GetModel() map[string]interface{}
}

// SetupHTTPServer ...
func SetupHTTPServer(responder HTTPResponder) {
	http.HandleFunc("/pullimage", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			recordHTTPRequest("pullimage")
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Errorf("unable to read body for pullimage: %s", err.Error())
				http.Error(w, err.Error(), 400)
				return
			}
			var image *common.Image
			err = json.Unmarshal(body, &image)
			if err != nil {
				log.Errorf("unable to ummarshal JSON for pullimage: %s", err.Error())
				http.Error(w, err.Error(), 400)
				return
			}
			pullError := responder.PullImage(image)

			if pullError == nil {
				log.Debugf("successfully handled pullimage for %s", image.PullSpec)
				fmt.Fprint(w, "")
			} else {
				http.Error(w, pullError.Error(), 503)
			}
		default:
			http.NotFound(w, r)
		}
	})

	http.HandleFunc("/checkimage", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			recordHTTPRequest("checkimage")
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Errorf("unable to read body for getimage: %s", err.Error())
				http.Error(w, err.Error(), 400)
				return
			}
			var image *common.Image
			err = json.Unmarshal(body, &image)
			if err != nil {
				log.Errorf("unable to ummarshal JSON for getimage: %s", err.Error())
				http.Error(w, err.Error(), 400)
				return
			}
			imageStatus := responder.GetImage(image)
			response := api.CheckImageResponse{ImageStatus: imageStatus, PullSpec: image.PullSpec}

			responseBytes, err := json.Marshal(response)
			if err != nil {
				log.Errorf("unable to unmarshal JSON for checkimage: %s", err.Error())
				http.Error(w, err.Error(), 500)
				return
			}

			log.Debugf("successfully handled checkimage for %s: %+v", image.PullSpec, response)
			fmt.Fprintf(w, string(responseBytes))
		default:
			http.NotFound(w, r)
		}
	})

	http.HandleFunc("/model", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			recordHTTPRequest("model")
			jsonBytes, err := json.MarshalIndent(responder.GetModel(), "", "  ")
			if err != nil {
				log.Errorf("unable to marshal JSON for model: %s", err.Error())
				http.Error(w, err.Error(), 500)
				return
			}
			header := w.Header()
			header.Set(http.CanonicalHeaderKey("content-type"), "application/json")
			fmt.Fprint(w, string(jsonBytes))
		default:
			http.NotFound(w, r)
		}
	})

	http.Handle("/metrics", prometheus.Handler())
}
