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
	"slices"
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
	// commands can be chained with && (e.g. two-pass encoding), every part is an ffmpeg call
	commands, err := splitCommandChain(request.Command, runtime.GOOS == "windows")
	if err != nil {
		return fmt.Errorf("FFMPEG - failed to parse command: %v", err)
	}
	ffmpegPath := cfg.GetString("ffmate.ffmpeg")
	for _, cmdStr := range commands {
		var args []string
		if runtime.GOOS == "windows" {
			args, err = s.shellwordsUnicodeSafe(cmdStr)
		} else {
			args, err = shellwords.NewParser().Parse(cmdStr)
		}
		if err != nil {
			return fmt.Errorf("FFMPEG - failed to parse command: %v", err)
		}
		// chained commands start with ${FFMPEG} (as documented), the binary is always the configured ffmpeg
		if len(args) > 0 && args[0] == ffmpegPath {
			args = args[1:]
		}
		args = append(args, "-progress", "pipe:2")
		if !slices.Contains(args, "-stats_period") {
			args = append(args, "-stats_period", "1")
		}
		cmd := exec.CommandContext(request.Ctx, ffmpegPath, args...)

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

// splitCommandChain splits a command at && operators that are not quoted, so a quoted value
// containing && (e.g. a file name) stays intact. Backslash escapes are honored outside of single
// quotes, except on windows where a backslash is a path separator (matching the respective parsers).
func splitCommandChain(command string, windows bool) ([]string, error) {
	var parts []string
	var current strings.Builder
	var quote rune
	escaped := false
	runes := []rune(command)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		switch {
		case escaped:
			escaped = false
		case r == '\\' && !windows && quote != '\'':
			escaped = true
		case quote != 0:
			if r == quote {
				quote = 0
			}
		case r == '\'' || r == '"':
			quote = r
		case r == '&' && i+1 < len(runes) && runes[i+1] == '&':
			parts = append(parts, strings.TrimSpace(current.String()))
			current.Reset()
			i++ // skip the second &
			continue
		}
		current.WriteRune(r)
	}
	if quote != 0 {
		return nil, errors.New("unclosed quote")
	}
	parts = append(parts, strings.TrimSpace(current.String()))
	for _, part := range parts {
		if part == "" {
			return nil, errors.New("empty command in && chain")
		}
	}
	return parts, nil
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
