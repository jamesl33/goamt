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
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jamesl33/goamt/utils"
	"github.com/jamesl33/goamt/utils/sqlite"
	"github.com/jamesl33/goamt/value"
	"github.com/jamesl33/goamt/version"

	"github.com/apex/log"
	_ "github.com/mattn/go-sqlite3" // SQLite database driver, unreferenced but required
	"github.com/pkg/errors"
)

// Database - Represents a connection to a goamt SQLite database and exposes a thread safe interface.
type Database struct {
	db   *sql.DB
	txns int
	lock sync.Mutex
}

// Create - Create a new database, returning an error if an existing database already exists.
func Create(path string) (*Database, error) {
	if utils.PathExists(path) {
		return nil, &ErrAlreadyExists{what: "database", where: path}
	}

	db, err := sql.Open("sqlite3", path+"?_journal=wal&_mutex=full&_sync=extra&mode=rwc")
	if err != nil {
		return nil, errors.Wrap(err, "failed to open SQLite database")
	}

	err = sqlite.SetPragma(db, sqlite.PragmaUserVersion, version.DatabaseVersionCurrent)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set 'user_version'")
	}

	err = sqlite.SetPragma(db, sqlite.PragmaForiegnKeys, "on")
	if err != nil {
		return nil, errors.Wrap(err, "failed to set 'foreign_keys'")
	}

	query := sqlite.Query{
		Query: `
			create table library (
				id integer primary key autoincrement,
				path text not null unique,
				discovered integer not null,
				transcoded integer,
				hash integer unique,
				unique (path, hash)
			);`,
	}

	_, err = sqlite.ExecuteQuery(db, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create library table")
	}

	query.Query = `
		create table jobs (
			id integer primary key autoincrement,
			library_id integer not null unique,
			start_time integer not null,
			foreign key (library_id) references library (id)
		);
	`

	_, err = sqlite.ExecuteQuery(db, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create jobs table")
	}

	log.WithField("version", version.DatabaseVersionCurrent).Info("Created new database")

	return &Database{db: db}, nil
}

// Open - Open an existing database returning an error if the provided database is missing or an unsupported version.
func Open(path string) (*Database, error) {
	if !utils.PathExists(path) {
		return nil, &ErrNotFound{what: "database", where: path}
	}

	db, err := sql.Open("sqlite3", path+"?_journal=wal&_mutex=full&_sync=extra&mode=rw")
	if err != nil {
		return nil, errors.Wrap(err, "failed to open SQLite database")
	}

	var userVersion uint32
	err = sqlite.GetPragma(db, sqlite.PragmaUserVersion, &userVersion)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get 'user_version'")
	}

	log.WithField("version", userVersion).Info("Opened existing database")

	if !version.DatabaseVersion(userVersion).Supported() {
		return nil, &ErrUnknownVersion{what: "database", where: path}
	}

	err = sqlite.SetPragma(db, sqlite.PragmaForiegnKeys, "on")
	if err != nil {
		return nil, errors.Wrap(err, "failed to set 'foreign_keys'")
	}

	database := &Database{db: db}

	err = database.recoverIncompleteJobs()
	if err != nil {
		return nil, errors.Wrap(err, "failed to recover incomplete jobs")
	}

	return database, nil
}

// recoverIncompleteJobs - Scan then handle any in-progress transcode jobs; this will revert or complete jobs depending
// on their status.
func (d *Database) recoverIncompleteJobs() error {
	callback := func(scan sqlite.ScanCallback) error {
		var entry value.Entry
		err := scan(&entry.ID, &entry.Path, &entry.Discovered, &entry.Transcoded, &entry.Hash)
		if err != nil {
			return errors.Wrap(err, "failed to scan incomplete job information")
		}

		log.WithFields(entry).Warn("Found incomplete job")

		hash, err := utils.HashFile(entry.Path)
		if (err == nil && hash != entry.Hash) || (!utils.PathExists(entry.Path) &&
			utils.PathExists(utils.ReplaceExtension(entry.Path, value.TranscodingExtension))) {
			return d.completeIncompleteJob(entry)
		}

		return d.rollbackIncompleteJob(entry)
	}

	query := sqlite.Query{
		Query: `select library.id, path,discovered,transcoded,hash from jobs
				inner join library on jobs.library_id = library.id`,
	}

	err := sqlite.QueryRows(d.db, query, callback)
	if err != nil && !errors.Is(err, sqlite.ErrQueryReturnedNoRows) {
		return errors.Wrap(err, "failed to query incomplete jobs")
	}

	return nil
}

