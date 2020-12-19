// Copyright 2020 James Lee <jamesl33info@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesl33/goamt/value"
)

func TestPathExists(t *testing.T) {
	var (
		tempDir = t.TempDir()
		path    = filepath.Join(tempDir, "test.file")
	)

	if PathExists(path) {
		t.Fatalf("Expected false but got true")
	}

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Expected to be able to create test file: %v", err)
	}

	err = file.Close()
	if err != nil {
		t.Fatalf("Expected to be able to close test file: %v", err)
	}

	if !PathExists(path) {
		t.Fatalf("Expected true but got false")
	}
}

func TestPathReplaceExtension(t *testing.T) {
	type test struct {
		path      string
		extension string
		expected  string
	}

	tests := []*test{
		{
			path:      "test.mp4",
			extension: value.TranscodingExtension,
			expected:  "test.transcoding.mp4",
		},
		{
			path:      "this/is/a/relative/path/test.mp4",
			extension: value.TranscodingExtension,
			expected:  "this/is/a/relative/path/test.transcoding.mp4",
		},
		{
			path:      "/this/is/an/absolute/path/test.mp4",
			extension: value.TranscodingExtension,
			expected:  "/this/is/an/absolute/path/test.transcoding.mp4",
		},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			actual := ReplaceExtension(test.path, test.extension)
			if actual != test.expected {
				t.Fatalf("Expected '%s' but got '%s'", test.expected, actual)
			}
		})
	}
}
