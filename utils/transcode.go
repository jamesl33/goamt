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
	"fmt"
	"os/exec"
	"syscall"

	"github.com/jamesl33/goamt/value"

	"github.com/apex/log"
	"golang.org/x/sys/unix"
)

// TranscodeFile - Use ffmpeg to transcode the file at the provided path, note that the resulting file will have the
// '.transcoding.mp4' extension.
func TranscodeFile(path string) error {
	command := exec.Command(
		"ffmpeg",
		"-i",
		path,
		"-map_chapters", "-1",
		"-map_metadata", "-1",
		"-metadata:s:a", "language=eng",
		"-metadata:s:v", "language=eng",
		"-sn",
		"-profile:v", "high",
		"-level:v", "4.0",
		"-pix_fmt", "yuv420p",
		"-acodec", "aac",
		"-vcodec", "h264",
		ReplaceExtension(path, value.TranscodingExtension),
	)

	command.SysProcAttr = &unix.SysProcAttr{
		Pdeathsig: syscall.SIGINT,
		Setpgid:   true,
	}

	output, err := command.CombinedOutput()
	if err != nil {
		log.Errorf("%s", output)
		return fmt.Errorf("failed to run 'ffmpeg': %s", err)
	}

	return nil
}
