package ffmpeg

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/welovemedia/ffmate/pkg/config"
)

// ExecuteFFmpeg runs the ffmpeg command, provides progress updates, and checks the result
func Execute(request *ExecutionRequest, updateFunc func(progress float64)) error {
	evaluateWildcards(request)

	args := strings.Split(request.Command, " ")
	args = append(args, "-progress", "pipe:2")
	cmd := exec.Command(config.Config().FFMpeg, args...)

	// Buffers for capturing full stderr
	var stderrBuf bytes.Buffer
	var lastLine string
	var duration float64

	// Stderr pipe for real-time progress parsing
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("FFMPEG - failed to get stderr pipe: %v", err)
	}

	// Start the ffmpeg process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("FFMPEG - failed to start ffmpeg: %v", err)
	}

	// Regex to extract the duration field
	reDuration := regexp.MustCompile(`Duration: (\d+:\d+:\d+\.\d+)`)

	// Parse progress in real-time
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			stderrBuf.WriteString(line + "\n")
			lastLine = line
			if match := reDuration.FindStringSubmatch(line); match != nil {
				durationStr := match[1]
				duration = parseDuration(durationStr)
			}
			if progress := parseFFmpegOutput(line, duration); progress != nil {
				request.Logger.Debugf("FFMPEG - progress: %f %+v (uuid: %s)\n", progress.Time/duration*100, progress, request.Task.Uuid)
				updateFunc(progress.Time / duration * 100)
			}
		}
		if err := scanner.Err(); err != nil {
			request.Logger.Warnf("FFMPEG - error reading progress: %v\n", err)
		}
	}()

	// Wait for the ffmpeg process to complete
	err = cmd.Wait()

	// Gather full output for final reporting
	stderr := stderrBuf.String()

	if err != nil {
		return errors.New(stderr)
	}

	fmt.Sprintf("last line: %s", lastLine)

	return nil
}
