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

const (
	// TargetExtension - The target extension for transcoded files i.e. we will be created mp4 files (ffmpeg uses the
	// extension to determine the target format).
	TargetExtension = ".mp4"

	// TranscodingExtension - The extension used for files which are being transcoded; this is a temporary extension
	// which will be renamed to the target extension upon completion.
	TranscodingExtension = ".transcoding" + TargetExtension
)

// SupportedExtensions - The list of extensions supported by goamt i.e. the files that will be detected by the update
// sub-command (all other files will be ignored).
var SupportedExtensions = []string{".mp4", ".mkv", ".avi"}
