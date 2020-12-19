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

package database

import (
	"database/sql"
	"hash/crc32"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"testing"

	"github.com/jamesl33/goamt/utils"
	"github.com/jamesl33/goamt/utils/sqlite"
	"github.com/jamesl33/goamt/value"

	"github.com/pkg/errors"
)

func createAndPopulate(t *testing.T, path string, entries []value.Entry, jobs []int) {
	db, err := Create(path)
	if err != nil {
		t.Fatalf("Expected to be able to create test database: %v", err)
	}
	defer db.Close()

	for _, entry := range entries {
		err = db.Upsert(entry)
		if err != nil {
			t.Fatalf("Expected to be able to upsert entry: %v", err)
		}
	}

	for _, job := range jobs {
		err := db.wrapTransaction(func(tx *sql.Tx) error {
			return db.addJob(tx, value.Entry{ID: job})
		})
		if err != nil {
			t.Fatalf("Expected to be able to add job: %v", err)
		}
	}
}

func openAndUpdate(t *testing.T, path string, entries []value.Entry) {
	db, err := Open(path)
	if err != nil {
		t.Fatalf("Expected to be able to open test database: %v", err)
	}
	defer db.Close()

	for _, entry := range entries {
		err = db.Upsert(entry)
		if err != nil {
			t.Fatalf("Expected to be able to upsert entry: %v", err)
		}
	}
}

func openAndRemove(t *testing.T, path string, entries []value.Entry) {
	db, err := Open(path)
	if err != nil {
		t.Fatalf("Expected to be able to open test database: %v", err)
	}
	defer db.Close()

	for _, entry := range entries {
		err = db.Remove(entry)
		if err != nil {
			t.Fatalf("Expected to be able to remove entry: %v", err)
		}
	}
}

func assertContains(t *testing.T, path string, expectedEntries []value.Entry, expectedJobs []int) {
	actualEntries := make([]value.Entry, 0, len(expectedEntries))

	callback := func(scan sqlite.ScanCallback) error {
		var entry value.Entry
		err := scan(&entry.Path, &entry.Discovered, &entry.Transcoded, &entry.Hash)
		if err != nil {
			return err
		}

		actualEntries = append(actualEntries, entry)
		return nil
	}

	db, err := Open(path)
	if err != nil {
		t.Fatalf("Expected to be able to open test database: %v", err)
	}
	defer db.Close()

	query := sqlite.Query{Query: "select path, discovered, transcoded, hash from library;"}

	err = sqlite.QueryRows(db.db, query, callback)
	if err != nil && !errors.Is(err, sqlite.ErrQueryReturnedNoRows) {
		t.Fatalf("Expected to be able to query entries: %v", err)
	}

	if len(actualEntries) != len(expectedEntries) {
		t.Fatalf("Expected %d entries but got %d", len(expectedEntries), len(actualEntries))
	}

	sort.Slice(actualEntries, func(i, j int) bool { return actualEntries[i].Path < actualEntries[j].Path })
	sort.Slice(expectedEntries, func(i, j int) bool { return expectedEntries[i].Path < expectedEntries[j].Path })

	for index := range expectedEntries {
		if (expectedEntries[index].Transcoded == nil) != (actualEntries[index].Transcoded == nil) {
			t.Fatalf("Expected %t but got %t", expectedEntries[index].Transcoded == nil,
				actualEntries[index].Transcoded == nil)
		}

		// We don't care what the timestamp is, as long as it's not <nil>
		expectedEntries[index].Transcoded = nil
		actualEntries[index].Transcoded = nil

		if !reflect.DeepEqual(expectedEntries[index], actualEntries[index]) {
			t.Fatalf("Database contained unexpected entries")
		}
	}

	actualJobs := make([]int, 0, len(expectedJobs))

	callback = func(scan sqlite.ScanCallback) error {
		var id int
		err := scan(&id)
		if err != nil {
			return err
		}

		actualJobs = append(actualJobs, id)
		return nil
	}

	query = sqlite.Query{Query: "select * from jobs;"}

	err = sqlite.QueryRows(db.db, query, callback)
	if err != nil && !errors.Is(err, sqlite.ErrQueryReturnedNoRows) {
		t.Fatalf("Expected to be able to query jobs: %v", err)
	}

	if !reflect.DeepEqual(actualJobs, expectedJobs) {
		t.Fatalf("Database contained unexpected jobs")
	}
}

