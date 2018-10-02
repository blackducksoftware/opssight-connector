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

package hub

import (
	"bytes"
	"io"
	"strings"

	"github.com/blackducksoftware/perceptor-protoform/pkg/util"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

func (hc *Creater) execContainer(request *rest.Request, command []string) error {
	var stdin io.Reader
	stdin = util.NewStringReader(command)

	log.Debugf("Request URL: %+v, request: %+v, command: %s", request.URL().String(), request, strings.Join(command, ""))

	exec, err := remotecommand.NewSPDYExecutor(hc.KubeConfig, "POST", request.URL())
	log.Debugf("exec: %+v, error: %+v", exec, err)
	if err != nil {
		log.Errorf("error while creating Executor: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})

	log.Debugf("stdout: %s, stderr: %s", stdout.String(), stderr.String())
	return err
}
