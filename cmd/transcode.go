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
	"runtime"

	"github.com/jamesl33/goamt/database"
	"github.com/jamesl33/goamt/utils"
	"github.com/jamesl33/goamt/utils/sqlite"
	"github.com/jamesl33/goamt/value"

	"github.com/apex/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// transcodeOptions - Encapsulates the options for the transcode sub-command.
var transcodeOptions = struct {
	database, path   string
	entries, threads int
}{}

// transcodeCommand - The transcode sub-command, used to transcode a number of entries in the goamt database.
var transcodeCommand = &cobra.Command{
	RunE:  transcode,
	Short: "Concurrently transcode a number of files",
	Use:   "transcode",
}

// init - Initialize the flags/arguments for the transcode sub-command.
func init() {
	transcodeCommand.Flags().StringVarP(
		&transcodeOptions.database,
		"database",
		"d",
		"",
		"path to a goamt SQLite database",
	)

	transcodeCommand.Flags().StringVarP(
		&transcodeOptions.path,
		"path",
		"p",
		"",
		"path to a media library",
	)

	transcodeCommand.Flags().IntVarP(
		&transcodeOptions.entries,
		"entries",
		"e",
		runtime.NumCPU(),
		"the number of entries to transcode, defaults to the number of vCPUs",
	)

	transcodeCommand.Flags().IntVarP(
		&transcodeOptions.threads,
		"threads",
		"t",
		runtime.NumCPU(),
		"the number of threads to use, defaults to the number of vCPUs",
	)

	markFlagRequired(transcodeCommand, "database")
	markFlagRequired(transcodeCommand, "path")
}

// transcode - Run the transcode sub-command, this will transcode a number of entries in the SQLite database then update
// the transcoded timestamp (to avoid re-transcoding).
func transcode(_ *cobra.Command, _ []string) error {
	ctx := signalHandler()

	db, err := database.Open(transcodeOptions.database)
	if err != nil {
		return errors.Wrap(err, "failed to open SQLite database")
	}

	entries := make([]value.Entry, 0, transcodeOptions.entries)

	for len(entries) != transcodeOptions.entries {
		entry, err := db.BeginTranscoding()
		if err != nil {
			if errors.Is(err, sqlite.ErrQueryReturnedNoRows) {
				break
			}

			return errors.Wrap(err, "failed to get transcode entry")
		}

		if !utils.PathExists(entry.Path) {
			log.WithFields(entry).Warn("Found an entry that no longer exists, will remove")

			err = db.Remove(entry)
			if err != nil {
				return errors.Wrap(err, "failed to remove entry")
			}

			continue
		}

		entries = append(entries, entry)
	}

	var (
		pool                     = NewTranscodePool(db)
		entryStream, errorStream = pool.Start(ctx, transcodeOptions.threads)
	)

	for _, entry := range entries {
		queued, err := queueEntry(ctx, entryStream, errorStream, entry)
		if err != nil {
			return errors.Wrap(err, "failed to queue entry")
		}

		if !queued {
			break
		}
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
