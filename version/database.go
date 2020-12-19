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

// DatabaseVersion - Represents a goamt SQLite database version (stored in the 'user_version').
type DatabaseVersion uint32

const (
	// DatabaseVersionOne - Initial release version.
	DatabaseVersionOne DatabaseVersion = iota + 1

	// DatabaseVersionCurrent - Convenience alias to avoid having to update the version in multiple places when bumping
	// the version number.
	DatabaseVersionCurrent = DatabaseVersionOne
)

// Supported - Returns a boolean indicating whether this database version is supported by goamt.
func (d DatabaseVersion) Supported() bool {
	return d != 0 && d <= 1
}
