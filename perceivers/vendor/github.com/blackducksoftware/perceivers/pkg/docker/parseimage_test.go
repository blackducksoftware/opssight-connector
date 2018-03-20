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
	"fmt"
	"testing"
)

func TestParseImageIDString(t *testing.T) {
	testcases := []struct {
		description string
		prefix      string
		name        string
		sha         string
		shouldPass  bool
	}{
		{
			description: "valid image name",
			prefix:      "docker-pullable://",
			name:        "abc/def",
			sha:         "cb4983d8399a59bb5ee6e68b6177d878966a8fe41abe18a45c3b1d8809f1d043",
			shouldPass:  true,
		},
		{
			description: "no prefix",
			prefix:      "",
			name:        "abc/def",
			sha:         "cb4983d8399a59bb5ee6e68b6177d878966a8fe41abe18a45c3b1d8809f1d043",
			shouldPass:  true,
		},
		{
			description: "missing image name",
			prefix:      "",
			name:        "",
			sha:         "cb4983d8399a59bb5ee6e68b6177d878966a8fe41abe18a45c3b1d8809f1d043",
			shouldPass:  false,
		},
		{
			description: "missing sha",
			prefix:      "",
			name:        "abc/def",
			sha:         "",
			shouldPass:  false,
		},
	}

	for _, tc := range testcases {
		imageID := fmt.Sprintf("%s%s@sha256:%s", tc.prefix, tc.name, tc.sha)
		name, sha, err := ParseImageIDString(imageID)
		if err != nil && tc.shouldPass {
			t.Errorf("[%s] unexpected error: %v, imageID %s", tc.description, err, imageID)
		}
		if name != tc.name && tc.shouldPass {
			t.Errorf("[%s] name is wrong.  Expected %s got %s", tc.description, tc.name, name)
		}
		if sha != tc.sha && tc.shouldPass {
			t.Errorf("[%s] sha is wrong.  Expected %s got %s", tc.description, tc.sha, sha)
		}
	}
}
