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

package docker

// TODO start using this code (need to clean it up a bit first, of course)
//   instead of the http stuff over a socket in order to hit the docker daemon
// import (
// 	"encoding/base64"
// 	"encoding/json"
// 	"io/ioutil"
// 	"time"
//
// 	"github.com/docker/docker/api/types"
// 	dockerclient "github.com/docker/docker/client"
// 	log "github.com/sirupsen/logrus"
// 	// "golang.org/x/net/context"
// 	"context"
// )
//
// func dockerEncodeAuthHeader(username string, password string) (string, error) {
// 	authConfig := types.AuthConfig{
// 		Username: username,
// 		Password: password,
// 	}
// 	encodedJSON, err := json.Marshal(authConfig)
// 	if err != nil {
// 		return "", err
// 	}
// 	return base64.URLEncoding.EncodeToString(encodedJSON), nil
// }
//
// func pullImage(username string, password string, image Image) error {
// 	ctx := context.Background()
// 	client, err := dockerclient.NewEnvClient()
// 	if err != nil {
// 		return err
// 	}
// 	//	client.ClientVersion()
// 	dockerclient.WithVersion("1.24")(client)
// 	version, err := client.ServerVersion(ctx)
// 	if err != nil {
// 		panic(err)
// 	}
// 	log.Infof("server version: %+v , APIVersion %s, Version string %s", version, version.APIVersion, version.Version)
// 	// TODO this is so weird -- to have to set the version twice (once to anything so you
// 	//   can use the client to get the version, once to set it to the right version)
// 	dockerclient.WithVersion(version.APIVersion)(client)
//
// 	authString, err := dockerEncodeAuthHeader(username, password)
//
// 	if err != nil {
// 		log.Errorf("unable to encode authentication: %s", err.Error())
// 		return err
// 	}
//
// 	log.Infof("auth string: %s", authString)
//
// 	out, err := client.ImagePull(ctx, image.DockerPullSpec(), types.ImagePullOptions{RegistryAuth: authString})
// 	if err != nil {
// 		return err
// 	}
//
// 	defer out.Close()
// 	bodyBytes, err := ioutil.ReadAll(out)
// 	if err != nil {
// 		recordDockerError(createStage, "unable to read POST response body", image, err)
// 		log.Errorf("unable to read response body for %s: %s", image.DockerPullSpec(), err.Error())
// 		return err
// 	}
// 	log.Infof("body of POST response from %s: %s", image.DockerPullSpec(), string(bodyBytes))
//
// 	return nil
// }
//
// func (ip *ImagePuller) DockerCreateImageInLocalDocker(image Image) error {
// 	start := time.Now()
// 	log.Infof("Attempting to create %s ......", image.DockerPullSpec())
//
// 	// TODO if the image *isn't* from the local registry, then don't do this auth stuff
//
// 	err := pullImage(ip.dockerUser, ip.dockerPassword, image)
//
// 	if err != nil {
// 		log.Errorf("unable to pull docker image %s: %s", image.DockerPullSpec(), err.Error())
// 	} else {
// 		log.Infof("successfully pulled docker image %s", image.DockerPullSpec())
// 	}
//
// 	recordDockerGetDuration(time.Now().Sub(start))
//
// 	return err
// }
