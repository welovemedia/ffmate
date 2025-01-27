//go:build linux || darwin

package ffmpeg

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"math"
	"os/exec"
	"regexp"
	"strings"

	"github.com/mattn/go-shellwords"
	"github.com/welovemedia/ffmate/internal/config"
	"github.com/yosev/debugo"
)

var debug = debugo.New("ffmpeg")

// ExecuteFFmpeg runs the ffmpeg command, provides progress updates, and checks the result
func Execute(request *ExecutionRequest) error {
	// args := SplitCommand(request.Command)
	args, err := shellwords.NewParser().Parse(request.Command)
	if err != nil {
		return fmt.Errorf("FFMPEG - failed to parse command: %v", err)
	}
	args = append(args, "-progress", "pipe:2")
	cmd := exec.CommandContext(request.Ctx, config.Config().FFMpeg, args...)

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
				p := math.Round((progress.Time/duration*100)*100) / 100
				debug.Debugf("progress: %f %+v (uuid: %s)", p, progress, request.Task.Uuid)
				request.UpdateFunc(p)
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

func SplitCommand(cmd string) []string {
	var args []string
	var current strings.Builder
	var escaping, inQuotes bool

	for _, r := range cmd {
		switch {
		case escaping:
			current.WriteRune(r)
			escaping = false
		case r == '\\':
			escaping = true
		case r == '"':
			inQuotes = !inQuotes
		case r == ' ' && !inQuotes:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}
