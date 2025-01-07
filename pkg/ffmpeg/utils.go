package ffmpeg

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

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
