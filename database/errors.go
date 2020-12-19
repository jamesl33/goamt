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
	"fmt"
)

// ErrUnknownVersion - Returned when the user attempts to open a database with an unknown version.
type ErrUnknownVersion struct {
	what, where string
}

func (e *ErrUnknownVersion) Error() string {
	return fmt.Sprintf("%s at '%s' is an unknown version", e.what, e.where)
}

// ErrAlreadyExists - Returned when the user attempts to create a database which already exists.
type ErrAlreadyExists struct {
	what, where string
}

func (e *ErrAlreadyExists) Error() string {
	return fmt.Sprintf("%s at '%s' already exists", e.what, e.where)
}

// ErrNotFound - Returned when the user attempts to open a database which doesn't exist.
type ErrNotFound struct {
	what, where string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("%s at '%s' not found", e.what, e.where)
}
