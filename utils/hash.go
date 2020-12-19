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

package utils

import (
	"hash/crc32"
	"io"
	"os"

	"github.com/pkg/errors"
)

const (
	// BufferSize - The amount of data read from disk before seeking to the next location in the file.
	BufferSize = 4096

	// MaxSeekSize - The max size of a seek operation, this has the affect of reading 'BufferSize' amount of data for
	// every 'MaxSeekSize' until we reach the end of the file.
	MaxSeekSize = 64 * 1024 * 1024
)

// table - IEEE CRC32 table, use a global variable to avoid atomic operations in 'MakeTable' function.
var table = crc32.MakeTable(crc32.IEEE)

// HashFile - Open then hash the file at the provided path.
func HashFile(path string) (uint32, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, errors.Wrap(err, "failed to open hash file")
	}
	defer file.Close()

	return hashReader(file)
}

// hashReader - Return the CRC32 hash of the provided ReadSeeker.
func hashReader(reader io.ReadSeeker) (uint32, error) {
	var (
		buffer [BufferSize]byte
		digest uint32
	)

	for {
		n, err := reader.Read(buffer[:])
		if err != nil {
			if n == 0 {
				return digest, nil
			}

			return 0, errors.Wrap(err, "failed to read from hash file")
		}

		digest = crc32.Update(digest, table, buffer[:n])

		_, err = reader.Seek(int64(digest%MaxSeekSize), io.SeekCurrent)
		if err != nil {
			return 0, errors.Wrap(err, "failed to seek to next offset")
		}
	}
}
