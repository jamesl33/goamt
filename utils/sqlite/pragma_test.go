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
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestGetSetPragma(t *testing.T) {
	var (
		tempDir = t.TempDir()
		path    = filepath.Join(tempDir, "test.db")
	)

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("Expected to be able to open test database: %v", err)
	}
	defer db.Close()

	var version uint32
	err = GetPragma(db, PragmaUserVersion, &version)
	if err != nil {
		t.Fatalf("Expected to be able to get 'user_version': %v", err)
	}

	if version != 0 {
		t.Fatalf("Expected 0 but got %d", version)
	}

	err = SetPragma(db, PragmaUserVersion, 42)
	if err != nil {
		t.Fatalf("Expected to be able to set 'user_version': %v", err)
	}

	err = GetPragma(db, PragmaUserVersion, &version)
	if err != nil {
		t.Fatalf("Expected to be able to get 'user_version': %v", err)
	}

	if version != 42 {
		t.Fatalf("Expected 42 but got %d", version)
	}
}
