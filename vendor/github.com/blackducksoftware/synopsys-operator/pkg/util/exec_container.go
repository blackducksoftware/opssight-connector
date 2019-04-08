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

package util

import (
	"bytes"
	"io"
	"strings"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// CreateExecContainerRequest will create the request to exec into kubernetes pod
func CreateExecContainerRequest(clientset *kubernetes.Clientset, pod *corev1.Pod, command string) *rest.Request {
	return clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec").
		Param("container", pod.Spec.Containers[0].Name).
		VersionedParams(&corev1.PodExecOptions{
			Container: pod.Spec.Containers[0].Name,
			Command:   []string{command},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)
}

// ExecContainer will exec into the container and run the commands provided in the input
func ExecContainer(kubeConfig *rest.Config, request *rest.Request, command []string) (string, error) {
	var stdin io.Reader
	stdin = NewStringReader(command)

	log.Debugf("Request URL: %+v, request: %+v, command: %s", request.URL().String(), request, strings.Join(command, ""))

	exec, err := remotecommand.NewSPDYExecutor(kubeConfig, "POST", request.URL())
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
	return stdout.String(), err
}
