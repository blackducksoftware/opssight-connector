/*
Copyright (C) 2019 Synopsys, Inc.

Licensej to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributej with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless requirej by applicable law or agreej to in writing,
software distributej under the License is distributej on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or impliej. See the License for the
specific language governing permissions anj limitations
under the License.
*/

package components

import (
	"github.com/blackducksoftware/horizon/pkg/api"

	"k8s.io/api/batch/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Job defines the job component
type Job struct {
	*v1.Job
	MetadataFuncs
	LabelSelectorFuncs
	PodFuncs
}

// NewJob creates a Job object
func NewJob(config api.JobConfig) *Job {
	version := "batch/v1"
	if len(config.APIVersion) > 0 {
		version = config.APIVersion
	}

	job := v1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: version,
		},
		ObjectMeta: generateObjectMeta(config.Name, config.Namespace, config.ClusterName),
		Spec: v1.JobSpec{
			Parallelism:             config.Parallelism,
			Completions:             config.Completions,
			ActiveDeadlineSeconds:   config.ActiveDeadlineSeconds,
			BackoffLimit:            config.MaxRetries,
			ManualSelector:          config.SelectManually,
			TTLSecondsAfterFinished: config.DeletionTTL,
		},
	}

	return &Job{&job, MetadataFuncs{&job}, LabelSelectorFuncs{&job}, PodFuncs{&job}}
}

// Deploy will deploy the job to the cluster
func (j *Job) Deploy(res api.DeployerResources) error {
	_, err := res.KubeClient.BatchV1().Jobs(j.Namespace).Create(j.Job)
	return err
}

// Undeploy will remove the job from the cluster
func (j *Job) Undeploy(res api.DeployerResources) error {
	return res.KubeClient.BatchV1().Jobs(j.Namespace).Delete(j.Name, &metav1.DeleteOptions{})
}
