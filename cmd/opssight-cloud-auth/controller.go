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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/blackducksoftware/synopsys-operator/pkg/opssight"
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

	// log.Printf("request: %+v, output: %+v", req, out)
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

func (c *controller) updateConfigMapWithAuthToken(tokens []AuthToken, namespace string, name string) error {
	cm, err := util.GetConfigMap(c.kubeClient, namespace, name)
	if err != nil {
		return fmt.Errorf("unable to get the %s config map in %s namespace", name, namespace)
	}

	opsssightData := opssight.MainOpssightConfigMap{}

	err = json.Unmarshal([]byte(cm.Data["opssight.json"]), &opsssightData)
	if err != nil {
		return fmt.Errorf("unable to unmarshal config map json: %s", err.Error())
	}

	for _, token := range tokens {
		var endPoint string
		if strings.HasPrefix(token.Endpoint, "https://") {
			i := strings.Index(token.Endpoint, "https://")
			endPoint = token.Endpoint[i+8:]
		} else if strings.HasPrefix(token.Endpoint, "http://") {
			i := strings.Index(token.Endpoint, "http://")
			endPoint = token.Endpoint[i+7:]
		} else {
			endPoint = token.Endpoint
		}

		for i := range opsssightData.ImageFacade.PrivateDockerRegistries {
			if strings.EqualFold(endPoint, opsssightData.ImageFacade.PrivateDockerRegistries[i].URL) {
				log.Infof("Inside for endPoint: %s and token: %s", endPoint, token.AccessToken)
				data, err := base64.StdEncoding.DecodeString(token.AccessToken)
				if err != nil {
					log.Errorf("unable to decode the password for token endpoint: %s", token.Endpoint)
				}
				strdata := string(data)
				opsssightData.ImageFacade.PrivateDockerRegistries[i].Password = strdata[strings.Index(strdata, ":")+1:]
			}
		}
	}

	log.Infof("opsssightData: %+v", opsssightData)
	jsonBytes, err := json.Marshal(opsssightData)
	if err != nil {
		return fmt.Errorf("unable to marshal json: %s", err.Error())
	}

	cm.Data["opssight.json"] = string(jsonBytes)

	err = util.UpdateConfigMap(c.kubeClient, namespace, cm)
	if err != nil {
		return fmt.Errorf("unable to update the %s config map in %s namespace", name, namespace)
	}
	log.Infof("updated the %s config map in %s namespace successfully", name, namespace)
	return nil
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
			log.Printf("Error getting tokens for provider %s. Skipping cloud provider! [Err: %s]", tokenGenerator.Name, err)
			continue
		}
		for _, opssight := range opssights.Items {
			log.Debugf("new tokens for %s provider is %+v", tokenGenerator.Name, newTokens)
			err = c.updateConfigMapWithAuthToken(newTokens, opssight.Spec.Namespace, "opssight")
			if err != nil {
				log.Error(err)
			}
		}
	}
	return nil
}
