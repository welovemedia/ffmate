package debug

import (
	"io"
	"os"

	"github.com/yosev/debugo"
)

// loggers holds all log levels and namespaces as plain functions.
type loggers struct {
	Info  func(format string, v ...any)
	Debug func(format string, v ...any)
	Warn  func(format string, v ...any)
	Error func(format string, v ...any)
}

// newLoggers creates a set of loggers for a given namespace prefix.
func newLoggers(prefix string) loggers {
	makeLogger := func(level string) func(string, ...any) {
		ns := level
		if prefix != "" {
			ns = level + ":" + prefix
		}
		logger := debugo.New(ns)
		return func(format string, v ...any) {
			logger.Debugf(format, v...)
		}
	}

	return loggers{
		Info:  makeLogger("info"),
		Debug: makeLogger("debug"),
		Warn:  makeLogger("warn"),
		Error: makeLogger("error"),
	}
}

// Global loggers
var (
	Log         = newLoggers("") // no namespace
	Watchfolder = newLoggers("watchfolder")
	Task        = newLoggers("task")
	Client      = newLoggers("client")
	FFmpeg      = newLoggers("ffmpeg")
	Websocket   = newLoggers("websocket")
	Controller  = newLoggers("controller")
	Service     = newLoggers("service")
	Middleware  = newLoggers("middleware")
	Telemetry   = newLoggers("telemetry")
	Webhook     = newLoggers("webhook")
	HTTP        = newLoggers("http")
	Test        = newLoggers("test")
)

// logBroadcaster is used to forward all logs to websocket clients via callback.
type logBroadcaster struct {
	Callback func([]byte)
}

func (cw *logBroadcaster) Write(p []byte) (int, error) {
	cw.Callback(p)
	return len(p), nil
}

// RegisterBroadcastLogger sets up a multi-writer that sends logs to stderr and a callback.
func RegisterBroadcastLogger(fn func([]byte)) {
	mw := io.MultiWriter(os.Stderr, &logBroadcaster{Callback: fn})
	debugo.SetOutput(mw)
}
