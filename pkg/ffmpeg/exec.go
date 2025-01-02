package ffmpeg

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/welovemedia/ffmate/pkg/database/model"
)

// FFmpegProgress holds parsed progress data
type FFmpegProgress struct {
	Frame   int
	FPS     float64
	Bitrate string
	Time    float64
	Speed   string
}

type ExceutionRequest struct {
	Task *model.Task

	Command    string
	InputFile  string
	OutputFile string

	Logger *logrus.Logger
}

func evaluateWildcards(request *ExceutionRequest) {
	request.Command = strings.ReplaceAll(request.Command, "${INPUT_FILE}", request.InputFile)
	request.Command = strings.ReplaceAll(request.Command, "${OUTPUT_FILE}", request.OutputFile)
}

// ExecuteFFmpeg runs the ffmpeg command, provides progress updates, and checks the result
func Execute(request *ExceutionRequest, updateFunc func(progress float64)) error {
	evaluateWildcards(request)

	args := strings.Split(request.Command, " ")
	args = append(args, "-progress", "pipe:2")
	cmd := exec.Command("ffmpeg", args...)

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
				request.Logger.Debugf("FFMPEG - progress: %+v (uuid: %s)\n", progress, request.Task.Uuid)
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

func parseDuration(duration string) float64 {
	parts := strings.Split(duration, ":")
	if len(parts) != 3 {
		return 0
	}

	hours, _ := strconv.ParseFloat(parts[0], 64)
	minutes, _ := strconv.ParseFloat(parts[1], 64)
	seconds, _ := strconv.ParseFloat(parts[2], 64)

	return hours*3600 + minutes*60 + seconds
}

func parseFFmpegOutput(line string, duration float64) *FFmpegProgress {
	if !strings.Contains(line, "frame=") {
		return nil
	}

	progress := &FFmpegProgress{}
	pairs := strings.Fields(line)
	reKeyValue := regexp.MustCompile(`(\w+)=([\w:./]+)`)
	for _, pair := range pairs {
		matches := reKeyValue.FindStringSubmatch(pair)
		if len(matches) != 3 {
			continue
		}
		key := matches[1]
		value := matches[2]

		switch key {
		case "frame":
			fmt.Sscanf(value, "%d", &progress.Frame)
		case "fps":
			fmt.Sscanf(value, "%f", &progress.FPS)
		case "bitrate":
			progress.Bitrate = value
		case "time":
			progress.Time = parseDuration(value)
		case "speed":
			progress.Speed = value
		}
	}
	if progress.Frame == 0 {
		return nil
	}
	if progress.Time == 0 {
		return &FFmpegProgress{Frame: progress.Frame, FPS: 0, Bitrate: "0kbit/s", Time: duration, Speed: "0x"}
	}
	return progress
}
