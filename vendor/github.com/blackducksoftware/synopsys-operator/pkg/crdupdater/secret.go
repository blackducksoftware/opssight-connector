/*
Copyright (C) 2019 Synopsys, Inc.

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

package crdupdater

import (
	"reflect"

	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Secret stores the configuration to add or delete the secret object
type Secret struct {
	config     *CommonConfig
	deployer   *util.DeployerHelper
	secrets    []*components.Secret
	oldSecrets map[string]corev1.Secret
	newSecrets map[string]*corev1.Secret
}

// NewSecret returns the secret
func NewSecret(config *CommonConfig, secrets []*components.Secret) (*Secret, error) {
	deployer, err := util.NewDeployer(config.kubeConfig)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to get deployer object for %s", config.namespace)
	}
	newSecrets := append([]*components.Secret{}, secrets...)
	for i := 0; i < len(newSecrets); i++ {
		if !isLabelsExist(config.expectedLabels, newSecrets[i].GetObj().Labels) {
			newSecrets = append(newSecrets[:i], newSecrets[i+1:]...)
			i--
		}
	}
	return &Secret{
		config:     config,
		deployer:   deployer,
		secrets:    newSecrets,
		oldSecrets: make(map[string]corev1.Secret, 0),
		newSecrets: make(map[string]*corev1.Secret, 0),
	}, nil
}

// buildNewAndOldObject builds the old and new secret
func (s *Secret) buildNewAndOldObject() error {
	// build old secret
	oldSecrets, err := s.list()
	if err != nil {
		return errors.Annotatef(err, "unable to get secrets for %s", s.config.namespace)
	}
	for _, oldSecret := range oldSecrets.(*corev1.SecretList).Items {
		s.oldSecrets[oldSecret.GetName()] = oldSecret
	}

	// build new secret
	for _, newSrt := range s.secrets {
		newSecretKube, err := newSrt.ToKube()
		if err != nil {
			return errors.Annotatef(err, "unable to convert secret %s to kube %s", newSrt.GetName(), s.config.namespace)
		}
		s.newSecrets[newSrt.GetName()] = newSecretKube.(*corev1.Secret)
	}

	return nil
}

// add adds the secret
func (s *Secret) add(isPatched bool) (bool, error) {
	isAdded := false
	isUpdated := false
	var err error
	for _, secret := range s.secrets {
		if _, ok := s.oldSecrets[secret.GetName()]; !ok {
			s.deployer.Deployer.AddSecret(secret)
			isAdded = true
		} else {
			isUpdated, err = s.patch(secret, isPatched)
			if err != nil {
				return false, errors.Annotatef(err, "patch secret:")
			}
		}
	}
	if isAdded && !s.config.dryRun {
		err := s.deployer.Deployer.Run()
		if err != nil {
			return false, errors.Annotatef(err, "unable to deploy secret in %s", s.config.namespace)
		}
	}
	return isAdded || isUpdated, nil
}

// get gets the secret
func (s *Secret) get(name string) (interface{}, error) {
	return util.GetSecret(s.config.kubeClient, s.config.namespace, name)
}

// list lists all the secrets
func (s *Secret) list() (interface{}, error) {
	return util.ListSecrets(s.config.kubeClient, s.config.namespace, s.config.labelSelector)
}

// delete deletes the secret
func (s *Secret) delete(name string) error {
	log.Infof("deleting the secret %s in %s namespace", name, s.config.namespace)
	return util.DeleteSecret(s.config.kubeClient, s.config.namespace, name)
}

// remove removes the secret
func (s *Secret) remove() error {
	// compare the old and new secret and delete if needed
	for _, oldSecret := range s.oldSecrets {
		if _, ok := s.newSecrets[oldSecret.GetName()]; !ok {
			err := s.delete(oldSecret.GetName())
			if err != nil {
				return errors.Annotatef(err, "unable to delete secret %s in namespace %s", oldSecret.GetName(), s.config.namespace)
			}
		}
	}
	return nil
}

// patch patches the secret
func (s *Secret) patch(i interface{}, isPatched bool) (bool, error) {
	secret := i.(*components.Secret)
	secretName := secret.GetName()
	oldSecret := s.oldSecrets[secretName]
	newSecret := s.newSecrets[secretName]
	if (!reflect.DeepEqual(newSecret.Data, oldSecret.Data) || !reflect.DeepEqual(newSecret.StringData, oldSecret.StringData)) && !s.config.dryRun {
		log.Infof("updating the secret %s in %s namespace", secretName, s.config.namespace)
		srt, err := s.get(secretName)
		if err != nil {
			return false, errors.Annotatef(err, "unable to get the secret %s in namespace %s", secretName, s.config.namespace)
		}
		oldLatestSecret := srt.(*corev1.Secret)
		oldLatestSecret.Data = newSecret.Data
		oldLatestSecret.StringData = newSecret.StringData
		err = util.UpdateSecret(s.config.kubeClient, s.config.namespace, oldLatestSecret)
		if err != nil {
			return false, errors.Annotatef(err, "unable to update the secret %s in namespace %s", secretName, s.config.namespace)
		}
		return true, nil
	}
	return false, nil
}

// UpdateSecret updates the secret by comparing the old and new secret data
func UpdateSecret(kubeConfig *rest.Config, kubeClient *kubernetes.Clientset, namespace string, secretName string, newConfig *components.Secret) (bool, error) {
	newSecretKube, err := newConfig.ToKube()
	if err != nil {
		return false, errors.Annotatef(err, "unable to convert secret %s to kube in namespace %s", secretName, namespace)
	}
	newSecret := newSecretKube.(*corev1.Secret)
	newSecretData := newSecret.Data
	newSecretStringData := newSecret.StringData

	// getting old secret data
	oldSecret, err := util.GetSecret(kubeClient, namespace, secretName)
	if err != nil {
		// if secret is not present, create the secret
		deployer, err := util.NewDeployer(kubeConfig)
		deployer.Deployer.AddSecret(newConfig)
		err = deployer.Deployer.Run()
		return false, errors.Annotatef(err, "unable to create the secret %s in namespace %s", secretName, namespace)
	}
	oldSecretData := oldSecret.Data
	oldSecretStringData := oldSecret.StringData

	// compare for difference between old and new secret data, if changed update the secret
	if !reflect.DeepEqual(newSecretData, oldSecretData) || !reflect.DeepEqual(newSecretStringData, oldSecretStringData) {
		oldSecret.Data = newSecretData
		oldSecret.StringData = newSecretStringData
		err = util.UpdateSecret(kubeClient, namespace, oldSecret)
		if err != nil {
			return false, errors.Annotatef(err, "unable to update the secret %s in namespace %s", secretName, namespace)
		}
		return true, nil
	}
	return false, nil
}
