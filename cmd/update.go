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
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/jamesl33/goamt/database"
	"github.com/jamesl33/goamt/utils"
	"github.com/jamesl33/goamt/value"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// updateOptions - Encapsulates the options for the update sub-command.
var updateOptions = struct {
	database, path string
	threads        int
}{}

// updateCommand - The update sub-command, used to update the goamt SQLite database by walking the provided path and
// hashing and inserting media files as untranscoded entries.
var updateCommand = &cobra.Command{
	RunE:  update,
	Short: "Update a goamt SQLite database",
	Use:   "update",
}

// init - Initialize the flags/arguments for the update sub-command.
func init() {
	updateCommand.Flags().StringVarP(
		&updateOptions.database,
		"database",
		"d",
		"",
		"path to a goamt SQLite database",
	)

	updateCommand.Flags().StringVarP(
		&updateOptions.path,
		"path",
		"p",
		"",
		"path to a media library",
	)

	updateCommand.Flags().IntVarP(
		&updateOptions.threads,
		"threads",
		"t",
		runtime.NumCPU(),
		"the number of threads to use, defaults to the number of vCPUs",
	)

	markFlagRequired(updateCommand, "database")
	markFlagRequired(updateCommand, "path")
}

// update - Run the update sub-command, this will walk the provided path hashing and inserting media files as
// untranscoded entries in the provided goamt SQLite database.
func update(_ *cobra.Command, _ []string) error {
	ctx := signalHandler()

	db, err := database.Open(updateOptions.database)
	if err != nil {
		return errors.Wrap(err, "failed to open SQLite database")
	}

	var (
		pool                     = NewUpdatePool(db)
		entryStream, errorStream = pool.Start(ctx, updateOptions.threads)
	)

	err = filepath.Walk(updateOptions.path, func(path string, _ os.FileInfo, err error) error {
		if err != nil ||
			strings.HasSuffix(path, value.TranscodingExtension) ||
			!utils.ContainsString(value.SupportedExtensions, filepath.Ext(path)) {
			return err
		}

		if len(errorStream) != 0 {
			return <-errorStream
		}

		queued, err := queueEntry(
			ctx,
			entryStream,
			errorStream,
			value.Entry{Path: path, Discovered: time.Now().Unix()},
		)
		if err != nil {
			return errors.Wrap(err, "failed to queue entry")
		}

		if !queued {
			return io.EOF
		}

		return nil
	})
	if err != nil && err != io.EOF {
		return errors.Wrap(err, "unexpected error during file walk")
	}

	err = pool.Stop()
	if err != nil {
		return errors.Wrap(err, "failed to stop worker pool")
	}

	err = db.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close database")
	}

	return nil
}