// completeIncompleteJob - Complete the incomplete transcode job for the provided entry.
func (d *Database) completeIncompleteJob(entry value.Entry) error {
	log.WithFields(entry).Info("Completing incomplete job")

	err := os.Rename(
		utils.ReplaceExtension(entry.Path, value.TranscodingExtension),
		utils.ReplaceExtension(entry.Path, value.TargetExtension),
	)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "failed to rename incomplete transcode file")
	}

	entry.Path = utils.ReplaceExtension(entry.Path, value.TargetExtension)

	err = d.CompleteTranscoding(entry)
	if err != nil {
		return errors.Wrap(err, "failed to mark transcoding complete")
	}

	return nil
}

// rollbackIncompleteJob - Rollback the incomplete transcode job for the provided entry.
func (d *Database) rollbackIncompleteJob(entry value.Entry) error {
	log.WithFields(entry).Info("Rolling back incomplete job")

	err := os.Remove(strings.TrimSuffix(entry.Path, filepath.Ext(entry.Path)) + value.TranscodingExtension)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "failed to remove incomplete transcode file")
	}

	return d.cancelTranscoding(entry, false)
}

// addJob - Add a new job to the jobs table indicating the provided entry is going to be transcoded.
func (d *Database) addJob(db sqlite.Executable, entry value.Entry) error {
	log.WithFields(entry).Debug("Added job for entry")

	query := sqlite.Query{
		Query:     "insert into jobs (library_id, start_time) values (?, ?);",
		Arguments: []interface{}{entry.ID, time.Now().Unix()},
	}

	_, err := sqlite.ExecuteQuery(db, query)
	return err
}

// removeJob - Remove the job corresponding to the provided entry from the jobs table.
func (d *Database) removeJob(db sqlite.Executable, entry value.Entry) error {
	log.WithFields(entry).Debug("Removing job for entry")

	query := sqlite.Query{
		Query:     "delete from jobs where library_id = ?;",
		Arguments: []interface{}{entry.ID},
	}

	_, err := sqlite.ExecuteQuery(db, query)
	return err
}

// Close - Close the database, the database should not be used after it has been closed.
func (d *Database) Close() error {
	d.lock.Lock()
	defer d.lock.Unlock()

	log.Info("Closing database")

	defer func() {
		d.db = nil
		d.txns = 0
	}()

	return d.db.Close()
}

// Upsert - Update or insert the provided entry into the database; the entry will be updated in the event of a hash
// conflict.
func (d *Database) Upsert(entry value.Entry) error {
	return d.wrapTransaction(func(tx *sql.Tx) error {
		log.WithFields(entry).Info("Adding entry")

		query := sqlite.Query{
			Query: `insert or replace into library (path, discovered, transcoded, hash) values (?, ?, ?, ?)
				on conflict(hash) do update set path=excluded.path where path != excluded.path;`,
			Arguments: []interface{}{entry.Path, entry.Discovered, entry.Transcoded, entry.Hash},
		}

		_, err := sqlite.ExecuteQuery(tx, query)
		if err != nil {
			return errors.Wrap(err, "failed to execute query")
		}

		return nil
	})
}

// Remove - Remove the provided entry from the database; this will also remove any incomplete jobs for the provided
// entry.
func (d *Database) Remove(entry value.Entry) error {
	return d.wrapTransaction(func(tx *sql.Tx) error {
		log.WithFields(entry).Info("Removing entry")

		err := d.removeJob(tx, entry)
		if err != nil {
			return errors.Wrap(err, "failed to remove job")
		}

		query := sqlite.Query{
			Query:     "delete from library where id = ?;",
			Arguments: []interface{}{entry.ID},
		}

		_, err = sqlite.ExecuteQuery(tx, query)
		if err != nil {
			return errors.Wrap(err, "failed to execute query")
		}

		return nil
	})
}

