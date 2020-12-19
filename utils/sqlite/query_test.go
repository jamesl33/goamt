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
	"errors"
	"path/filepath"
	"testing"
)

func TestExecuteQuery(t *testing.T) {
	var (
		tempDir = t.TempDir()
		path    = filepath.Join(tempDir, "test.db")
	)

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("Expected to be able to open test database: %v", err)
	}
	defer db.Close()

	_, err = ExecuteQuery(db, Query{Query: "pragma user_version=42;"})
	if err != nil {
		t.Fatalf("Expected to be able to execute query: %v", err)
	}

	var version uint32
	err = GetPragma(db, PragmaUserVersion, &version)
	if err != nil {
		t.Fatalf("Expected to be able to get user_version: %v", err)
	}

	if version != 42 {
		t.Fatalf("Expected 42 but got %d", version)
	}
}

func TestQueryRow(t *testing.T) {
	var (
		tempDir = t.TempDir()
		path    = filepath.Join(tempDir, "test.db")
	)

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("Expected to be able to open test database: %v", err)
	}
	defer db.Close()

	_, err = ExecuteQuery(db, Query{Query: "create table test (id integer primary key);"})
	if err != nil {
		t.Fatalf("Expected to be able to execute query: %v", err)
	}

	_, err = ExecuteQuery(db, Query{Query: "insert into test (id) values (42)"})
	if err != nil {
		t.Fatalf("Expected to be able to execute query: %v", err)
	}

	var id int
	err = QueryRow(db, Query{Query: "select id from test limit 1;"}, &id)
	if err != nil {
		t.Fatalf("Expected to be able query rows: %v", err)
	}

	if id != 42 {
		t.Fatalf("Expected a single row with an id of 42")
	}
}

func TestQueryRowNoRows(t *testing.T) {
	var (
		tempDir = t.TempDir()
		path    = filepath.Join(tempDir, "test.db")
	)

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("Expected to be able to open test database: %v", err)
	}
	defer db.Close()

	_, err = ExecuteQuery(db, Query{Query: "create table test (id integer primary key);"})
	if err != nil {
		t.Fatalf("Expected to be able to execute query: %v", err)
	}

	var id int
	err = QueryRow(db, Query{Query: "select id from test limit 1;"}, &id)
	if !errors.Is(err, ErrQueryReturnedNoRows) {
		t.Fatalf("Expected an 'ErrQueryReturnedNoRows' error but got '%#v'", err)
	}
}

func TestQueryRows(t *testing.T) {
	var (
		tempDir = t.TempDir()
		path    = filepath.Join(tempDir, "test.db")
	)

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("Expected to be able to open test database: %v", err)
	}
	defer db.Close()

	_, err = ExecuteQuery(db, Query{Query: "create table test (id integer primary key);"})
	if err != nil {
		t.Fatalf("Expected to be able to execute query: %v", err)
	}

	_, err = ExecuteQuery(db, Query{Query: "insert into test (id) values (42)"})
	if err != nil {
		t.Fatalf("Expected to be able to execute query: %v", err)
	}

	ids := make([]int, 0)

	callback := func(scan ScanCallback) error {
		var id int
		err := scan(&id)
		ids = append(ids, id)
		return err
	}

	err = QueryRows(db, Query{Query: "select id from test;"}, callback)
	if err != nil {
		t.Fatalf("Expected to be able query rows: %v", err)
	}

	if len(ids) != 1 || ids[0] != 42 {
		t.Fatalf("Expected a single row with an id of 42")
	}
}

func TestQueryRowsNoRows(t *testing.T) {
	var (
		tempDir = t.TempDir()
		path    = filepath.Join(tempDir, "test.db")
	)

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("Expected to be able to open test database: %v", err)
	}
	defer db.Close()

	_, err = ExecuteQuery(db, Query{Query: "create table test (id integer primary key);"})
	if err != nil {
		t.Fatalf("Expected to be able to execute query: %v", err)
	}

	err = QueryRows(db, Query{Query: "select id from test;"}, func(scan ScanCallback) error { return nil })
	if !errors.Is(err, ErrQueryReturnedNoRows) {
		t.Fatalf("Expected an 'ErrQueryReturnedNoRows' error but got '%#v'", err)
	}
}