func TestCreateAlreadyExists(t *testing.T) {
	var (
		tempDir = t.TempDir()
		path    = filepath.Join(tempDir, "test.db")
	)

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Expected to be able to create test file: %v", err)
	}

	err = file.Close()
	if err != nil {
		t.Fatalf("Expected to be able to close test file: %v", err)
	}

	_, err = Create(path)

	var alreadyExists *ErrAlreadyExists
	if !errors.As(err, &alreadyExists) {
		t.Fatalf("Expected an 'ErrAlreadyExists' but got '%#v'", err)
	}
}

func TestOpenNotFound(t *testing.T) {
	var (
		tempDir = t.TempDir()
		path    = filepath.Join(tempDir, "test.db")
	)

	_, err := Open(path)

	var notFound *ErrNotFound
	if !errors.As(err, &notFound) {
		t.Fatalf("Expected an 'ErrNotFound' but got '%#v'", err)
	}
}

func TestOpenRecoverIncompleteJobs(t *testing.T) {
	hash := func(data []byte) uint32 {
		return crc32.Checksum(data, crc32.MakeTable(crc32.IEEE))
	}

	type test struct {
		name            string
		initialEntries  []value.Entry
		initialFiles    []string
		initialJobs     []int
		expectedEntries []value.Entry
		expectedFiles   []string
		expectedJobs    []int
	}

	tests := []*test{
		{
			name:            "NoJobs",
			initialEntries:  []value.Entry{{Path: "test.mp4", Discovered: 42, Hash: hash([]byte("0"))}},
			initialFiles:    []string{"test.mp4"},
			expectedEntries: []value.Entry{{Path: "test.mp4", Discovered: 42, Hash: hash([]byte("0"))}},
			expectedFiles:   []string{"test.mp4"},
			expectedJobs:    make([]int, 0),
		},
		{
			name:            "OneJobBothFilesExist",
			initialEntries:  []value.Entry{{Path: "test.mp4", Discovered: 42, Hash: hash([]byte("0"))}},
			initialFiles:    []string{"test.mp4", "test.transcoding.mp4"},
			initialJobs:     []int{1},
			expectedEntries: []value.Entry{{Path: "test.mp4", Discovered: 42, Hash: hash([]byte("0"))}},
			expectedFiles:   []string{"test.mp4"},
			expectedJobs:    make([]int, 0),
		},
		{
			name:            "OneJobOnlySourceFileExists",
			initialEntries:  []value.Entry{{Path: "test.mp4", Discovered: 42, Hash: hash([]byte("0"))}},
			initialFiles:    []string{"test.mp4"},
			initialJobs:     []int{1},
			expectedEntries: []value.Entry{{Path: "test.mp4", Discovered: 42, Hash: hash([]byte("0"))}},
			expectedFiles:   []string{"test.mp4"},
			expectedJobs:    make([]int, 0),
		},
		{
			name:           "OneJobOnlyTargetFileExists",
			initialEntries: []value.Entry{{Path: "test.avi", Discovered: 42, Hash: hash([]byte("old_contents"))}},
			initialFiles:   []string{"test.transcoding.mp4"},
			initialJobs:    []int{1},
			expectedEntries: []value.Entry{
				{Path: "test.mp4", Discovered: 42, Transcoded: utils.Int64P(0), Hash: hash([]byte("0"))},
			},
			expectedFiles: []string{"test.mp4"},
			expectedJobs:  make([]int, 0),
		},
		{
			name:           "OneJobOnlyTargetFileExistsSameName",
			initialEntries: []value.Entry{{Path: "test.mp4", Discovered: 42, Hash: hash([]byte("old_contents"))}},
			initialFiles:   []string{"test.mp4"},
			initialJobs:    []int{1},
			expectedEntries: []value.Entry{
				{Path: "test.mp4", Discovered: 42, Transcoded: utils.Int64P(0), Hash: hash([]byte("0"))},
			},
			expectedFiles: []string{"test.mp4"},
			expectedJobs:  make([]int, 0),
		},
		{
			name:           "OneJobOnlyTargetFileExistsNotYetRenamed",
			initialEntries: []value.Entry{{Path: "test.mp4", Discovered: 42, Hash: hash([]byte("old_contents"))}},
			initialFiles:   []string{"test.transcoding.mp4"},
			initialJobs:    []int{1},
			expectedEntries: []value.Entry{
				{Path: "test.mp4", Discovered: 42, Transcoded: utils.Int64P(0), Hash: hash([]byte("0"))},
			},
			expectedFiles: []string{"test.mp4"},
			expectedJobs:  make([]int, 0),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var (
				tempDir = t.TempDir()
				path    = filepath.Join(tempDir, "test.db")
			)

			for index := range test.initialEntries {
				test.initialEntries[index].Path = filepath.Join(tempDir, test.initialEntries[index].Path)
			}

			for index := range test.expectedEntries {
				test.expectedEntries[index].Path = filepath.Join(tempDir, test.expectedEntries[index].Path)
			}

			defer os.Remove(path)

			createAndPopulate(t, path, test.initialEntries, test.initialJobs)

			for index, path := range test.initialFiles {
				err := ioutil.WriteFile(filepath.Join(tempDir, path), []byte(strconv.Itoa(index)), 0o755)
				if err != nil {
					t.Fatalf("Expected to be able to create test file: %v", err)
				}
			}

			openAndUpdate(t, path, nil)

			assertContains(t, path, test.expectedEntries, test.expectedJobs)

			for _, path := range test.expectedFiles {
				if !utils.PathExists(filepath.Join(tempDir, path)) {
					t.Fatalf("Expected file '%s' to exist", path)
				}
			}

			for _, path := range test.initialFiles {
				if !utils.ContainsString(test.expectedFiles, path) && utils.PathExists(filepath.Join(tempDir, path)) {
					t.Fatalf("Expected file '%s' to not exist", path)
				}
			}
		})
	}
}

