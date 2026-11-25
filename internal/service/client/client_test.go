package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
)

func TestFFmpegPathEmptyWhenNotFound(t *testing.T) {
	defer cfg.Set("ffmate.isFFmpeg", false)
	defer cfg.Set("ffmate.ffmpeg", "ffmpeg")

	// lookup failed: the configured (unresolved) value must not be reported
	cfg.Set("ffmate.ffmpeg", "/nonexistent/ffmpeg")
	cfg.Set("ffmate.isFFmpeg", false)
	assert.Empty(t, ffmpegPath())

	// lookup succeeded: the resolved path is reported
	cfg.Set("ffmate.ffmpeg", "/usr/bin/ffmpeg")
	cfg.Set("ffmate.isFFmpeg", true)
	assert.Equal(t, "/usr/bin/ffmpeg", ffmpegPath())
}
