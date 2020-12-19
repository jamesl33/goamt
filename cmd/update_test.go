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
	"hash/crc32"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/jamesl33/goamt/database"
	"github.com/jamesl33/goamt/utils"
	"github.com/jamesl33/goamt/value"

	"github.com/pkg/errors"
)

func TestUpdateDatabaseNotFound(t *testing.T) {
	tempDir := t.TempDir()

	updateOptions.database = filepath.Join(tempDir, "goamt.db")
	updateOptions.path = tempDir

	err := update(nil, nil)

	var notFound *database.ErrNotFound
	if !errors.As(err, &notFound) {
		t.Fatalf("Expected an 'ErrNotFound' but got '%#v'", err)
	}
}

func TestUpdate(t *testing.T) {
	tempDir := t.TempDir()

	updateOptions.database = filepath.Join(tempDir, "goamt.db")
	updateOptions.path = tempDir

	expected := []value.Entry{
		{
			Path:       "transcoded1.mp4",
			Discovered: 8,
			Transcoded: utils.Int64P(0),
		},
		{
			Path:       "untranscoded1.mp4",
			Discovered: 16,
		},
		{
			Path: "untranscoded2.mp4",
		},
	}

	var count int

	for index := range expected {
		contents := []byte(strconv.Itoa(count))

		expected[index].Path = filepath.Join(tempDir, expected[index].Path)
		expected[index].Hash = crc32.Checksum(contents, crc32.MakeTable(crc32.IEEE))

		err := ioutil.WriteFile(expected[index].Path, contents, 0o755)
		if err != nil {
			t.Fatalf("Expected to be able to create test file: %v", err)
		}

		count++
	}

	createDatabaseAndPopulate(t, updateOptions.database, expected[:2])

	err := update(nil, nil)
	if err != nil {
		t.Fatalf("Expected to be able to update database: %v", err)
	}

	assertDatabaseContains(t, updateOptions.database, expected)
}