func TestDatabaseUpsert(t *testing.T) {
	var (
		tempDir = t.TempDir()
		path    = filepath.Join(tempDir, "test.db")
	)

	createAndPopulate(t, path, nil, nil)

	update := []value.Entry{
		{
			Path:       "test.mp4",
			Discovered: 8,
			Transcoded: utils.Int64P(0),
			Hash:       32,
		},
	}

	openAndUpdate(t, path, update)
	assertContains(t, path, update, make([]int, 0))
}

func TestDatabaseUpsertIgnoreEntry(t *testing.T) {
	var (
		tempDir = t.TempDir()
		path    = filepath.Join(tempDir, "test.db")
	)

	initial := []value.Entry{
		{
			Path:       "test.mp4",
			Discovered: 8,
			Transcoded: utils.Int64P(0),
			Hash:       32,
		},
	}

	createAndPopulate(t, path, initial, nil)

	update := []value.Entry{
		{
			Path:       "test.mp4",
			Discovered: 8,
			Hash:       32,
		},
	}

	openAndUpdate(t, path, update)
	assertContains(t, path, initial, make([]int, 0))
}

func TestDatabaseUpsertRenameEntry(t *testing.T) {
	var (
		tempDir = t.TempDir()
		path    = filepath.Join(tempDir, "test.db")
	)

	initial := []value.Entry{
		{
			Path:       "test.mp4",
			Discovered: 8,
			Transcoded: utils.Int64P(0),
			Hash:       32,
		},
	}

	createAndPopulate(t, path, initial, nil)

	update := []value.Entry{
		{
			Path:       "renamed.mp4",
			Discovered: 8,
			Hash:       32,
		},
	}

	openAndUpdate(t, path, update)

	expected := []value.Entry{
		{
			Path:       "renamed.mp4",
			Discovered: 8,
			Transcoded: utils.Int64P(0),
			Hash:       32,
		},
	}

	assertContains(t, path, expected, make([]int, 0))
}

func TestDatabaseUpsertHashUpdated(t *testing.T) {
	var (
		tempDir = t.TempDir()
		path    = filepath.Join(tempDir, "test.db")
	)

	initial := []value.Entry{
		{
			Path:       "test.mp4",
			Discovered: 8,
			Transcoded: utils.Int64P(0),
			Hash:       32,
		},
	}

	createAndPopulate(t, path, initial, nil)

	update := []value.Entry{
		{
			Path:       "test.mp4",
			Discovered: 8,
			Hash:       64,
		},
	}

	openAndUpdate(t, path, update)

	expected := []value.Entry{
		{
			Path:       "test.mp4",
			Discovered: 8,
			Hash:       64,
		},
	}

	assertContains(t, path, expected, make([]int, 0))
}

