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

package sqlite

import (
	"fmt"
)

// Pragma - Represents an SQLite pragma which can be used to modify the behavior of the unerlying SQLite library.
type Pragma string

const (
	// PragmaUserVersion - The pragma to get/set the SQLite user version; this value is ignored by the SQLite library.
	PragmaUserVersion Pragma = "user_version"

	// PragmaForiegnKeys - The pragma to enable/disable foreign keys between tables; this will ensure foreign references
	// exist when creating/updating/modifying rows.
	PragmaForiegnKeys Pragma = "foreign_keys"
)

// GetPragma - Query the provided pragma and store it in the given interface, note that it's the responsibility of the
// caller to ensure the provided interface is of the correct type.
func GetPragma(db Queryable, pragma Pragma, data interface{}) error {
	query := Query{
		Query: fmt.Sprintf("pragma %s;", pragma),
	}

	return QueryRow(db, query, data)
}

// SetPragma - Set the provided pragma to the given value, note that it's the responsibility of the caller to ensure the
// value is of the correct type.
func SetPragma(db Executable, pragma Pragma, value interface{}) error {
	query := Query{
		Query: fmt.Sprintf("pragma %s=%v;", pragma, value),
	}

	_, err := ExecuteQuery(db, query)
	return err
}
