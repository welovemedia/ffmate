package task

import (
	"reflect"
	"strings"
	"testing"

	"github.com/mattn/go-shellwords"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/internal/dto"
)

// TestWildcardReplacerCommandQuotesInjectedFilename exercises the command path
// (wildcardReplacerCommand builds the ffmpeg/script command line) with an
// attacker-influenced input filename, the way a watchfolder drop would, and
// asserts the rendered command parses into exactly the intended argv with
// nothing injected (CWE-88).
func TestWildcardReplacerCommandQuotesInjectedFilename(t *testing.T) {
	cfg.Set("ffmate.ffmpeg", "ffmpeg")
	s := &Service{}

	inputFile := `in.mp4" -i rtsp://attacker/collect -loglevel "quiet`
	outputFile := "out.mkv"

	got := s.wildcardReplacerCommand("-i ${INPUT_FILE} -y ${OUTPUT_FILE}", inputFile, outputFile, dto.WATCHFOLDER, nil)

	args, err := shellwords.NewParser().Parse(got)
	if err != nil {
		t.Fatalf("Parse(%q) returned error: %v", got, err)
	}
	want := []string{"-i", inputFile, "-y", outputFile}
	if !reflect.DeepEqual(args, want) {
		t.Errorf("wildcardReplacerCommand rendered %q, which parsed to %#v, want %#v (argument injection)", got, args, want)
	}
}

// TestWildcardReplacerCommandExpandsAllWildcards runs every supported wildcard
// through wildcardReplacerCommand (including the metadata branch and the
// brace-fixed ${OUTPUT_FILE_EXTENSION}) and checks that none is left unexpanded
// and that no stray brace survives from a malformed key.
func TestWildcardReplacerCommandExpandsAllWildcards(t *testing.T) {
	cfg.Set("ffmate.ffmpeg", "ffmpeg")
	s := &Service{}
	meta := dto.MetadataMap{"profile": "hq"}

	tmpl := strings.Join([]string{
		"${INPUT_FILE}", "${INPUT_FILE_BASE}", "${INPUT_FILE_EXTENSION}", "${INPUT_FILE_BASENAME}", "${INPUT_FILE_DIR}",
		"${OUTPUT_FILE}", "${OUTPUT_FILE_BASE}", "${OUTPUT_FILE_EXTENSION}", "${OUTPUT_FILE_BASENAME}", "${OUTPUT_FILE_DIR}",
		"${FFMPEG}", "${SOURCE}", "${METADATA_profile}",
	}, " ")

	got := s.wildcardReplacerCommand(tmpl, "/in/clip.mov", "/out/clip.mp4", dto.API, &meta)

	if strings.Contains(got, "${") {
		t.Errorf("wildcardReplacerCommand left an unexpanded wildcard: %q", got)
	}
	// With the ${OUTPUT_FILE_EXTENSION} key brace-fixed, no stray closing brace
	// should remain; every template token here is a wildcard and expands to a
	// value that contains no braces.
	if strings.Contains(got, "}") {
		t.Errorf("stray brace remained (malformed wildcard key): %q", got)
	}
	if !strings.Contains(got, "hq") {
		t.Errorf("metadata wildcard ${METADATA_profile} not expanded: %q", got)
	}
}

// TestWildcardReplacerValueModeStaysPlain checks that the value path
// (wildcardReplacer, used to resolve the input/output/sidecar paths that are
// shown in the API and passed to os.MkdirAll/os.WriteFile) does NOT shell-quote
// the substituted value. Quoting there would put literal quote characters into a
// real filesystem path.
func TestWildcardReplacerValueModeStaysPlain(t *testing.T) {
	s := &Service{}

	inputFile := `/tmp/wf/big_buck_bunny_1080p_h264(2) copy.mp4`
	got := s.wildcardReplacer("${INPUT_FILE}", inputFile, "out.mov", dto.WATCHFOLDER, nil)

	if got != inputFile {
		t.Errorf("value mode altered the path: got %q, want %q", got, inputFile)
	}
	if strings.ContainsAny(got, `'"`) {
		t.Errorf("value mode must not add shell quotes (it feeds real paths to MkdirAll/WriteFile): %q", got)
	}
}

