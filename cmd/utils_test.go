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

package cmd

import (
	"database/sql"
	"reflect"
	"sort"
	"testing"

	"github.com/jamesl33/goamt/database"
	"github.com/jamesl33/goamt/utils/sqlite"
	"github.com/jamesl33/goamt/value"

	"github.com/pkg/errors"
)

func createDatabaseAndPopulate(t *testing.T, path string, entries []value.Entry) {
	db, err := database.Create(path)
	if err != nil {
		t.Fatalf("Expected to be able to create database: %v", err)
	}
	defer db.Close()

	for _, entry := range entries {
		err = db.Upsert(entry)
		if err != nil {
			t.Fatalf("Expected to be able to upsert entry: %v", err)
		}
	}
}

func assertDatabaseContains(t *testing.T, path string, expected []value.Entry) {
	actual := make([]value.Entry, 0, len(expected))

	callback := func(scan sqlite.ScanCallback) error {
		var entry value.Entry
		err := scan(&entry.Path, &entry.Discovered, &entry.Transcoded, &entry.Hash)
		if err != nil {
			return err
		}

		actual = append(actual, entry)
		return nil
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("Expected to be able to open test database: %v", err)
	}
	defer db.Close()

	query := sqlite.Query{Query: "select path, discovered, transcoded, hash from library;"}

	err = sqlite.QueryRows(db, query, callback)
	if err != nil && !errors.Is(err, sqlite.ErrQueryReturnedNoRows) {
		t.Fatalf("Expected to be able to query entries: %v", err)
	}

	if len(actual) != len(expected) {
		t.Fatalf("Expected %d entries but got %d", len(expected), len(actual))
	}

	sort.Slice(actual, func(i, j int) bool { return actual[i].Path < actual[j].Path })
	sort.Slice(expected, func(i, j int) bool { return expected[i].Path < expected[j].Path })

	for index := range expected {
		if actual[index].Discovered == 0 {
			t.Fatalf("Expected a non-zero discovered value")
		}

		if (expected[index].Transcoded == nil) != (actual[index].Transcoded == nil) {
			t.Fatalf("Expected %t but got %t", expected[index].Transcoded == nil, actual[index].Transcoded == nil)
		}

		if actual[index].Hash == 0 {
			t.Fatalf("Expected a non-zero hash value")
		}

		actual[index].Discovered = 0
		expected[index].Discovered = 0

		actual[index].Transcoded = nil
		expected[index].Transcoded = nil

		actual[index].Hash = 0
		expected[index].Hash = 0

		if !reflect.DeepEqual(expected[index], actual[index]) {
			t.Fatalf("Database contained unexpected entries")
		}
	}
}
