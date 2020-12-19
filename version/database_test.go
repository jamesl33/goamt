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

package version

import "testing"

func TestDatabaseVersionSupported(t *testing.T) {
	if !DatabaseVersionOne.Supported() {
		t.Fatalf("Expected true but got false")
	}

	if !DatabaseVersionCurrent.Supported() {
		t.Fatalf("Expected true but got false")
	}

	if DatabaseVersion(42).Supported() {
		t.Fatalf("Expected false but got true")
	}
}
