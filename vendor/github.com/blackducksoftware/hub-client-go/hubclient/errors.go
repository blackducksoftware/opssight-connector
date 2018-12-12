// Copyright 2018 Synopsys, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hubclient

import (
	"github.com/juju/errors"
)

type HubClientError struct {
	errors.Err
	StatusCode int
	HubError   HubResponseError
}

type HubResponseError struct {
	ErrorMessage string                   `json:"errorMessage"`
	Arguments    HubResponseErrorArgument `json:"arguments"`
	Errors       []HubResponseError       `json:"errors"`
	ErrorCode    string                   `json:"errorCode"`
}

type HubResponseErrorArgument struct {
	FieldName    string `json:"fieldname"`
	Type         string `json:"type"`
	Message      string `json:"message"`
	InvalidValue string `json:"invalidValue"`
}

func AnnotateHubClientError(old error, format string) error {
	var hce *HubClientError

	if old == nil {
		return nil
	}

	hce, ok := old.(*HubClientError)
	if !ok {
		hce = &HubClientError{}
	}

	newErr := errors.Annotate(old, format).(*errors.Err)
	err := &HubClientError{*newErr, hce.StatusCode, hce.HubError}
	return err
}

func AnnotateHubClientErrorf(old error, format string, args ...interface{}) error {
	var hce *HubClientError

	if old == nil {
		return nil
	}

	hce, ok := old.(*HubClientError)
	if !ok {
		hce = &HubClientError{}
	}

	newErr := errors.Annotatef(old, format, args...).(*errors.Err)
	err := &HubClientError{*newErr, hce.StatusCode, hce.HubError}
	return err
}

func TraceHubClientError(old error) error {
	var hce *HubClientError

	if old == nil {
		return nil
	}

	hce, ok := old.(*HubClientError)
	if !ok {
		hce = &HubClientError{}
	}

	newErr := errors.Trace(old).(*errors.Err)
	err := &HubClientError{*newErr, hce.StatusCode, hce.HubError}
	return err
}

func HubClientErrorf(format string, args ...interface{}) error {
	newErr := errors.Errorf(format, args).(*errors.Err)
	err := &HubClientError{*newErr, 0, HubResponseError{}}
	return err
}
