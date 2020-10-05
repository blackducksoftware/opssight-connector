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

package main

import (
	"flag"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/blackducksoftware/synopsysctl/dev-tests/testutils"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

var (
	flags             = flag.NewFlagSet("", flag.ContinueOnError)
	argGCRURL         = flags.String(gcrURL, "https://gcr.io", `Default GCR URL`)
	argAWSRegion      = flags.String(awsRegion, "us-east-1", `Default AWS region`)
	argAWSAssumeRole  = flags.String(awsAssumeRole, "", `If specified AWS will assume this role and use it to retrieve tokens`)
	argRefreshMinutes = flags.Int(refreshInMinutes, 60, `Default time to wait before refreshing (60 minutes)`)
	namespace         = flags.String(opsSightNamespace, "", `OpsSight namespace`)
	name              = flags.String(opsSightName, "", `OpsSight name`)
)

func main() {
	log.Print("Starting up...")

	flags.Parse(os.Args)

	validateParams()

	log.Infof("Using GCR URL: %s", *argGCRURL)
	log.Infof("Using AWS Account: %s", strings.Join(awsAccountIDs, ","))
	log.Infof("Using AWS Region: %s", *argAWSRegion)
	log.Infof("Using AWS Assume Role: %s", *argAWSAssumeRole)
	log.Infof("Refresh Interval (minutes): %d", *argRefreshMinutes)
	log.Infof("OpsSight namespace: %s", *namespace)
	log.Infof("OpsSight name: %s", *name)

	kubeConfig, err := testutils.GetKubeConfig("", false)
	if err != nil {
		log.Errorf("unable to create config for both in-cluster and external to cluster due to %+v", err)
		os.Exit(1)
	}

	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		log.Errorf("unable to create kubernetes client due to %+v", err)
		os.Exit(1)
	}

	c := &controller{kubeConfig, kubeClient, newEcrClient(), newGcrClient()}

	for {
		err = c.updateAuthTokens()
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
		time.Sleep(time.Duration(*argRefreshMinutes) * time.Minute)
	}
}

func validateParams() {
	// Allow environment variables to overwrite args
	awsAccountIDEnv := os.Getenv(awsAccountIds)
	awsRegionEnv := os.Getenv(awsRegion)
	argAWSAssumeRoleEnv := os.Getenv(awsAssumeRole)
	argRefreshMinutesEnv := os.Getenv(refreshInMinutes)
	gcrURLEnv := os.Getenv(gcrURL)
	releaseNamespace := os.Getenv(opsSightNamespace)
	releaseName := os.Getenv(opsSightName)

	if len(awsRegionEnv) > 0 {
		argAWSRegion = &awsRegionEnv
	}

	if len(awsAccountIDEnv) > 0 {
		awsAccountIDs = strings.Split(awsAccountIDEnv, ",")
	} else {
		awsAccountIDs = []string{""}
	}

	if len(gcrURLEnv) > 0 {
		argGCRURL = &gcrURLEnv
	}

	if len(argAWSAssumeRoleEnv) > 0 {
		argAWSAssumeRole = &argAWSAssumeRoleEnv
	}

	if len(argRefreshMinutesEnv) > 0 {
		refreshInterval, _ := strconv.Atoi(argRefreshMinutesEnv)
		argRefreshMinutes = &refreshInterval
	}

	if len(releaseNamespace) > 0 {
		namespace = &releaseNamespace
	} else {
		log.Fatal("OpsSight release namespace is mandatory!")
	}

	if len(releaseName) > 0 {
		name = &releaseName
	} else {
		log.Fatal("OpsSight release name is mandatory!")
	}
}
