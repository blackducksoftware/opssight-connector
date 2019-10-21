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
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	opssightapi "github.com/blackducksoftware/synopsys-operator/pkg/api/opssight/v1"
	opssightclientset "github.com/blackducksoftware/synopsys-operator/pkg/opssight/client/clientset/versioned"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// AuthToken will store the providers auth token
type AuthToken struct {
	AccessToken string
	Endpoint    string
}

// TokenGenerator will store the token provider function
type TokenGenerator struct {
	TokenGenFxn func() ([]AuthToken, error)
	Name        string
}

type controller struct {
	kubeConfig *rest.Config
	kubeClient *kubernetes.Clientset
	ecrClient  ecrInterface
	gcrClient  gcrInterface
}

// RegistryAuth ...
type RegistryAuth struct {
	URL      string
	User     string
	Password string
	Token    string
}

// getGCRAuthorizationKey will get the authorization key from Google Container Registry (GCR)
func (c *controller) getGCRAuthorizationKey() ([]AuthToken, error) {
	ts, err := c.gcrClient.DefaultTokenSource(context.TODO(), "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return []AuthToken{}, err
	}

	token, err := ts.Token()
	if err != nil {
		return []AuthToken{}, err
	}

	if !token.Valid() {
		return []AuthToken{}, fmt.Errorf("token was invalid")
	}

	if token.Type() != "Bearer" {
		return []AuthToken{}, fmt.Errorf(fmt.Sprintf("expected token type \"Bearer\" but got \"%s\"", token.Type()))
	}

	return []AuthToken{
		AuthToken{
			AccessToken: token.AccessToken,
			Endpoint:    *argGCRURL},
	}, nil
}

// getECRAuthorizationKey will get the authorization key from Elastic Container Registry (ECR)
func (c *controller) getECRAuthorizationKey() ([]AuthToken, error) {

	if len(awsAccountIDs) == 0 {
		return []AuthToken{}, fmt.Errorf("unable to get auth token for ECR account due to missing account id's")
	}

	var tokens []AuthToken
	var regIds []*string
	regIds = make([]*string, len(awsAccountIDs))

	for i, awsAccountID := range awsAccountIDs {
		regIds[i] = aws.String(awsAccountID)
	}

	sess := session.Must(session.NewSession())
	awsConfig := aws.NewConfig().WithRegion(*argAWSRegion).WithCredentialsChainVerboseErrors(true)

	if *argAWSAssumeRole != "" {
		creds := stscreds.NewCredentials(sess, *argAWSAssumeRole)
		awsConfig.Credentials = creds
	}

	svc := ecr.New(sess, awsConfig)

	params := &ecr.GetAuthorizationTokenInput{
		RegistryIds: regIds,
	}

	resp, err := svc.GetAuthorizationToken(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		log.Errorf("unable to get ECR auth token because %+v", err)
		return []AuthToken{}, err
	}

	for _, auth := range resp.AuthorizationData {
		tokens = append(tokens, AuthToken{
			AccessToken: *auth.AuthorizationToken,
			Endpoint:    *auth.ProxyEndpoint,
		})

	}
	return tokens, nil
}

func (c *controller) getTokenGenerators() []TokenGenerator {
	tokenGenerators := []TokenGenerator{}

	tokenGenerators = append(tokenGenerators, TokenGenerator{
		TokenGenFxn: c.getGCRAuthorizationKey,
		Name:        "GCR",
	})

	tokenGenerators = append(tokenGenerators, TokenGenerator{
		TokenGenFxn: c.getECRAuthorizationKey,
		Name:        "ECR",
	})

	return tokenGenerators
}

func (c *controller) getTokenEndpoint(tokenEndPoint string) string {
	if strings.HasPrefix(tokenEndPoint, "https://") {
		i := strings.Index(tokenEndPoint, "https://")
		return tokenEndPoint[i+8:]
	} else if strings.HasPrefix(tokenEndPoint, "http://") {
		i := strings.Index(tokenEndPoint, "http://")
		return tokenEndPoint[i+7:]
	}
	return tokenEndPoint
}

