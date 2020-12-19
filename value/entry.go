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

package value

import (
	"github.com/apex/log"
)

// Entry - Represents an entry in the SQLite database, used when interacting with a '*database.Database'.
type Entry struct {
	ID         int
	Path       string
	Discovered int64
	Transcoded *int64
	Hash       uint32
}

// Fields - Implement the fielder interface for the apex log module, note that fields with a default value will be
// omitted.
func (e Entry) Fields() log.Fields {
	fields := make(log.Fields)

	if e.ID != 0 {
		fields["id"] = e.ID
	}

	if e.Path != "" {
		fields["path"] = e.Path
	}

	if e.Discovered != 0 {
		fields["discovered"] = e.Discovered
	}

	if e.Transcoded != nil {
		fields["transcoded"] = e.Transcoded
	}

	if e.Hash != 0 {
		fields["transcoded"] = e.Hash
	}

	return fields
}
