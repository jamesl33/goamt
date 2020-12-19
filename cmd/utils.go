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
	"context"
	"os"

	"github.com/jamesl33/goamt/database"
	"github.com/jamesl33/goamt/utils"
	"github.com/jamesl33/goamt/value"

	"github.com/apex/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// markFlagRequired - Mark the provided flag as required panicking if it was not found.
func markFlagRequired(command *cobra.Command, flag string) {
	err := command.MarkFlagRequired(flag)
	if err != nil {
		panic(err)
	}
}

// queueEntry - Queue the provided entry, returns a boolean indicating whether the entry was successfully queued; the
// calling function should begin gracefully terminating in the event of a queue failure.
func queueEntry(ctx context.Context, entryStream chan<- value.Entry, errorStream <-chan error,
	entry value.Entry) (bool, error) {
	select {
	case <-ctx.Done():
		return false, nil
	case entryStream <- entry:
		return true, nil
	case err := <-errorStream:
		return false, err
	}
}

// upsertEntry - Update the hash for the provided entry then upsert it into the SQLite database.
func upsertEntry(db *database.Database, entry value.Entry) error {
	var err error
	entry.Hash, err = utils.HashFile(entry.Path)
	if err != nil {
		return err
	}

	return db.Upsert(entry)
}

// transcodeEntry - Transcode the provided entry, note that this entry should already exist in the provided database.
func transcodeEntry(db *database.Database, entry value.Entry) error {
	log.WithFields(entry).Info("Beginning job to transcode entry")

	err := transcodeFunc(entry.Path)
	if err != nil {
		return errors.Wrap(err, "failed to transcode file")
	}

	err = os.Remove(entry.Path)
	if err != nil {
		return errors.Wrap(err, "failed to remove source file")
	}

	err = os.Rename(utils.ReplaceExtension(entry.Path, value.TranscodingExtension),
		utils.ReplaceExtension(entry.Path, value.TargetExtension))
	if err != nil {
		return errors.Wrap(err, "failed to rename transcoded file")
	}

	entry.Path = utils.ReplaceExtension(entry.Path, value.TargetExtension)
	return db.CompleteTranscoding(entry)
}

// cancelTranscoding - Cancel the queued job to transcode an entry.
func cancelTranscoding(db *database.Database, entry value.Entry) error {
	err := db.CancelTranscoding(entry)
	if err != nil {
		return errors.Wrap(err, "failed to cancel job")
	}

	return nil
}