func (c *controller) getTokenPassword(tokenEndPoint string, tokenAccessToken string, tokenType string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(tokenAccessToken)
	if err != nil {
		return "", fmt.Errorf("unable to decode the password for token endpoint: %s", tokenEndPoint)
	}
	strdata := string(data)
	if strings.EqualFold(tokenType, "ECR") {
		return strdata[strings.Index(strdata, ":")+1:], nil
	}
	return strdata, nil
}

func (c *controller) updateOpsSightWithAuthToken(opssightClient *opssightclientset.Clientset, opssight *opssightapi.OpsSight, tokens []AuthToken, tokenType string) error {
	for _, token := range tokens {
		endPoint := c.getTokenEndpoint(token.Endpoint)
		for i := range opssight.Spec.ScannerPod.ImageFacade.InternalRegistries {
			if strings.EqualFold(endPoint, opssight.Spec.ScannerPod.ImageFacade.InternalRegistries[i].URL) {
				log.Infof("found %s url in opssight private registries", endPoint)
				strdata, err := c.getTokenPassword(endPoint, token.AccessToken, tokenType)
				if err != nil {
					log.Errorf("unable to get the token password for %s because %+v", endPoint, err)
					continue
				}
				opssight.Spec.ScannerPod.ImageFacade.InternalRegistries[i].Password = strdata
			}
		}
	}

	// AWS EKS specific - if the registry url that is present in OpsSight Configmap but not in AWS token list,
	// then update the token with the 1st endpoint's password. This is because the EKS kube pods are from
	// different account id than the used one

	if strings.EqualFold(tokenType, "ECR") {
		for i := range opssight.Spec.ScannerPod.ImageFacade.InternalRegistries {
			isRegistryExists := false
			for _, token := range tokens {
				endPoint := c.getTokenEndpoint(token.Endpoint)

				if strings.EqualFold(endPoint, opssight.Spec.ScannerPod.ImageFacade.InternalRegistries[i].URL) {
					isRegistryExists = true
					break
				}
			}

			if !isRegistryExists && len(tokens) > 0 && strings.HasSuffix(opssight.Spec.ScannerPod.ImageFacade.InternalRegistries[i].URL, "amazonaws.com") {
				strdata, err := c.getTokenPassword(tokens[0].Endpoint, tokens[0].AccessToken, tokenType)
				if err != nil {
					log.Errorf("unable to get the token password for %s because %+v", tokens[0].Endpoint, err)
					continue
				}
				opssight.Spec.ScannerPod.ImageFacade.InternalRegistries[i].Password = strdata
			}
		}
	}

	_, err := util.UpdateOpsSight(opssightClient, "", opssight)
	return err
}

func (c *controller) updateAuthTokens() error {
	log.Print("Refreshing credentials...")
	tokenGenerators := c.getTokenGenerators()
	opsSightClient, err := opssightclientset.NewForConfig(c.kubeConfig)
	if err != nil {
		return fmt.Errorf("error in creating the opssight client due to %+v", err)
	}
	opssights, err := util.ListOpsSights(opsSightClient, "")
	if err != nil {
		return fmt.Errorf("error in getting the list of opssight due to %+v", err)
	}
	for _, tokenGenerator := range tokenGenerators {
		newTokens, err := tokenGenerator.TokenGenFxn()
		if err != nil {
			log.Infof("error getting tokens for provider %s. Skipping cloud provider! [Err: %s]", tokenGenerator.Name, err)
			continue
		}
		for _, opssight := range opssights.Items {
			log.Debugf("new tokens for %s provider is %+v", tokenGenerator.Name, newTokens)
			err = c.updateOpsSightWithAuthToken(opsSightClient, &opssight, newTokens, tokenGenerator.Name)
			if err != nil {
				log.Error(err)
				continue
			}
		}
	}
	return nil
}
