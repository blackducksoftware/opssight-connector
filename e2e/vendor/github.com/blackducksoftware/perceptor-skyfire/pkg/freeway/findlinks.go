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

package freeway

import (
	"fmt"
	"reflect"
)

//
// func FindLinksRestricted(obj interface{}) ([]string, []error) {
// 	links := []string{}
// 	errors := []error{}
// 	switch obj.(type) {
// 		case hubapi
// 	}
// }

func FindLinks(obj interface{}) ([]string, []error) {
	switch v := obj.(type) {
	case map[string]interface{}:
		return FindLinksDict(v)
	case []interface{}:
		return FindLinksArray(v)
	default:
		return []string{}, []error{}
	}
}

func FindLinksDict(dict map[string]interface{}) ([]string, []error) {
	links := []string{}
	errors := []error{}
	for key, val := range dict {
		if key == "links" {
			linkObjs, ok := val.([]interface{})
			if ok {
				for _, obj := range linkObjs {
					objDict, ok := obj.(map[string]interface{})
					if ok {
						linkValue, ok := objDict["href"].(string)
						if ok {
							links = append(links, linkValue)
						} else {
							errors = append(errors, fmt.Errorf("invalid type of href: %s", reflect.TypeOf(objDict["href"])))
						}
					} else {
						errors = append(errors, fmt.Errorf("invalid type of links object: %s", reflect.TypeOf(obj)))
					}
				}
			} else {
				errors = append(errors, fmt.Errorf("invalid type of links: %s", reflect.TypeOf(val)))
			}
			continue
		}
		switch v := val.(type) {
		case []interface{}:
			arrayLinks, errs := FindLinksArray(v)
			links = append(links, arrayLinks...)
			errors = append(errors, errs...)
		case map[string]interface{}:
			dictLinks, errs := FindLinksDict(v)
			links = append(links, dictLinks...)
			errors = append(errors, errs...)
		default:
			break
		}
	}
	return links, errors
}

func FindLinksArray(array []interface{}) ([]string, []error) {
	links := []string{}
	errors := []error{}
	for _, item := range array {
		itemLinks, errs := FindLinks(item)
		links = append(links, itemLinks...)
		errors = append(errors, errs...)
	}
	return links, errors
}
