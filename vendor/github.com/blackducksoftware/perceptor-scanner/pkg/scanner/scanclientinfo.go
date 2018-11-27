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

package scanner

import "fmt"

// ScanClientInfo ...
type ScanClientInfo struct {
	HubVersion string
	RootPath   string
	OSType     OSType
}

// NewScanClientInfo ...
func NewScanClientInfo(hubVersion string, rootPath string, osType OSType) *ScanClientInfo {
	return &ScanClientInfo{HubVersion: hubVersion, RootPath: rootPath, OSType: osType}
}

// ScanCliZipPath ...
func (sci *ScanClientInfo) ScanCliZipPath() string {
	return fmt.Sprintf("%s/scanclient.zip", sci.RootPath)
}

// ScanCliShPath ...
func (sci *ScanClientInfo) ScanCliShPath() string {
	return fmt.Sprintf("%s/scan.cli-%s/bin/scan.cli.sh", sci.RootPath, sci.HubVersion)
}

// ScanCliImplJarPath ...
func (sci *ScanClientInfo) ScanCliImplJarPath() string {
	return fmt.Sprintf("%s/scan.cli-%s/lib/cache/scan.cli.impl-standalone.jar", sci.RootPath, sci.HubVersion)
}

// ScanCliJarPath ...
func (sci *ScanClientInfo) ScanCliJarPath() string {
	return fmt.Sprintf("%s/scan.cli-%s/lib/scan.cli-%s-standalone.jar", sci.RootPath, sci.HubVersion, sci.HubVersion)
}

// ScanCliJavaPath ...
func (sci *ScanClientInfo) ScanCliJavaPath() string {
	switch sci.OSType {
	case OSTypeLinux:
		return fmt.Sprintf("%s/scan.cli-%s/jre/bin/java", sci.RootPath, sci.HubVersion)
	case OSTypeMac:
		return fmt.Sprintf("%s/scan.cli-%s/jre/Contents/Home/bin/java", sci.RootPath, sci.HubVersion)
	}
	panic(fmt.Errorf("invalid os type: %d", sci.OSType))
}
