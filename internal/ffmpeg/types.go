package ffmpeg

import (
	"github.com/sirupsen/logrus"
	"github.com/welovemedia/ffmate/internal/database/model"
)

// FFmpegProgress holds parsed progress data
type FFmpegProgress struct {
	Frame   int
	FPS     float64
	Bitrate string
	Time    float64
	Speed   string
}

type ExecutionRequest struct {
	Task *model.Task

	Command string

	Logger *logrus.Logger
}
