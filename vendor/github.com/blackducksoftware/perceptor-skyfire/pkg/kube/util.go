/*
Copyright (C) 2018 Black Duck Software, Inc.

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

package kube

import (
	"fmt"
	"strconv"
)

func getString(dict map[string]string, key string) (string, error) {
	val, ok := dict[key]
	if !ok {
		return "", fmt.Errorf("key %s not in dictionary", key)
	}
	return val, nil
}

func getInt(dict map[string]string, key string) (int, error) {
	str, err := getString(dict, key)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(str)
}

func CopyMap(dict map[string]string) map[string]string {
	copy := map[string]string{}
	for key, val := range dict {
		copy[key] = val
	}
	return copy
}

func RemoveKeys(dict map[string]string, keys []string) map[string]string {
	copy := CopyMap(dict)
	for _, key := range keys {
		delete(copy, key)
	}
	return copy
}

func HasAllKeys(dict map[string]string, keys []string) bool {
	for _, key := range keys {
		_, ok := dict[key]
		if !ok {
			return false
		}
	}
	return true
}

func HasAnyKeys(dict map[string]string, keys []string) bool {
	for _, key := range keys {
		_, ok := dict[key]
		if ok {
			return true
		}
	}
	return false
}
