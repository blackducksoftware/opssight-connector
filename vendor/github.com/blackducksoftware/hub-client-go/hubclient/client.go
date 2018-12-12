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
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
)

type HubClientDebug uint16

const (
	HubClientDebugTimings HubClientDebug = 1 << iota
	HubClientDebugContent
)

// Client will need to support CSRF tokens for session-based auth for Hub 4.1.x (or was it 4.0?)
type Client struct {
	httpClient    *http.Client
	baseURL       string
	authToken     string
	useAuthToken  bool
	haveCsrfToken bool
	csrfToken     string
	debugFlags    HubClientDebug
}

func NewWithSession(baseURL string, debugFlags HubClientDebug, timeout time.Duration) (*Client, error) {

	jar, err := cookiejar.New(nil) // Look more at this function

	if err != nil {
		return nil, AnnotateHubClientError(err, "unable to instantiate cookie jar")
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Jar:       jar,
		Transport: tr,
		Timeout:   timeout,
	}

	return &Client{
		httpClient:   client,
		baseURL:      baseURL,
		useAuthToken: false,
		debugFlags:   debugFlags,
	}, nil
}

func NewWithToken(baseURL string, authToken string, debugFlags HubClientDebug, timeout time.Duration) (*Client, error) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}

	return &Client{
		httpClient:   client,
		baseURL:      baseURL,
		authToken:    authToken,
		useAuthToken: true,
		debugFlags:   debugFlags,
	}, nil
}

func (c *Client) BaseURL() string {
	return c.baseURL
}

func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

func readBytes(readCloser io.ReadCloser) ([]byte, error) {

	defer readCloser.Close()
	buf := new(bytes.Buffer)

	if _, err := buf.ReadFrom(readCloser); err != nil {
		return nil, TraceHubClientError(err)
	}

	return buf.Bytes(), nil
}

func validateHTTPResponse(resp *http.Response, expectedStatusCode int) error {

	if resp.StatusCode != expectedStatusCode { // Should this be a list at some point?
		body := readResponseBody(resp)
		return newHubClientError(body, resp, fmt.Sprintf("got a %d response instead of a %d", resp.StatusCode, expectedStatusCode))
	}

	return nil
}

func (c *Client) processResponse(resp *http.Response, result interface{}, expectedStatusCode int) error {

	var bodyBytes []byte

	if err := validateHTTPResponse(resp, expectedStatusCode); err != nil {
		return AnnotateHubClientError(err, "error validating HTTP Response")
	}

	if result == nil {
		// Don't have a result to deserialize to, skip it
		return nil
	}

	bodyBytes, err := readBytes(resp.Body)
	if err != nil {
		return newHubClientError(bodyBytes, resp, fmt.Sprintf("error reading HTTP Response: %+v", err))
	}

	if c.debugFlags&HubClientDebugContent != 0 {
		log.Debugf("START DEBUG: --------------------------------------------------------------------------- \n\n")
		log.Debugf("TEXT OF RESPONSE: \n %s", string(bodyBytes[:]))
		log.Debugf("END DEBUG: --------------------------------------------------------------------------- \n\n\n\n")
	}

	if err := json.Unmarshal(bodyBytes, result); err != nil {
		return newHubClientError(bodyBytes, resp, fmt.Sprintf("error parsing HTTP Response: %+v", err))
	}

	return nil
}

func (c *Client) HttpGetJSON(url string, result interface{}, expectedStatusCode int) error {

	// TODO: Content type?

	var resp *http.Response

	if c.debugFlags&HubClientDebugTimings != 0 {
		log.Debugf("DEBUG HTTP STARTING GET REQUEST: %s", url)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return newHubClientError(nil, nil, fmt.Sprintf("error creating http get request for %s: %+v", url, err))
	}

	c.doPreRequest(req)

	httpStart := time.Now()
	if resp, err = c.httpClient.Do(req); err != nil {
		body := readResponseBody(resp)
		return newHubClientError(body, resp, fmt.Sprintf("error getting HTTP Response from %s: %+v", url, err))
	}

	httpElapsed := time.Since(httpStart)

	if c.debugFlags&HubClientDebugTimings != 0 {
		log.Debugf("DEBUG HTTP GET ELAPSED TIME: %d ms.   -- Request: %s", (httpElapsed / 1000 / 1000), url)
	}

	return AnnotateHubClientErrorf(c.processResponse(resp, result, expectedStatusCode), "unable to process response from GET to %s", url)
}

func (c *Client) HttpPutJSON(url string, data interface{}, contentType string, expectedStatusCode int) error {

	var resp *http.Response

	if c.debugFlags&HubClientDebugTimings != 0 {
		log.Debugf("DEBUG HTTP STARTING PUT REQUEST: %s", url)
	}

	// Encode json
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)

	if err := enc.Encode(&data); err != nil {
		return newHubClientError(nil, nil, fmt.Sprintf("error encoding json: %+v", err))
	}

	req, err := http.NewRequest(http.MethodPut, url, &buf)
	if err != nil {
		return newHubClientError(nil, nil, fmt.Sprintf("error creating http put request for %s: %+v", url, err))
	}

	req.Header.Set(HeaderNameContentType, contentType)

	c.doPreRequest(req)
	log.Debugf("PUT Request: %+v.", req)

	httpStart := time.Now()
	if resp, err = c.httpClient.Do(req); err != nil {
		body := readResponseBody(resp)
		return newHubClientError(body, resp, fmt.Sprintf("error getting HTTP Response from PUT to %s: %+v", url, err))
	}

	httpElapsed := time.Since(httpStart)

	if c.debugFlags&HubClientDebugTimings != 0 {
		log.Debugf("DEBUG HTTP PUT ELAPSED TIME: %d ms.   -- Request: %s", (httpElapsed / 1000 / 1000), url)
	}

	return AnnotateHubClientErrorf(c.processResponse(resp, nil, expectedStatusCode), "unable to process response from PUT to %s", url) // TODO: Maybe need a response too?
}

