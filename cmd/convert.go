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
	"fmt"
	"os"
	"runtime"
	"sort"
	"time"

	"github.com/jamesl33/goamt/database"
	"github.com/jamesl33/goamt/utils"
	"github.com/jamesl33/goamt/value"

	"github.com/apex/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// convertOptions - Encapsulates the options for the convert sub-command.
var convertOptions = struct {
	source, sink string
	threads      int
}{}

// convertCommand - The convert sub-command, used to convert a pytranscoder yaml file into a goamt SQLite database.
var convertCommand = &cobra.Command{
	RunE:  convert,
	Short: "Convert from the pytranscoder yaml format into the goamt SQLite format",
	Use:   "convert",
}

// init - Initialize the flags/arguments for the convert sub-command.
func init() {
	convertCommand.Flags().StringVarP(
		&convertOptions.source,
		"source",
		"s",
		"",
		"path to a yaml store created by pytranscoder",
	)

	convertCommand.Flags().StringVarP(
		&convertOptions.sink,
		"database",
		"d",
		"",
		"output path for the converted database",
	)

	convertCommand.Flags().IntVarP(
		&convertOptions.threads,
		"threads",
		"t",
		runtime.NumCPU(),
		"the number of threads to use, defaults to the number of vCPUs",
	)

	markFlagRequired(convertCommand, "source")
	markFlagRequired(convertCommand, "database")
}

// convert - Run the convert sub-command, this will create a new goamt SQLite database then concurrently hash and insert
// any media files found in the existing pytranscoder yaml file.
func convert(_ *cobra.Command, _ []string) error {
	ctx := signalHandler()

	if !utils.PathExists(convertOptions.source) {
		return fmt.Errorf("source file '%s' not found", convertOptions.source)
	}

	if utils.PathExists(convertOptions.sink) {
		return fmt.Errorf("sink file '%s' already exists", convertOptions.sink)
	}

	source, err := os.Open(convertOptions.source)
	if err != nil {
		return errors.Wrap(err, "failed to open source file")
	}
	defer source.Close()

	overlay := struct {
		Transcoded   []string `yaml:"transcoded,omitempty"`
		Untranscoded []string `yaml:"untranscoded,omitempty"`
	}{}

	err = yaml.NewDecoder(source).Decode(&overlay)
	if err != nil {
		return errors.Wrap(err, "failed to decode source file")
	}

	fields := log.Fields{"transcoded": len(overlay.Transcoded), "untranscoded": len(overlay.Untranscoded)}
	log.WithFields(fields).Debug("Successfully decoded source file")

	db, err := database.Create(convertOptions.sink)
	if err != nil {
		return errors.Wrap(err, "failed to create sink database")
	}

	var (
		pool                     = NewUpdatePool(db)
		entryStream, errorStream = pool.Start(ctx, convertOptions.threads)
	)

	// We should insert the untranscoded list first so that any more up-to-date entries in the transcoded list overwrite
	// those in the untranscoded list.
	err = queueEntries(ctx, entryStream, errorStream, sort.StringSlice(overlay.Untranscoded), false)
	if err != nil {
		return err // Purposefully not wrapped
	}

	err = queueEntries(ctx, entryStream, errorStream, sort.StringSlice(overlay.Transcoded), true)
	if err != nil {
		return err // Purposefully not wrapped
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

// queueEntries - Convert the provided slice of paths into entries and queue them for processing by the worker pool.
func queueEntries(ctx context.Context, entryStream chan<- value.Entry, errorStream <-chan error, paths []string,
	populateTranscoded bool) error {
	for _, path := range paths {
		var (
			discovered = time.Now().Unix()
			transcoded *int64
		)

		if populateTranscoded {
			transcoded = utils.Int64P(discovered)
		}

		queued, err := queueEntry(
			ctx,
			entryStream,
			errorStream,
			value.Entry{Path: path, Discovered: discovered, Transcoded: transcoded},
		)
		if err != nil {
			return errors.Wrap(err, "failed to queue entry")
		}

		if !queued {
			return nil
		}
	}

	return nil
}
