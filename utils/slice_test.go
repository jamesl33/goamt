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

import "testing"

func TestContainsString(t *testing.T) {
	type test struct {
		name     string
		slice    []string
		item     string
		expected bool
	}

	tests := []*test{
		{
			name: "NilSlice",
			item: "string",
		},
		{
			name:  "EmptySlice",
			slice: make([]string, 0),
			item:  "string",
		},
		{
			name:  "NonEmptyNotFound",
			slice: []string{"not", "here"},
			item:  "string",
		},
		{
			name:     "NonEmptyFound",
			slice:    []string{"the", "string", "is", "here"},
			item:     "string",
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := ContainsString(test.slice, test.item)
			if actual != test.expected {
				t.Fatalf("Expected %t but got %t", test.expected, actual)
			}
		})
	}
}
