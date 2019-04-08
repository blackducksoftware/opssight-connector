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

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
)

// ScanClientInterface ...
type ScanClientInterface interface {
	Scan(scheme string, host string, port int, username string, password string, path string, projectName string, versionName string, scanName string) error
	//ScanCliSh(job ScanJob) error
	//ScanDockerSh(job ScanJob) error
}

// ScanClient implements ScanClientInterface using
// the Black Duck hub and scan client programs.
type ScanClient struct {
	tlsVerification bool
	scanClientInfo  *ScanClientInfo
}

// NewScanClient requires hub login credentials
func NewScanClient(tlsVerification bool) (*ScanClient, error) {
	sc := ScanClient{tlsVerification: tlsVerification}
	return &sc, nil
}

// ensureScanClientIsDownloaded will make sure that the Black Duck scan client is Downloaded for scanning
func (sc *ScanClient) ensureScanClientIsDownloaded(scheme string, host string, port int, username string, password string) error {
	if sc.scanClientInfo != nil {
		return nil
	}
	cliRootPath := "/tmp/scanner"
	scanClientInfo, err := DownloadScanClient(
		OSTypeLinux,
		cliRootPath,
		scheme,
		host,
		username,
		password,
		port,
		time.Duration(300)*time.Second)
	if err != nil {
		return errors.Annotate(err, "unable to download scan client")
	}
	sc.scanClientInfo = scanClientInfo
	return nil
}

// getTLSVerification return the TLS verfiication of the Black Duck host
func (sc *ScanClient) getTLSVerification() string {
	if sc.tlsVerification {
		return ""
	}
	return "--insecure"
}

// Scan executes the Black Duck scan for the input artifact
func (sc *ScanClient) Scan(scheme string, host string, port int, username string, password string, path string, projectName string, versionName string, scanName string) error {
	if err := sc.ensureScanClientIsDownloaded(scheme, host, port, username, password); err != nil {
		return errors.Annotate(err, "cannot run scan cli")
	}
	startTotal := time.Now()

	scanCliImplJarPath := sc.scanClientInfo.ScanCliImplJarPath()
	scanCliJarPath := sc.scanClientInfo.ScanCliJarPath()
	scanCliJavaPath := sc.scanClientInfo.ScanCliJavaPath()
	cmd := exec.Command(scanCliJavaPath,
		"-Xms512m",
		"-Xmx4096m",
		"-Dblackduck.scan.cli.benice=true",
		"-Dblackduck.scan.skipUpdate=true",
		"-Done-jar.silent=true",
		"-Done-jar.jar.path="+scanCliImplJarPath,
		"-jar", scanCliJarPath,
		"--host", host,
		"--port", fmt.Sprintf("%d", port),
		"--scheme", scheme,
		"--project", projectName,
		"--release", versionName,
		"--username", username,
		"--name", scanName,
		sc.getTLSVerification(),
		"-v",
		path)
	log.Infof("running command %+v for path %s\n", cmd, path)
	cmd.Env = append(cmd.Env, fmt.Sprintf("BD_HUB_PASSWORD=%s", password))

	startScanClient := time.Now()
	stdoutStderr, err := cmd.CombinedOutput()

	recordScanClientDuration(time.Now().Sub(startScanClient), err == nil)
	recordTotalScannerDuration(time.Now().Sub(startTotal), err == nil)

	if err != nil {
		recordScannerError("scan client failed")
		log.Errorf("java scanner failed for path %s with error %s and output:\n%s\n", path, err.Error(), string(stdoutStderr))
		return errors.Trace(err)
	}
	log.Infof("successfully completed java scanner for path %s", path)
	log.Debugf("output from path %s: %s", path, stdoutStderr)
	return nil
}

// ScanSh invokes scan.cli.sh
// example:
// 	BD_HUB_PASSWORD=??? ./bin/scan.cli.sh --host ??? --port 443 --scheme https --username sysadmin --insecure --name ??? --release ??? --project ??? ???.tar
func (sc *ScanClient) ScanSh(hubScheme string, host string, port int, username string, password string, path string, projectName string, versionName string, scanName string) error {
	if err := sc.ensureScanClientIsDownloaded(hubScheme, host, port, username, password); err != nil {
		return errors.Annotate(err, "cannot run scan.cli.sh")
	}
	startTotal := time.Now()

	cmd := exec.Command(sc.scanClientInfo.ScanCliShPath(),
		"-Xms512m",
		"-Xmx4096m",
		"-Dblackduck.scan.cli.benice=true",
		"-Dblackduck.scan.skipUpdate=true",
		"-Done-jar.silent=true",
		// "-Done-jar.jar.path="+scanCliImplJarPath,
		// "-jar", scanCliJarPath,
		"--host", host,
		"--port", fmt.Sprintf("%d", port),
		"--scheme", hubScheme,
		"--project", projectName,
		"--release", versionName,
		"--username", username,
		"--name", scanName,
		sc.getTLSVerification(),
		"-v",
		path)
	cmd.Env = append(cmd.Env, fmt.Sprintf("BD_HUB_PASSWORD=%s", password))

	log.Infof("running command %+v for path %s\n", cmd, path)
	startScanClient := time.Now()
	stdoutStderr, err := cmd.CombinedOutput()

	recordScanClientDuration(time.Now().Sub(startScanClient), err == nil)
	recordTotalScannerDuration(time.Now().Sub(startTotal), err == nil)

	if err != nil {
		recordScannerError("scan.cli.sh failed")
		log.Errorf("scan.cli.sh failed for path %s with error %s and output:\n%s\n", path, err.Error(), string(stdoutStderr))
		return errors.Trace(err)
	}
	log.Infof("successfully completed scan.cli.sh for path %s", path)
	log.Debugf("output from path %s: %s", path, stdoutStderr)
	return nil
}
