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

package components

import (
	"fmt"
	"reflect"
	"strings"

	"k8s.io/api/core/v1"
)

// PodFuncs defines common functions for managing pods contained in other components
type PodFuncs struct {
	obj interface{}
}

// AddPod adds a pod to the object
func (pf *PodFuncs) AddPod(pod *Pod) {
	if pod != nil {
		template := v1.PodTemplateSpec{
			ObjectMeta: pod.ObjectMeta,
			Spec:       pod.Spec,
		}

		field := reflect.ValueOf(pf.obj).Elem().FieldByName("Spec").FieldByName("Template")
		if field.Kind() == reflect.Ptr {
			field.Set(reflect.ValueOf(&template))
		} else {
			field.Set(reflect.ValueOf(template))
		}
	}
}

// RemovePod removes a pod from the object
func (pf *PodFuncs) RemovePod(name string) error {
	field := reflect.ValueOf(pf.obj).Elem().FieldByName("Spec").FieldByName("Template")
	specName := field.FieldByName("Name").Interface().(string)
	if !strings.EqualFold(specName, name) {
		return fmt.Errorf("pod with name %s doesn't exist on %s", name, reflect.TypeOf(pf.obj))
	}

	if field.Kind() == reflect.Ptr {
		field.Set(reflect.ValueOf(&v1.PodTemplateSpec{}))
	} else {
		field.Set(reflect.ValueOf(v1.PodTemplateSpec{}))
	}
	return nil
}
