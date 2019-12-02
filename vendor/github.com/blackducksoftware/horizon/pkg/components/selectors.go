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
	"reflect"

	"github.com/blackducksoftware/horizon/pkg/api"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createLabelSelector(config api.SelectorConfig) metav1.LabelSelector {
	selector := metav1.LabelSelector{
		MatchLabels: config.Labels,
	}

	requirements := []metav1.LabelSelectorRequirement{}
	for _, exp := range config.Expressions {
		requirements = append(requirements, createLabelSelectorRequirement(exp))
	}
	selector.MatchExpressions = requirements

	return selector
}

func createLabelSelectorRequirement(config api.ExpressionRequirementConfig) metav1.LabelSelectorRequirement {
	req := metav1.LabelSelectorRequirement{
		Key:    config.Key,
		Values: config.Values,
	}

	switch config.Op {
	case api.ExpressionRequirementOpIn:
		req.Operator = metav1.LabelSelectorOpIn
	case api.ExpressionRequirementOpNotIn:
		req.Operator = metav1.LabelSelectorOpNotIn
	case api.ExpressionRequirementOpExists:
		req.Operator = metav1.LabelSelectorOpExists
	case api.ExpressionRequirementOpDoesNotExist:
		req.Operator = metav1.LabelSelectorOpDoesNotExist
	}

	return req
}

// LabelSelectorFuncs defines functions to act upon Label Selectors
type LabelSelectorFuncs struct {
	obj interface{}
}

// AddMatchLabelsSelectors adds the given match label selectors to the object
func (l *LabelSelectorFuncs) AddMatchLabelsSelectors(new map[string]string) {
	field := l.getSelectorField()
	if field.IsNil() {
		field.Set(reflect.ValueOf(&metav1.LabelSelector{MatchLabels: make(map[string]string)}))
	}
	if field.Kind() == reflect.Ptr {
		field = field.Elem()
	}
	field = field.FieldByName("MatchLabels")

	for k, v := range new {
		field.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
	}
}

// RemoveMatchLabelsSelectors removes the given match label selectors from the object
func (l *LabelSelectorFuncs) RemoveMatchLabelsSelectors(remove []string) {
	field := l.getSelectorField()
	if field.Kind() == reflect.Ptr {
		field = field.Elem()
	}

	if field.IsValid() {
		field = field.FieldByName("MatchLabels")
		old := field.Interface().(map[string]string)
		for _, k := range remove {
			delete(old, k)
		}
		field.Set(reflect.ValueOf(old))
	}
}

// AddMatchExpressionsSelector will add match expressions selectors to the object
func (l *LabelSelectorFuncs) AddMatchExpressionsSelector(add api.ExpressionRequirementConfig) {
	field := l.getSelectorField()
	if field.IsNil() {
		field.Set(reflect.ValueOf(&metav1.LabelSelector{}))
	}
	if field.Kind() == reflect.Ptr {
		field = field.Elem()
	}
	field = field.FieldByName("MatchExpressions")

	fieldValue := field.Interface().([]metav1.LabelSelectorRequirement)
	fieldValue = append(fieldValue, createLabelSelectorRequirement(add))
	field.Set(reflect.ValueOf(fieldValue))
}

// RemoveMatchExpressionsSelector removes the match expressions selector from the object
func (l *LabelSelectorFuncs) RemoveMatchExpressionsSelector(remove api.ExpressionRequirementConfig) {
	field := l.getSelectorField()
	if field.Kind() == reflect.Ptr {
		field = field.Elem()
	}

	if field.IsValid() {
		exp := createLabelSelectorRequirement(remove)
		field = field.FieldByName("MatchExpressions")
		currentExp := field.Interface().([]metav1.LabelSelectorRequirement)
		for i, e := range currentExp {
			if reflect.DeepEqual(exp, e) {
				currentExp = append(currentExp[:i], currentExp[i+1:]...)
				break
			}
		}
		field.Set(reflect.ValueOf(currentExp))
	}
}

func (l *LabelSelectorFuncs) getSelectorField() reflect.Value {
	field := reflect.ValueOf(l.obj).Elem().FieldByName("Spec").FieldByName("Selector")
	return field
}
