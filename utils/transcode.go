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
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"syscall"

	"github.com/jamesl33/goamt/value"

	"github.com/apex/log"
	"golang.org/x/sys/unix"
)

// LoudnormStats - Represents the raw stats from the first pass with the loudnorm filter which will be used in the
// second pass.
type LoudnormStats struct {
	MeasuredI         string `json:"input_i"`
	MeasuredTP        string `json:"input_tp"`
	MeasuredLRA       string `json:"input_lra"`
	MeasuredThreshold string `json:"input_thresh"`
	TargetOffset      string `json:"target_offset"`
}

// TranscodeFile - Use ffmpeg to transcode the file at the provided path, note that the resulting file will have the
// '.transcoding.mp4' extension.
func TranscodeFile(path string) error {
	lns, err := firstPass(path)
	if err != nil {
		return fmt.Errorf("failed to run first pass: %w", err)
	}

	err = secondPass(path, lns)
	if err != nil {
		return fmt.Errorf("failed to run second pass: %w", err)
	}

	return nil
}

// firstPass - Run the first pass, this doesn't perform any transcoding; it simply gets the loudnorm stats which will be
// used in the second pass the achieve the best normalisation results.
func firstPass(path string) (*LoudnormStats, error) {
	command := exec.Command(
		"ffmpeg",
		"-i",
		path,
		"-hide_banner",
		"-vn",
		"-af",
		"loudnorm=print_format=json",
		"-f",
		"null",
		"-",
	)

	command.SysProcAttr = &unix.SysProcAttr{
		Pdeathsig: syscall.SIGINT,
		Setpgid:   true,
	}

	fields := log.Fields{
		"path":    path,
		"command": command.String(),
	}

	log.WithFields(fields).Debugf("Running first pass")

	output, err := command.CombinedOutput()
	if err != nil {
		log.Errorf("%s", output)
		return nil, fmt.Errorf("failed to run 'ffmpeg': %s", err)
	}

	split := bytes.Split(output, []byte("\n"))
	stats := split[len(split)-13:]

	var lns *LoudnormStats
	err = json.Unmarshal(bytes.Join(stats, []byte("\n")), &lns)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal loudnorm stats: %w", err)
	}

	fields = log.Fields{
		"path":           path,
		"loudnorm_stats": lns,
	}

	log.WithFields(fields).Debugf("Completed first pass")

	return lns, nil
}

// secondPass - Run the second pass transcoding the input file using the loudnorm stats from the first pass.
func secondPass(path string, lns *LoudnormStats) error {
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
		"-af",
		fmt.Sprintf(
			"loudnorm=linear=true:measured_i=%s:measured_tp=%s:measured_lra=%s:measured_thresh=%s:offset=%s",
			lns.MeasuredI,
			lns.MeasuredTP,
			lns.MeasuredLRA,
			lns.MeasuredThreshold,
			lns.TargetOffset,
		),
		ReplaceExtension(path, value.TranscodingExtension),
	)

	command.SysProcAttr = &unix.SysProcAttr{
		Pdeathsig: syscall.SIGINT,
		Setpgid:   true,
	}

	fields := log.Fields{
		"path":    path,
		"command": command.String(),
	}

	log.WithFields(fields).Debugf("Running second pass")

	output, err := command.CombinedOutput()
	if err != nil {
		log.Errorf("%s", output)
		return fmt.Errorf("failed to run 'ffmpeg': %s", err)
	}

	return nil
}