// BeginTranscoding - Retrieve an untranscoded entry from the database, note that a job will be created for the provided
// entry which should be completed/cancelled (in the event of a failure, this will happen the next time the database is
// opened).
func (d *Database) BeginTranscoding() (value.Entry, error) {
	var entry value.Entry

	return entry, d.wrapTransaction(func(tx *sql.Tx) error {
		query := sqlite.Query{
			Query: "select library.id, path, hash from library where transcoded is null and " +
				"id not in (select library_id from jobs) order by discovered asc limit 1;",
		}

		err := sqlite.QueryRow(tx, query, &entry.ID, &entry.Path, &entry.Hash)
		if err != nil {
			return errors.Wrap(err, "failed to query database")
		}

		log.WithFields(entry).Info("Scheduling job to transcode entry")

		err = d.addJob(tx, entry)
		if err != nil {
			return errors.Wrap(err, "failed to add job")
		}

		return nil
	})
}

// CompleteTranscoding - Rehash and mark the provided entry as having been transcoded.
func (d *Database) CompleteTranscoding(entry value.Entry) error {
	hash, err := utils.HashFile(entry.Path)
	if err != nil {
		return errors.Wrap(err, "failed to hash file")
	}

	return d.wrapTransaction(func(tx *sql.Tx) error {
		query := sqlite.Query{
			Query:     "update library set path = ?, transcoded = ?, hash = ? where id = ?;",
			Arguments: []interface{}{entry.Path, utils.Int64P(time.Now().Unix()), hash, entry.ID},
		}

		_, err = sqlite.ExecuteQuery(tx, query)
		if err != nil {
			return errors.Wrap(err, "failed to update database")
		}

		log.WithFields(entry).Info("Completing job to transcode entry")

		err = d.removeJob(tx, entry)
		if err != nil {
			return errors.Wrapf(err, "failed to remove job %d", entry.ID)
		}

		return nil
	})
}

// CancelTranscoding - Cancel the job for the provided entry.
func (d *Database) CancelTranscoding(entry value.Entry) error {
	return d.cancelTranscoding(entry, true)
}

// cancelTranscoding - Convenience function to conditionally log, then remove the job for the provided entry.
func (d *Database) cancelTranscoding(entry value.Entry, shouldLog bool) error {
	return d.wrapTransaction(func(tx *sql.Tx) error {
		if shouldLog {
			log.WithFields(entry).Info("Cancelling job to transcode entry")
		}

		err := d.removeJob(tx, entry)
		if err != nil {
			return errors.Wrapf(err, "failed to remove job %d", entry.ID)
		}

		return nil
	})
}

// wrapTransaction - Run the provided callback within a transaction (correctly handling the completion/rollback).
func (d *Database) wrapTransaction(callback func(tx *sql.Tx) error) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	tx, err := d.beginLOCKED()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	err = callback(tx)
	if err == nil {
		err = d.commitLOCKED(tx)
		if err != nil {
			return errors.Wrap(err, "failed to commit transaction")
		}

		return nil
	}

	if !errors.Is(err, sqlite.ErrQueryReturnedNoRows) {
		log.WithError(err).Error("Unexpected error, rolling back transaction")
	}

	if err := d.rollbackLOCKED(tx); err != nil {
		return errors.Wrap(err, "failed to rollback transaction")
	}

	return err
}

// beginLOCKED - Utility function to log and being a new transaction.
func (d *Database) beginLOCKED() (*sql.Tx, error) {
	log.WithField("number", d.txns+1).Debug("Beginning transaction")
	return d.db.Begin()
}

// commitLOCKED - Utility function to log and commit the provided transaction.
func (d *Database) commitLOCKED(tx *sql.Tx) error {
	log.WithField("number", d.txns+1).Debug("Committing transaction")

	defer func() {
		d.txns++
	}()

	return tx.Commit()
}

// rollbackLOCKED - Utility function to log and rollback the provided transaction.
func (d *Database) rollbackLOCKED(tx *sql.Tx) error {
	log.WithField("number", d.txns+1).Debug("Rolling back transaction")

	defer func() {
		d.txns++
	}()

	return tx.Rollback()
}