func TestDatabaseRemove(t *testing.T) {
	var (
		tempDir = t.TempDir()
		path    = filepath.Join(tempDir, "test.db")
	)

	initial := []value.Entry{
		{
			Path:       "test.mp4",
			Discovered: 8,
			Transcoded: utils.Int64P(0),
			Hash:       32,
		},
	}

	createAndPopulate(t, path, initial, nil)

	remove := []value.Entry{
		{
			ID: 1,
		},
	}

	openAndRemove(t, path, remove)
	assertContains(t, path, make([]value.Entry, 0), make([]int, 0))
}

func TestDatabaseBeginTranscoding(t *testing.T) {
	var (
		tempDir = t.TempDir()
		path    = filepath.Join(tempDir, "test.db")
	)

	initial := []value.Entry{
		{
			Path:       "test.mp4",
			Discovered: 8,
			Hash:       16,
		},
	}

	createAndPopulate(t, path, initial, nil)

	db, err := Open(path)
	if err != nil {
		t.Fatalf("Expected to be able to open test database: %v", err)
	}
	defer db.Close()

	expected := value.Entry{
		ID:   1,
		Path: "test.mp4",
		Hash: 16,
	}

	entry, err := db.BeginTranscoding()
	if err != nil {
		t.Fatalf("Expected to be able to begin transcoding entry: %v", err)
	}

	if !reflect.DeepEqual(entry, expected) {
		t.Fatalf("Received an unexpected entry")
	}
}

func TestDatabaseBeginTranscodingOldestEntry(t *testing.T) {
	var (
		tempDir = t.TempDir()
		path    = filepath.Join(tempDir, "test.db")
	)

	initial := []value.Entry{
		{
			Path:       "test.avi",
			Discovered: 32,
			Hash:       64,
		},
		{
			Path:       "test.mp4",
			Discovered: 8,
			Hash:       16,
		},
	}

	createAndPopulate(t, path, initial, nil)

	db, err := Open(path)
	if err != nil {
		t.Fatalf("Expected to be able to open test database: %v", err)
	}
	defer db.Close()

	expected := value.Entry{
		ID:   2,
		Path: "test.mp4",
		Hash: 16,
	}

	entry, err := db.BeginTranscoding()
	if err != nil {
		t.Fatalf("Expected to be able to begin transcoding entry: %v", err)
	}

	if !reflect.DeepEqual(entry, expected) {
		t.Fatalf("Received an unexpected entry")
	}
}

func TestDatabaseBeginTranscodingNoEntries(t *testing.T) {
	var (
		tempDir = t.TempDir()
		path    = filepath.Join(tempDir, "test.db")
	)

	db, err := Create(path)
	if err != nil {
		t.Fatalf("Expected to be able to create test database: %v", err)
	}
	defer db.Close()

	_, err = db.BeginTranscoding()
	if err == nil || !errors.Is(err, sqlite.ErrQueryReturnedNoRows) {
		t.Fatalf("Expected to get an 'ErrQueryReturnedNoRows' but got '%#v'", err)
	}
}

func TestDatabaseCompleteTranscoding(t *testing.T) {
	var (
		tempDir = t.TempDir()
		path    = filepath.Join(tempDir, "test.db")
	)

	initial := []value.Entry{
		{
			Path:       filepath.Join(tempDir, "test.mp4"),
			Discovered: 8,
			Hash:       16,
		},
	}

	err := ioutil.WriteFile(filepath.Join(tempDir, "test.mp4"), []byte("Hello, World!"), 0o755)
	if err != nil {
		t.Fatalf("Expected to be able to create test file: %v", err)
	}

	createAndPopulate(t, path, initial, nil)

	db, err := Open(path)
	if err != nil {
		t.Fatalf("Expected to be able to open test database: %v", err)
	}

	err = db.CompleteTranscoding(value.Entry{
		ID:         1,
		Path:       filepath.Join(tempDir, "test.mp4"),
		Transcoded: utils.Int64P(0),
	})
	if err != nil {
		t.Fatalf("Expected to be able to mark transcoding complete: %v", err)
	}

	err = db.Close()
	if err != nil {
		t.Fatalf("Expected to be able to close test database: %v", err)
	}

	expected := []value.Entry{
		{
			Path:       filepath.Join(tempDir, "test.mp4"),
			Discovered: 8,
			Transcoded: utils.Int64P(0),
			Hash:       crc32.Checksum([]byte("Hello, World!"), crc32.MakeTable(crc32.IEEE)),
		},
	}

	assertContains(t, path, expected, make([]int, 0))
}
