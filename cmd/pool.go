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
	"sync"

	"github.com/jamesl33/goamt/database"
	"github.com/jamesl33/goamt/utils"
	"github.com/jamesl33/goamt/value"
)

// transcodeFunc - The function used by the worker pool when transcoding entries, used to allow unit testing of the
// worker pool.
var transcodeFunc = utils.TranscodeFile

// Pool - Worker pool which concurrently updates/transcodes entries (depending on which constructor is used).
type Pool struct {
	entryStream chan value.Entry
	errorStream chan error
	wg          sync.WaitGroup
	db          *database.Database
	consume     func(db *database.Database, entry value.Entry) error
	drain       func(db *database.Database, entry value.Entry) error
}

// NewUpdatePool - Create a new worker pool which will hash and upsert entries into the provided database.
func NewUpdatePool(db *database.Database) *Pool {
	return &Pool{
		db:      db,
		consume: upsertEntry,
		drain:   func(_ *database.Database, _ value.Entry) error { return nil },
	}
}

// NewTranscodePool - Create a new worker pool which will transcode entries from the provided database.
func NewTranscodePool(db *database.Database) *Pool {
	return &Pool{
		db:      db,
		consume: transcodeEntry,
		drain:   cancelTranscoding,
	}
}

// Start - Spawn 'threads' number of workers to process entries queued in the returned entry channel.
func (p *Pool) Start(ctx context.Context, threads int) (chan<- value.Entry, <-chan error) {
	p.entryStream = make(chan value.Entry, 1024)
	p.errorStream = make(chan error, threads)

	for w := 0; w < threads; w++ {
		p.wg.Add(1)

		go func() {
			defer p.wg.Done()

			for entry := range p.entryStream {
				err := p.consume(p.db, entry)
				if err != nil {
					p.errorStream <- err
					return
				}

				if ctx.Err() != nil {
					return
				}
			}
		}()
	}

	return p.entryStream, p.errorStream
}

// Stop - Gracefully stop the worker pool, draining 'entryStream' in the event that the user interrupted goamt during
// the convert/update/transcode sub-command.
func (p *Pool) Stop() error {
	close(p.entryStream)
	p.wg.Wait()

	if len(p.errorStream) != 0 {
		return <-p.errorStream
	}

	for entry := range p.entryStream {
		err := p.drain(p.db, entry)
		if err != nil {
			return err
		}
	}

	return nil
}
