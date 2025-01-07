package ffmpeg

import (
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

type ExecutionRequest struct {
	Task *model.Task

	Command    string
	InputFile  string
	OutputFile string

	Logger *logrus.Logger
}
