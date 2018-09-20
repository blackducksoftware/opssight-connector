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

package util

// MapMerge merges two maps together and returns the results.
// If both maps contain the same key, then the value of the
// existing key will be overwritten with the value from the new map
func MapMerge(base map[string]string, new map[string]string) map[string]string {
	newMap := make(map[string]string)
	if base != nil {
		for k, v := range base {
			newMap[k] = v
		}
	}
	for k, v := range new {
		newMap[k] = v
	}
	return newMap
}

// RemoveElement will remove a key from a map if it exists and return
// the map with the key removed
func RemoveElement(data map[string]string, key string) map[string]string {
	if _, exists := data[key]; exists {
		delete(data, key)
	}
	return data
}
