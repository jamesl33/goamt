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
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/jamesl33/goamt/database"
	"github.com/jamesl33/goamt/utils"

	"github.com/pkg/errors"
)

func TestCreateAlreadyExists(t *testing.T) {
	tempDir := t.TempDir()
	createOptions.database = filepath.Join(tempDir, "goamt.db")

	err := ioutil.WriteFile(createOptions.database, make([]byte, 0), 0o755)
	if err != nil {
		t.Fatalf("Expected to be able to create test database file: %v", err)
	}

	err = create(nil, nil)

	var alreadyExists *database.ErrAlreadyExists
	if !errors.As(err, &alreadyExists) {
		t.Fatalf("Expected an 'ErrAlreadyExists' but got '%v'", err)
	}
}

func TestCreate(t *testing.T) {
	tempDir := t.TempDir()
	createOptions.database = filepath.Join(tempDir, "goamt.db")

	err := create(nil, nil)
	if err != nil {
		t.Fatalf("Expected to be able to create database: %v", err)
	}

	if !utils.PathExists(createOptions.database) {
		t.Fatalf("Expected database file to have been created")
	}
}
