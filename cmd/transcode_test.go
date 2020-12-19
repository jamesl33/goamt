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
	"reflect"
	"strconv"
	"testing"

	"github.com/jamesl33/goamt/database"
	"github.com/jamesl33/goamt/utils"
	"github.com/jamesl33/goamt/value"

	"github.com/pkg/errors"
)

func TestTranscodeDatabaseNotFound(t *testing.T) {
	tempDir := t.TempDir()

	transcodeOptions.database = filepath.Join(tempDir, "goamt.db")
	transcodeOptions.path = tempDir

	err := transcode(nil, nil)

	var notFound *database.ErrNotFound
	if !errors.As(err, &notFound) {
		t.Fatalf("Expected an 'ErrNotFound' but got '%#v'", err)
	}
}

func TestTranscode(t *testing.T) {
	tempDir := t.TempDir()

	transcodeOptions.database = filepath.Join(tempDir, "goamt.db")
	transcodeOptions.path = tempDir

	initial := []value.Entry{
		{
			Path:       "transcoded1.mp4",
			Discovered: 8,
			Transcoded: utils.Int64P(0),
		},
		{
			Path:       "untranscoded1.mp4",
			Discovered: 16,
		},
	}

	var count int

	for index := range initial {
		contents := []byte(strconv.Itoa(count))

		initial[index].Path = filepath.Join(tempDir, initial[index].Path)
		initial[index].Hash = crc32.Checksum(contents, crc32.MakeTable(crc32.IEEE))

		err := ioutil.WriteFile(initial[index].Path, contents, 0o755)
		if err != nil {
			t.Fatalf("Expected to be able to create test file: %v", err)
		}

		count++
	}

	createDatabaseAndPopulate(t, transcodeOptions.database, initial)

	transcoded := make([]string, 0)

	transcodeFunc = func(path string) error {
		transcoded = append(transcoded, path)

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return errors.Wrap(err, "failed to read file contents")
		}

		// Update the copied data so that we don't end up with a hash collision
		data = append(data, []byte("transcoded")...)
		return ioutil.WriteFile(utils.ReplaceExtension(path, value.TranscodingExtension), data, 0o755)
	}

	err := transcode(nil, nil)
	if err != nil {
		t.Fatalf("Expected to be able to transcode entries: %v", err)
	}

	if !reflect.DeepEqual(transcoded, []string{filepath.Join(tempDir, "untranscoded1.mp4")}) {
		t.Fatalf("Expected to have transcoded a single entry")
	}

	expected := []value.Entry{
		{
			Path:       filepath.Join(tempDir, "transcoded1.mp4"),
			Discovered: 8,
			Transcoded: utils.Int64P(0),
		},
		{
			Path:       filepath.Join(tempDir, "untranscoded1.mp4"),
			Discovered: 16,
			Transcoded: utils.Int64P(0),
		},
	}

	assertDatabaseContains(t, transcodeOptions.database, expected)
}

func TestTranscodeNoneToTranscode(t *testing.T) {
	tempDir := t.TempDir()

	transcodeOptions.database = filepath.Join(tempDir, "goamt.db")
	transcodeOptions.path = tempDir

	entries := []value.Entry{
		{
			Path:       "transcoded1.mp4",
			Discovered: 8,
			Transcoded: utils.Int64P(0),
		},
		{
			Path:       "transcoded2.mp4",
			Discovered: 16,
			Transcoded: utils.Int64P(0),
		},
	}

	var count int

	for index := range entries {
		contents := []byte(strconv.Itoa(count))

		entries[index].Path = filepath.Join(tempDir, entries[index].Path)
		entries[index].Hash = crc32.Checksum(contents, crc32.MakeTable(crc32.IEEE))

		err := ioutil.WriteFile(entries[index].Path, contents, 0o755)
		if err != nil {
			t.Fatalf("Expected to be able to create test file: %v", err)
		}

		count++
	}

	createDatabaseAndPopulate(t, transcodeOptions.database, entries)

	transcoded := make([]string, 0)

	transcodeFunc = func(path string) error {
		transcoded = append(transcoded, path)
		return nil
	}

	err := transcode(nil, nil)
	if err != nil {
		t.Fatalf("Expected to be able to transcode entries: %v", err)
	}

	if !reflect.DeepEqual(transcoded, make([]string, 0)) {
		t.Fatalf("Expected not to have transcoded any entries")
	}

	assertDatabaseContains(t, transcodeOptions.database, entries)
}
