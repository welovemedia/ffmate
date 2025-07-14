package ffmpeg

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"math"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/mattn/go-shellwords"
	"github.com/welovemedia/ffmate/internal/config"
	"github.com/yosev/debugo"
)

var debug = debugo.New("ffmpeg")

// ExecuteFFmpeg runs the ffmpeg command, provides progress updates, and checks the result
func Execute(request *ExecutionRequest) error {
	commands := strings.Split(request.Command, "&&")
	for index, cmdStr := range commands {
		cmdStr = strings.TrimSpace(cmdStr)
		var args []string
		var err error
		if runtime.GOOS == "windows" {
			args, err = shellwordsUnicodeSafe(cmdStr)
		} else {
			args, err = shellwords.NewParser().Parse(cmdStr)
		}
		if err != nil {
			return fmt.Errorf("FFMPEG - failed to parse command: %v", err)
		}
		args = append(args, "-progress", "pipe:2")
		config.Config().Mutex.RLock()
		var cmd *exec.Cmd
		if index > 0 {
			cmd = exec.CommandContext(request.Ctx, "", args...)
		} else {
			cmd = exec.CommandContext(request.Ctx, config.Config().FFMpeg, args...)
		}
		config.Config().Mutex.RUnlock()

		var stderrBuf bytes.Buffer
		var lastLine string
		var duration float64

		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			return fmt.Errorf("FFMPEG - failed to get stderr pipe: %v", err)
		}

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("FFMPEG - failed to start ffmpeg: %v", err)
		}

		reDuration := regexp.MustCompile(`Duration: (\d+:\d+:\d+\.\d+)`)

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
					p := math.Min(100, math.Round((progress.Time/duration*100)*100)/100)
					debug.Debugf("progress: %f %+v (uuid: %s)", p, progress, request.Task.Uuid)
					remainingTime, err := progress.EstimateRemainingTime(duration)
					if err != nil {
						debug.Debugf("failed to estimate remaining time: %v", err)
						remainingTime = -1
					}
					request.UpdateFunc(p, remainingTime)
				}
			}
			if err := scanner.Err(); err != nil {
				request.Logger.Warnf("FFMPEG - error reading progress: %v\n", err)
			}
		}()

		err = cmd.Wait()
		stderr := stderrBuf.String()
		if err != nil {
			return errors.New(stderr)
		}
		fmt.Sprintf("last line: %s", lastLine)
	}
	return nil
}
