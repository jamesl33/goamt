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

package sqlite

import (
	"database/sql"
)

// Query - Encapsulates the options for an SQLite query.
type Query struct {
	Query     string
	Arguments []interface{}
}

// Queryable - Narrow interface which represents a queryable struct e.g. *sql.DB or *sql.Tx.
type Queryable interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

// Executable - Narrow interface which represents an executable struct e.g. *sql.DB or *sql.Tx.
type Executable interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// ScanCallback - Readability wrapper around the sql.Scan function.
type ScanCallback func(dest ...interface{}) error

// RowCallback - The function which will be will run for each row returned by a query.
type RowCallback func(scan ScanCallback) error

// ExecuteQuery - Execute the given query against the provided database returning the number of rows affected.
func ExecuteQuery(db Executable, query Query) (int64, error) {
	res, err := db.Exec(query.Query, query.Arguments...)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}

// QueryRow - Utility function to execute a query which is only expected to return a single row. Note that in the event
// that more than one row is returned, only the first will be scanned.
func QueryRow(db Queryable, query Query, dest ...interface{}) error {
	rows, err := db.Query(query.Query, query.Arguments...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		if rows.Err() != nil {
			return rows.Err()
		}

		return ErrQueryReturnedNoRows
	}

	return rows.Scan(dest...)
}

// QueryRows - Utility function to execute a query an run the provided callback for each row returned.
func QueryRows(db Queryable, query Query, callback RowCallback) error {
	rows, err := db.Query(query.Query, query.Arguments...)
	if err != nil {
		return err
	}
	defer rows.Close()

	var containedRows bool
	for rows.Next() {
		err = callback(rows.Scan)
		if err != nil {
			return err
		}

		containedRows = true
	}

	if !containedRows {
		return ErrQueryReturnedNoRows
	}

	return rows.Err()
}
