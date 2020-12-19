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
	"github.com/jamesl33/goamt/database"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// createOptions - Encapsulates the options for the create sub-command.
var createOptions = struct {
	database string
}{}

// createCommand - The create sub-command, used to create a new empty goamt SQLite database.
var createCommand = &cobra.Command{
	RunE:  create,
	Short: "Create a new goamt SQLite database",
	Use:   "create",
}

// init - Initialize the flags/arguments for the create sub-command.
func init() {
	createCommand.Flags().StringVarP(
		&createOptions.database,
		"database",
		"d",
		"",
		"path where the database will be created",
	)

	markFlagRequired(createCommand, "database")
}

// create - Run the create sub-command, this will create a new empty goamt SQLite database file.
func create(_ *cobra.Command, _ []string) error {
	db, err := database.Create(createOptions.database)
	if err != nil {
		return errors.Wrap(err, "failed to create database")
	}

	err = db.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close database")
	}

	return nil
}
