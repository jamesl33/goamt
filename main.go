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

package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/jamesl33/goamt/cmd"
	"github.com/jamesl33/goamt/utils"

	"github.com/apex/log"
	"github.com/pkg/errors"
)

// main - Setup logging, then execute goamt.
func main() {
	log.SetHandler(utils.NewLoggingHandler())

	level, err := log.ParseLevel(os.Getenv("GOAMT_LOG_LEVEL"))
	if err != nil {
		level = log.DebugLevel
	}

	log.SetLevel(level)

	err = cmd.Execute()
	if err == nil {
		return
	}

	// The sub-command failed for some reason, ensure that we exit with a non-zero exit code
	defer os.Exit(1)

	stacktrace := os.Getenv("GOAMT_DISPLAY_STACKTRACE")
	if display, parseError := strconv.ParseBool(stacktrace); parseError == nil && display {
		fmt.Printf("Error: %+v\n", err)
	} else {
		fmt.Printf("Error: %s\n", errors.Cause(err))
	}
}
