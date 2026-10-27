package task

import (
	"reflect"
	"testing"

	"github.com/mattn/go-shellwords"
)

// TestQuoteArgPreventsArgumentInjection feeds attacker-shaped file values through
// quoteArg into a command template and parses it with the same parser used to
// build the real ffmpeg and script command lines. Each value must come back as a
// single argument, with no extra tokens injected (CWE-88).
func TestQuoteArgPreventsArgumentInjection(t *testing.T) {
	cases := []string{
		`in.mp4`,
		`a file with spaces.mp4`,
		`in.mp4" -i rtsp://attacker/collect -loglevel "quiet`,
		`it's a quote.mp4`,
		`both'and"quotes.mp4`,
		`; rm -rf / ;`,
		`$(touch pwned).mp4`,
		"`touch pwned`.mp4",
	}
	for _, v := range cases {
		cmd := "-i " + quoteArg(v) + " -y out.mp4"
		args, err := shellwords.NewParser().Parse(cmd)
		if err != nil {
			t.Errorf("Parse(%q) returned error: %v", cmd, err)
			continue
		}
		want := []string{"-i", v, "-y", "out.mp4"}
		if !reflect.DeepEqual(args, want) {
			t.Errorf("quoteArg(%q): %q parsed to %#v, want %#v (argument injection or value corruption)", v, cmd, args, want)
		}
	}
}

// TestQuoteArgRoundTrip checks that a quoted value always parses back to exactly
// the original single token, including values with spaces, both quote kinds, and
// Windows-style backslash paths.
func TestQuoteArgRoundTrip(t *testing.T) {
	for _, v := range []string{`plain`, `a b`, `quote'inside`, `dq"inside`, `C:\dir\file.mkv`} {
		args, err := shellwords.NewParser().Parse(quoteArg(v))
		if err != nil {
			t.Fatalf("Parse(quoteArg(%q)) returned error: %v", v, err)
		}
		if len(args) != 1 || args[0] != v {
			t.Errorf("quoteArg(%q) round-tripped to %#v, want single token %q", v, args, v)
		}
	}
}
