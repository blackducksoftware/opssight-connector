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
)

// PersistentVolumeClaim stores the configuration to add or delete the persistent volume claim object
type PersistentVolumeClaim struct {
	config                    *CommonConfig
	deployer                  *util.DeployerHelper
	persistentVolumeClaims    []*components.PersistentVolumeClaim
	oldPersistentVolumeClaims map[string]corev1.PersistentVolumeClaim
	newPersistentVolumeClaims map[string]*corev1.PersistentVolumeClaim
}

// NewPersistentVolumeClaim returns the persistent volume claim
func NewPersistentVolumeClaim(config *CommonConfig, persistentVolumeClaims []*components.PersistentVolumeClaim) (*PersistentVolumeClaim, error) {
	deployer, err := util.NewDeployer(config.kubeConfig)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to get deployer object for %s", config.namespace)
	}
	newPersistentVolumeClaims := append([]*components.PersistentVolumeClaim{}, persistentVolumeClaims...)
	for i := 0; i < len(newPersistentVolumeClaims); i++ {
		if !isLabelsExist(config.expectedLabels, newPersistentVolumeClaims[i].GetObj().Labels) {
			newPersistentVolumeClaims = append(newPersistentVolumeClaims[:i], newPersistentVolumeClaims[i+1:]...)
			i--
		}
	}
	return &PersistentVolumeClaim{
		config:                    config,
		deployer:                  deployer,
		persistentVolumeClaims:    newPersistentVolumeClaims,
		oldPersistentVolumeClaims: make(map[string]corev1.PersistentVolumeClaim, 0),
		newPersistentVolumeClaims: make(map[string]*corev1.PersistentVolumeClaim, 0),
	}, nil
}

// buildNewAndOldObject builds the old and new persistent volume claim
func (r *PersistentVolumeClaim) buildNewAndOldObject() error {
	// build old persistent volume claim
	oldRCs, err := r.list()
	if err != nil {
		return errors.Annotatef(err, "unable to get persistent volume claims for %s", r.config.namespace)
	}
	for _, oldRC := range oldRCs.(*corev1.PersistentVolumeClaimList).Items {
		r.oldPersistentVolumeClaims[oldRC.GetName()] = oldRC
	}

	// build new persistent volume claim
	for _, newRc := range r.persistentVolumeClaims {
		newPersistentVolumeClaimKube, err := newRc.ToKube()
		if err != nil {
			return errors.Annotatef(err, "unable to convert persistent volume claim %s to kube %s", newRc.GetName(), r.config.namespace)
		}
		r.newPersistentVolumeClaims[newRc.GetName()] = newPersistentVolumeClaimKube.(*corev1.PersistentVolumeClaim)
	}

	return nil
}

// add adds the persistent volume claim
func (r *PersistentVolumeClaim) add(isPatched bool) (bool, error) {
	isAdded := false
	for _, persistentVolumeClaim := range r.persistentVolumeClaims {
		if _, ok := r.oldPersistentVolumeClaims[persistentVolumeClaim.GetName()]; !ok {
			r.deployer.Deployer.AddPVC(persistentVolumeClaim)
			isAdded = true
		}
		// else {
		// 	_, err := r.patch(persistentVolumeClaim, isPatched)
		// 	if err != nil {
		// 		return false, errors.Annotatef(err, "patch persistent volume claim:")
		// 	}
		// }
	}
	if isAdded && !r.config.dryRun {
		err := r.deployer.Deployer.Run()
		if err != nil {
			return false, errors.Annotatef(err, "unable to deploy persistent volume claim in %s", r.config.namespace)
		}
	}
	return false, nil
}

// get gets the persistent volume claim
func (r *PersistentVolumeClaim) get(name string) (interface{}, error) {
	return util.GetPVC(r.config.kubeClient, r.config.namespace, name)
}

// list lists all the persistent volume claims
func (r *PersistentVolumeClaim) list() (interface{}, error) {
	return util.ListPVCs(r.config.kubeClient, r.config.namespace, r.config.labelSelector)
}

// delete deletes the persistent volume claim
func (r *PersistentVolumeClaim) delete(name string) error {
	log.Infof("deleting the persistent volume claim %s in %s namespace", name, r.config.namespace)
	return util.DeletePVC(r.config.kubeClient, r.config.namespace, name)
}

// remove removes the persistent volume claim
func (r *PersistentVolumeClaim) remove() error {
	// compare the old and new persistent volume claim and delete if needed
	// for _, oldPersistentVolumeClaim := range r.oldPersistentVolumeClaims {
	// 	if _, ok := r.newPersistentVolumeClaims[oldPersistentVolumeClaim.GetName()]; !ok {
	// 		err := r.delete(oldPersistentVolumeClaim.GetName())
	// 		if err != nil {
	// 			return errors.Annotatef(err, "unable to delete persistent volume claim %s in namespace %s", oldPersistentVolumeClaim.GetName(), r.config.namespace)
	// 		}
	// 	}
	// }
	return nil
}

// patch patches the persistent volume claim
func (r *PersistentVolumeClaim) patch(rc interface{}, isPatched bool) (bool, error) {
	persistentVolumeClaim := rc.(*components.PersistentVolumeClaim)
	persistentVolumeClaimName := persistentVolumeClaim.GetName()
	oldPersistentVolumeClaim := r.oldPersistentVolumeClaims[persistentVolumeClaimName]
	newPersistentVolumeClaim := r.newPersistentVolumeClaims[persistentVolumeClaimName]

	if (!reflect.DeepEqual(oldPersistentVolumeClaim.Spec.StorageClassName, newPersistentVolumeClaim.Spec.StorageClassName) ||
		!reflect.DeepEqual(oldPersistentVolumeClaim.Spec.AccessModes, newPersistentVolumeClaim.Spec.AccessModes) ||
		!reflect.DeepEqual(oldPersistentVolumeClaim.Spec.Resources, newPersistentVolumeClaim.Spec.Resources)) && !r.config.dryRun {
		log.Infof("updating the config map %s in %s namespace", persistentVolumeClaimName, r.config.namespace)
		getPvc, err := r.get(persistentVolumeClaimName)
		if err != nil {
			return false, errors.Annotatef(err, "unable to get the config map %s in namespace %s", persistentVolumeClaimName, r.config.namespace)
		}
		oldLatestPersistentVolumeClaim := getPvc.(*corev1.PersistentVolumeClaim)
		oldLatestPersistentVolumeClaim.Spec.StorageClassName = newPersistentVolumeClaim.Spec.StorageClassName
		oldLatestPersistentVolumeClaim.Spec.AccessModes = newPersistentVolumeClaim.Spec.AccessModes
		oldLatestPersistentVolumeClaim.Spec.Resources = newPersistentVolumeClaim.Spec.Resources
		_, err = util.UpdatePVC(r.config.kubeClient, r.config.namespace, oldLatestPersistentVolumeClaim)
		if err != nil {
			return false, errors.Annotatef(err, "unable to update the config map %s in namespace %s", persistentVolumeClaimName, r.config.namespace)
		}
		return true, nil
	}

	return false, nil
}
