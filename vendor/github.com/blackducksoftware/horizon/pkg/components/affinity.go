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

	"github.com/blackducksoftware/horizon/pkg/api"

	"k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func generateNodeSelectorTerm(config api.NodeAffinityConfig) (*v1.NodeSelectorTerm, error) {
	var term v1.NodeSelectorTerm
	var err error

	term.MatchExpressions, err = generateNodeSelectorRequirements(config.Expressions)
	if err != nil {
		return nil, fmt.Errorf("unable to create NodeSelectorTerm MatchExpressions: %+v", err)
	}

	term.MatchFields, err = generateNodeSelectorRequirements(config.Fields)
	if err != nil {
		return nil, fmt.Errorf("unable to create NodeSelectorTerm MatchFields: %+v", err)
	}

	return &term, nil
}

func generateNodeSelectorRequirements(expressions []api.NodeExpression) ([]v1.NodeSelectorRequirement, error) {
	exps := []v1.NodeSelectorRequirement{}

	for _, e := range expressions {
		expression, err := generateNodeSelectorRequirement(e)
		if err != nil {
			return nil, err
		}

		exps = append(exps, *expression)
	}

	return exps, nil
}

func generateNodeSelectorRequirement(e api.NodeExpression) (*v1.NodeSelectorRequirement, error) {
	expression := v1.NodeSelectorRequirement{
		Key:    e.Key,
		Values: e.Values,
	}

	switch e.Op {
	case api.NodeOperatorIn:
		expression.Operator = v1.NodeSelectorOpIn
	case api.NodeOperatorNotIn:
		expression.Operator = v1.NodeSelectorOpNotIn
	case api.NodeOperatorExists:
		expression.Operator = v1.NodeSelectorOpExists
	case api.NodeOperatorDoesNotExist:
		expression.Operator = v1.NodeSelectorOpDoesNotExist
	case api.NodeOperatorGt:
		expression.Operator = v1.NodeSelectorOpGt
	case api.NodeOperatorLt:
		expression.Operator = v1.NodeSelectorOpLt
	default:
		return nil, fmt.Errorf("invalid NodeOperator %d", e.Op)
	}

	return &expression, nil
}

func generatePodAffinityTerm(config api.PodAffinityConfig) v1.PodAffinityTerm {
	var term v1.PodAffinityTerm

	selector := createLabelSelector(config.Selector)
	if !reflect.DeepEqual(selector, metav1.LabelSelector{}) {
		term.LabelSelector = &selector
	}

	term.Namespaces = config.Namespaces
	term.TopologyKey = config.Topology

	return term
}