// TestWildcardReplacerNoDoubleQuoteMultiPass reproduces the real three-pass flow
// in prepareTaskFiles (resolve input value, resolve output value, then build the
// command from those resolved values). The output path is itself a wildcard
// template, which is the common case. The regression: quoting at every pass
// quoted the output value once while resolving it and again while building the
// command, producing the '"'"' soup reported on the PR and a corrupted path.
func TestWildcardReplacerNoDoubleQuoteMultiPass(t *testing.T) {
	cfg.Set("ffmate.ffmpeg", "ffmpeg")
	s := &Service{}

	cases := []struct {
		name        string
		inputRaw    string
		outputRaw   string
		commandRaw  string
		wantOutput  string   // resolved output must be this plain path (MkdirAll-safe)
		wantArgs    []string // command must parse to exactly this argv
		wantNoApost bool     // output with no apostrophe must contain no '"'"' idiom
	}{
		{
			name:        "spaces and parens, templated output",
			inputRaw:    `/tmp/wf/big_buck_bunny_1080p_h264(2) copy.mp4`,
			outputRaw:   `${INPUT_FILE_DIR}/${INPUT_FILE_BASENAME}.mov`,
			commandRaw:  `-y -i ${INPUT_FILE} ${OUTPUT_FILE}`,
			wantOutput:  `/tmp/wf/big_buck_bunny_1080p_h264(2) copy.mov`,
			wantArgs:    []string{"-y", "-i", `/tmp/wf/big_buck_bunny_1080p_h264(2) copy.mp4`, `/tmp/wf/big_buck_bunny_1080p_h264(2) copy.mov`},
			wantNoApost: true,
		},
		{
			name:        "apostrophe in filename, templated output",
			inputRaw:    `/tmp/wf/Bob's clip.mp4`,
			outputRaw:   `${INPUT_FILE_DIR}/${INPUT_FILE_BASENAME}.mov`,
			commandRaw:  `-y -i ${INPUT_FILE} ${OUTPUT_FILE}`,
			wantOutput:  `/tmp/wf/Bob's clip.mov`,
			wantArgs:    []string{"-y", "-i", `/tmp/wf/Bob's clip.mp4`, `/tmp/wf/Bob's clip.mov`},
			wantNoApost: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Mirrors prepareTaskFiles lines: resolve values, then build command.
			inResolved := s.wildcardReplacer(tc.inputRaw, tc.inputRaw, tc.outputRaw, dto.WATCHFOLDER, nil)
			outResolved := s.wildcardReplacer(tc.outputRaw, tc.inputRaw, tc.outputRaw, dto.WATCHFOLDER, nil)
			cmdResolved := s.wildcardReplacerCommand(tc.commandRaw, inResolved, outResolved, dto.WATCHFOLDER, nil)

			// Resolved output is a real filesystem path: it must stay plain so
			// os.MkdirAll(filepath.Dir(...)) does not create a quoted directory.
			if outResolved != tc.wantOutput {
				t.Errorf("resolved output path = %q, want %q", outResolved, tc.wantOutput)
			}
			// When the filename itself has no quote characters, the resolved path
			// must have none either; a quote here would mean a shell quote leaked
			// into a real path handed to os.MkdirAll. (The apostrophe case keeps
			// its genuine apostrophe, which the exact-match check above covers.)
			if tc.wantNoApost && strings.ContainsAny(outResolved, `'"`) {
				t.Errorf("resolved output path contains shell quotes (would corrupt MkdirAll): %q", outResolved)
			}

			// The '"'"' idiom must never appear for a value that has no
			// apostrophe; its presence there is the double-quoting bug.
			if tc.wantNoApost && strings.Contains(cmdResolved, `'"'"'`) {
				t.Errorf("command double-quoted (the '\"'\"' soup regression): %q", cmdResolved)
			}

			args, err := shellwords.NewParser().Parse(cmdResolved)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", cmdResolved, err)
			}
			if !reflect.DeepEqual(args, tc.wantArgs) {
				t.Errorf("command %q parsed to %#v, want %#v", cmdResolved, args, tc.wantArgs)
			}
		})
	}
}
