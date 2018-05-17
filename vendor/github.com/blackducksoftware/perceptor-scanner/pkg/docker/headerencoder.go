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

package docker

import (
	b64 "encoding/base64"
	"fmt"
)

func base64Encode(data string) string {
	return b64.StdEncoding.EncodeToString([]byte(data))
}

func encodeAuthHeader(username string, password string) string {
	data := fmt.Sprintf("{ \"username\": \"%s\", \"password\": \"%s\" }", username, password)
	// fmt.Printf("debug:<\n%s\n>done debug %d\n\n", data, len(data))
	// bytesIn := []byte(data)
	// fmt.Printf("bytes in:<\n%#v\n>done bytes in %d\n\n", bytesIn, len(bytesIn))
	encoded := base64Encode(data)
	// fmt.Printf("encoded:<\n%s\n>done encoded %d\n\n", encoded, len(encoded))
	// bytesOut := []byte(encoded)
	// fmt.Printf("bytes out:<\n%#v\n>done bytes out %d\n\n", bytesOut, len(bytesOut))
	return encoded
}
