package ffmpeg

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"unicode"

	"github.com/mattn/go-shellwords"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/debug"
	"github.com/welovemedia/ffmate/v2/internal/service"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

type FFmpegProgress struct {
	Bitrate string
	Speed   string
	Frame   int
	FPS     float64
	Time    float64
}

type ExecutionRequest struct {
	Ctx        context.Context
	Task       *model.Task
	UpdateFunc func(progress float64, remaining float64)
	Command    string
}

// Execute runs the ffmpeg command, provides progress updates, and checks the result
func (s *Service) Execute(request *ExecutionRequest) error {
	commands := strings.Split(request.Command, "&&")
	for index, cmdStr := range commands {
		cmdStr = strings.TrimSpace(cmdStr)
		var args []string
		var err error
		if runtime.GOOS == "windows" {
			args, err = s.shellwordsUnicodeSafe(cmdStr)
		} else {
			args, err = shellwords.NewParser().Parse(cmdStr)
		}
		if err != nil {
			return fmt.Errorf("FFMPEG - failed to parse command: %v", err)
		}
		args = append(args, "-progress", "pipe:2")
		if !strings.Contains(cmdStr, "-stats_period") {
			args = append(args, "-stats_period", "1")
		}
		var cmd *exec.Cmd
		if index > 0 {
			cmd = exec.CommandContext(request.Ctx, "", args...)
		} else {
			cmd = exec.CommandContext(request.Ctx, cfg.GetString("ffmate.ffmpeg"), args...)
		}

		var stderrBuf bytes.Buffer
		var duration float64

		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			return fmt.Errorf("FFMPEG - failed to get stderr pipe: %v", err)
		}

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("FFMPEG - failed to start ffmpeg: %v", err)
		}

		reDuration := regexp.MustCompile(`Duration: (\d+:\d+:\d+\.\d+)`)

		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			stderrBuf.WriteString(line + "\n")
			if match := reDuration.FindStringSubmatch(line); match != nil {
				durationStr := match[1]
				duration = s.parseDuration(durationStr)
			}
			if progress := s.parseFFmpegOutput(line, duration); progress != nil {
				p := math.Min(100, math.Round((progress.Time/duration*100)*100)/100)
				debug.FFmpeg.Debug("progress: %f %+v (uuid: %s)", p, progress, request.Task.UUID)
				remainingTime, err := progress.EstimateRemainingTime(duration)
				if err != nil {
					debug.FFmpeg.Debug("failed to estimate remaining time: %v", err)
					remainingTime = -1
				}
				request.UpdateFunc(p, remainingTime)
			}
		}
		if err := scanner.Err(); err != nil {
			debug.FFmpeg.Error("FFMPEG - error reading progress: %v\n", err)
		}

		err = cmd.Wait()
		stderr := stderrBuf.String()
		if err != nil {
			return errors.New(stderr)
		}
	}
	return nil
}

// EstimateRemainingTime calculates the estimated remaining time based on the current progress and speed.
func (p *FFmpegProgress) EstimateRemainingTime(duration float64) (float64, error) {
	speed, err := p.parseSpeed(p.Speed)
	if err != nil {
		return 0, err
	}
	remainingTime := (duration - p.Time) / speed
	return math.Round(remainingTime), nil
}

// parseSpeed parses the speed string and returns the speed as a float64.
func (p *FFmpegProgress) parseSpeed(speedStr string) (float64, error) {
	var speed float64
	_, err := fmt.Sscanf(speedStr, "%fx", &speed)
	if err != nil {
		return 0, err
	}
	return speed, nil
}

func (s *Service) parseDuration(duration string) float64 {
	parts := strings.Split(duration, ":")
	if len(parts) != 3 {
		return 0
	}

	hours, _ := strconv.ParseFloat(parts[0], 64)
	minutes, _ := strconv.ParseFloat(parts[1], 64)
	seconds, _ := strconv.ParseFloat(parts[2], 64)

	return hours*3600 + minutes*60 + seconds
}

func (s *Service) parseFFmpegOutput(line string, duration float64) *FFmpegProgress {
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
			_, _ = fmt.Sscanf(value, "%d", &progress.Frame)
		case "fps":
			_, _ = fmt.Sscanf(value, "%f", &progress.FPS)
		case "bitrate":
			progress.Bitrate = value
		case "time":
			progress.Time = s.parseDuration(value)
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

func (s *Service) shellwordsUnicodeSafe(input string) ([]string, error) {
	var args []string
	var current strings.Builder
	inQuotes := false
	var quoteChar rune

	for _, r := range input {
		switch {
		case unicode.IsSpace(r) && !inQuotes:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		case r == '"' || r == '\'':
			if inQuotes && r == quoteChar {
				inQuotes = false
			} else if !inQuotes {
				inQuotes = true
				quoteChar = r
			} else {
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	if inQuotes {
		return nil, fmt.Errorf("unclosed quote")
	}
	return args, nil
}

func (s *Service) Name() string {
	return service.FFMpeg
}
