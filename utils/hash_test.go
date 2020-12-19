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
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func TestHashFile(t *testing.T) {
	type test struct {
		name     string
		contents string
		expected uint32
	}

	tests := []*test{
		{
			name:     "LessThan4K",
			contents: "Hello, World!",
			expected: 3964322768,
		},
		{
			name:     "EqualTo4K",
			contents: strings.Repeat("x", 4096),
			expected: 1041266625,
		},
		{
			// It's expected for these two different files to have the same hash, this is due to the fact that we're
			// seeking further than the end of the file. In reality this shouldn't be too much of an issue since media
			// files are generally very large.
			name:     "GreaterThan4K",
			contents: strings.Repeat("x", 8192),
			expected: 1041266625,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var (
				tempDir = t.TempDir()
				path    = filepath.Join(tempDir, "test.file")
			)

			err := ioutil.WriteFile(path, []byte(test.contents), 0o755)
			if err != nil {
				t.Fatalf("Expected to be able to create test file: %v", err)
			}

			actual, err := HashFile(path)
			if err != nil {
				t.Fatalf("Expected to be able to hash test file: %v", err)
			}

			if actual != test.expected {
				t.Fatalf("Expected %d but got %d", test.expected, actual)
			}
		})
	}
}
