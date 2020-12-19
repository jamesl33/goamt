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
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/jamesl33/goamt/utils"
	"github.com/jamesl33/goamt/value"

	"gopkg.in/yaml.v2"
)

func TestConvertSourceFileNotExists(t *testing.T) {
	tempDir := t.TempDir()
	convertOptions.source = filepath.Join(tempDir, "pytranscoder.yml")

	err := convert(nil, nil)
	if err == nil || err.Error() != fmt.Sprintf("source file '%s' not found", convertOptions.source) {
		t.Fatalf("Expected an error if source file does not exist")
	}
}

func TestConvertSinkFileAlreadyExists(t *testing.T) {
	tempDir := t.TempDir()
	convertOptions.source = filepath.Join(tempDir, "pytranscoder.yml")
	convertOptions.sink = filepath.Join(tempDir, "goamt.db")

	err := ioutil.WriteFile(convertOptions.source, make([]byte, 0), 0o755)
	if err != nil {
		t.Fatalf("Expected to be able to create test source file: %v", err)
	}

	err = ioutil.WriteFile(convertOptions.sink, make([]byte, 0), 0o755)
	if err != nil {
		t.Fatalf("Expected to be able to create test sink file: %v", err)
	}

	err = convert(nil, nil)
	if err == nil || err.Error() != fmt.Sprintf("sink file '%s' already exists", convertOptions.sink) {
		t.Fatalf("Expected an error if source file does not exist")
	}
}

func TestConvert(t *testing.T) {
	tempDir := t.TempDir()

	convertOptions.source = filepath.Join(tempDir, "pytranscoder.yml")
	convertOptions.sink = filepath.Join(tempDir, "goamt.db")

	contents := struct {
		Transcoded   []string `yaml:"transcoded"`
		Untranscoded []string `yaml:"untranscoded"`
	}{
		Transcoded:   []string{"transcoded1.mp4"},
		Untranscoded: []string{"untranscoded1.avi"},
	}

	var count int

	for index := range contents.Transcoded {
		contents.Transcoded[index] = filepath.Join(tempDir, contents.Transcoded[index])

		err := ioutil.WriteFile(contents.Transcoded[index], []byte(strconv.Itoa(count)), 0o755)
		if err != nil {
			t.Fatalf("Expected to be able to create test file: %v", err)
		}

		count++
	}

	for index := range contents.Untranscoded {
		contents.Untranscoded[index] = filepath.Join(tempDir, contents.Untranscoded[index])

		err := ioutil.WriteFile(contents.Untranscoded[index], []byte(strconv.Itoa(count)), 0o755)
		if err != nil {
			t.Fatalf("Expected to be able to create test file: %v", err)
		}

		count++
	}

	data, err := yaml.Marshal(contents)
	if err != nil {
		t.Fatalf("Expected to be able to marshal contents: %v", err)
	}

	err = ioutil.WriteFile(convertOptions.source, data, 0o755)
	if err != nil {
		t.Fatalf("Expected to be able to create test source file: %v", err)
	}

	err = convert(nil, nil)
	if err != nil {
		t.Fatalf("Expected to be able to convert file: %v", err)
	}

	expected := []value.Entry{
		{
			Path:       filepath.Join(tempDir, "transcoded1.mp4"),
			Transcoded: utils.Int64P(0),
		},
		{
			Path: filepath.Join(tempDir, "untranscoded1.avi"),
		},
	}

	assertDatabaseContains(t, convertOptions.sink, expected)
}