func (c *Client) HttpPostJSON(url string, data interface{}, contentType string, expectedStatusCode int) (string, error) {

	var resp *http.Response

	if c.debugFlags&HubClientDebugTimings != 0 {
		log.Debugf("DEBUG HTTP STARTING POST REQUEST: %s", url)
	}

	// Encode json
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)

	if err := enc.Encode(&data); err != nil {
		return "", newHubClientError(nil, nil, fmt.Sprintf("error encoding json: %+v", err))
	}

	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		return "", newHubClientError(nil, nil, fmt.Sprintf("error creating http post request for %s: %+v", url, err))
	}

	req.Header.Set(HeaderNameContentType, contentType)

	c.doPreRequest(req)
	log.Debugf("POST Request: %+v.", req)

	httpStart := time.Now()
	if resp, err = c.httpClient.Do(req); err != nil {
		body := readResponseBody(resp)
		return "", newHubClientError(body, resp, fmt.Sprintf("error getting HTTP Response from POST to %s: %+v", url, err))
	}

	httpElapsed := time.Since(httpStart)

	if c.debugFlags&HubClientDebugTimings != 0 {
		log.Debugf("DEBUG HTTP POST ELAPSED TIME: %d ms.   -- Request: %s", (httpElapsed / 1000 / 1000), url)
	}

	if err := c.processResponse(resp, nil, expectedStatusCode); err != nil {
		return "", AnnotateHubClientErrorf(err, "unable to process response from POST to %s", url)
	}

	return resp.Header.Get("Location"), nil
}

func (c *Client) HttpPostJSONExpectResult(url string, data interface{}, result interface{}, contentType string, expectedStatusCode int) (string, error) {

	var resp *http.Response

	if c.debugFlags&HubClientDebugTimings != 0 {
		log.Debugf("DEBUG HTTP STARTING POST REQUEST: %s", url)
	}

	// Encode json
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)

	if err := enc.Encode(&data); err != nil {
		return "", newHubClientError(nil, nil, fmt.Sprintf("error encoding json: %+v", err))
	}

	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		return "", newHubClientError(nil, nil, fmt.Sprintf("error creating http post request for %s: %+v", url, err))
	}

	req.Header.Set(HeaderNameContentType, contentType)

	c.doPreRequest(req)
	log.Debugf("POST Request: %+v.", req)

	httpStart := time.Now()
	if resp, err = c.httpClient.Do(req); err != nil {
		body := readResponseBody(resp)
		return "", newHubClientError(body, resp, fmt.Sprintf("error getting HTTP Response from POST to %s: %+v", url, err))
	}

	httpElapsed := time.Since(httpStart)

	if c.debugFlags&HubClientDebugTimings != 0 {
		log.Debugf("DEBUG HTTP POST ELAPSED TIME: %d ms.   -- Request: %s", (httpElapsed / 1000 / 1000), url)
	}

	if err := c.processResponse(resp, result, expectedStatusCode); err != nil {
		return "", AnnotateHubClientErrorf(err, "unable to process response from POST to %s", url)
	}

	return resp.Header.Get("Location"), nil
}

func (c *Client) HttpDelete(url string, contentType string, expectedStatusCode int) error {

	var resp *http.Response
	var err error

	if c.debugFlags&HubClientDebugTimings != 0 {
		log.Debugf("DEBUG HTTP STARTING DELETE REQUEST: %s", url)
	}

	req, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return newHubClientError(nil, nil, fmt.Sprintf("error creating http delete request for %s: %+v", url, err))
	}

	req.Header.Set(HeaderNameContentType, contentType)

	c.doPreRequest(req)
	log.Debugf("DELETE Request: %+v.", req)

	httpStart := time.Now()
	if resp, err = c.httpClient.Do(req); err != nil {
		body := readResponseBody(resp)
		return newHubClientError(body, resp, fmt.Sprintf("error getting HTTP Response from DELETE to %s: %+v", url, err))
	}

	httpElapsed := time.Since(httpStart)

	if c.debugFlags&HubClientDebugTimings != 0 {
		log.Debugf("DEBUG HTTP DELETE ELAPSED TIME: %d ms.   -- Request: %s", (httpElapsed / 1000 / 1000), url)
	}

	return AnnotateHubClientErrorf(c.processResponse(resp, nil, expectedStatusCode), "unable to process response from DELETE to %s", url)
}

func (c *Client) doPreRequest(request *http.Request) {

	if c.useAuthToken {
		request.Header.Set(HeaderNameAuthorization, fmt.Sprintf("Bearer %s", c.authToken))
	}

	if c.haveCsrfToken {
		request.Header.Set(HeaderNameCsrfToken, c.csrfToken)
	}
}

func readResponseBody(resp *http.Response) []byte {

	var bodyBytes []byte
	var err error

	if bodyBytes, err = readBytes(resp.Body); err != nil {
		log.Errorf("Error reading HTTP Response: %+v.", err)
	}

	log.Debugf("TEXT OF RESPONSE: \n %s", string(bodyBytes[:]))
	return bodyBytes
}

func newHubClientError(respBody []byte, resp *http.Response, message string) *HubClientError {
	var hre HubResponseError

	// Do not try to read the body of the response
	hce := &HubClientError{errors.NewErr(message), resp.StatusCode, hre}
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &hre); err != nil {
			hce = AnnotateHubClientError(hce, fmt.Sprintf("error unmarshaling HTTP response body: %+v", err)).(*HubClientError)
		}
		hce.HubError = hre
	}

	return hce
}
