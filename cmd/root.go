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
	"github.com/spf13/cobra"
)

// rootCommand - Represents the root goamt command and encapsulates all the supported sub-commands.
var rootCommand = &cobra.Command{
	Short:         "An automatic media transcoder written in Go with an emphasis on ease of management and performance",
	SilenceErrors: true,
	SilenceUsage:  true,
}

// init - Initialize the root command by adding all the supported sub-commands.
func init() {
	rootCommand.AddCommand(versionCommand, convertCommand, createCommand, updateCommand, transcodeCommand)
}

// Execute - Execute goamt, returning any errors raised during the operation of the chosen sub-command.
func Execute() error {
	return rootCommand.Execute()
}
