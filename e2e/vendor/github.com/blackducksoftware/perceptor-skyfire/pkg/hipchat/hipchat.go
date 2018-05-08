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

package hipchat

import (
	"fmt"

	"github.com/go-resty/resty"
)

var authToken = "pRuvP8nq4nX0D8TpS1Kh5yyCrqt21d047PbBEWyv"

type Hipchat struct {
	Room string
}

func NewHipchat(room string) *Hipchat {
	return &Hipchat{
		Room: room,
	}
}

func (h *Hipchat) Send(message string) (*resty.Response, error) {
	url := BuildURL(h.Room)
	data := map[string]string{
		"message":        fmt.Sprintf("<pre>%s</pre>", message),
		"message_format": "html",
	}
	return resty.R().
		SetBody(data).
		SetHeader("Content-Type", "application/json").
		SetAuthToken(authToken).
		Post(url)
}

func BuildURL(room string) string {
	return fmt.Sprintf("https://api.hipchat.com/v2/room/%s/notification", room)
}
