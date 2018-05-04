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

package utils

import (
	"reflect"
	"testing"
)

func TestMapMerge(t *testing.T) {
	testcases := []struct {
		description string
		map1        map[string]string
		map2        map[string]string
		resultMap   map[string]string
	}{
		{
			description: "same maps",
			map1:        map[string]string{"key1": "value1", "key2": "value2"},
			map2:        map[string]string{"key1": "value1", "key2": "value2"},
			resultMap:   map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			description: "different maps",
			map1:        map[string]string{"key1": "value1", "key2": "value2"},
			map2:        map[string]string{"key3": "value3", "key4": "value4"},
			resultMap:   map[string]string{"key1": "value1", "key2": "value2", "key3": "value3", "key4": "value4"},
		},
		{
			description: "map with different value",
			map1:        map[string]string{"key1": "value1", "key2": "value2"},
			map2:        map[string]string{"key1": "value1", "key2": "value3"},
			resultMap:   map[string]string{"key1": "value1", "key2": "value3"},
		},
	}

	for _, tc := range testcases {
		result := MapMerge(tc.map1, tc.map2)
		if !reflect.DeepEqual(result, tc.resultMap) {
			t.Errorf("[%s] expected %v got %v: map1 %v, map2 %v", tc.description, tc.resultMap, result, tc.map1, tc.map2)
		}
	}
}
