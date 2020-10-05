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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	refreshInMinutes  = "REFRESH_IN_MINUTES"
	awsAssumeRole     = "AWS_ASSUME_ROLE"
	awsRegion         = "AWS_REGION"
	awsAccountIds     = "AWS_ACCOUNT_IDS"
	gcrURL            = "GCR_URL"
	opsSightName      = "NAME"
	opsSightNamespace = "NAMESPACE"
)

var (
	awsAccountIDs []string
)

// GCR interface
type gcrInterface interface {
	DefaultTokenSource(ctx context.Context, scope ...string) (oauth2.TokenSource, error)
}

type gcrClient struct{}

func (gcr gcrClient) DefaultTokenSource(ctx context.Context, scope ...string) (oauth2.TokenSource, error) {
	return google.DefaultTokenSource(ctx, scope...)
}

func newGcrClient() gcrInterface {
	return gcrClient{}
}

// ECR interface
type ecrInterface interface {
	GetAuthorizationToken(input *ecr.GetAuthorizationTokenInput) (*ecr.GetAuthorizationTokenOutput, error)
}

func newEcrClient() ecrInterface {
	sess := session.Must(session.NewSession())
	awsConfig := aws.NewConfig().WithRegion(*argAWSRegion)

	if *argAWSAssumeRole != "" {
		creds := stscreds.NewCredentials(sess, *argAWSAssumeRole)
		awsConfig.Credentials = creds
	}

	return ecr.New(sess, awsConfig)
}
