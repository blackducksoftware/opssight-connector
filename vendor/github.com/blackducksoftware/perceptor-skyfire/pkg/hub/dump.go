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

package hub

// Dump .....
type Dump struct {
	Version       string
	Projects      []*Project
	Scans         []*CodeLocation
	ScansByName   map[string]*CodeLocation
	DuplicateShas map[string]bool
}

// NewDump .....
func NewDump(version string, projects []*Project) *Dump {
	dump := &Dump{
		Version:       version,
		Projects:      projects,
		Scans:         []*CodeLocation{},
		ScansByName:   map[string]*CodeLocation{},
		DuplicateShas: map[string]bool{}}
	dump.computeDerivedData()
	return dump
}

func (hd *Dump) computeDerivedData() {
	for _, project := range hd.Projects {
		for _, version := range project.Versions {
			for _, scan := range version.CodeLocations {
				hd.Scans = append(hd.Scans, scan)
				_, ok := hd.ScansByName[scan.Name]
				if !ok {
					hd.ScansByName[scan.Name] = scan
				} else {
					hd.DuplicateShas[scan.Name] = true
				}
			}
		}
	}
}
