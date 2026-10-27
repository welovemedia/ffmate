package task

import (
	"encoding/json"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/internal/dto"
)

// wildcardReplacer resolves wildcards into a plain value, used for a resolved
// input, output or sidecar path. The result is consumed as-is: shown in the API
// and passed to os.MkdirAll/os.WriteFile as a real filesystem path, so file-path
// wildcards must NOT be shell-quoted here.
func (s *Service) wildcardReplacer(input string, inputFile string, outputFile string, source dto.TaskSource, metadata *dto.MetadataMap) string {
	return s.replaceWildcards(input, inputFile, outputFile, source, metadata, false)
}

// wildcardReplacerCommand resolves wildcards into a command line that is later
// re-parsed into argv (the ffmpeg command and the pre/post-processing script).
// File-path wildcards are wrapped with quoteArg so an attacker-influenced value
// (for example an input filename taken from a watchfolder) stays a single
// argument and cannot inject extra arguments (CWE-88).
//
// Quoting is applied only here, never when resolving the input/output/sidecar
// path values above, because those resolved values are themselves substituted
// back into the command. Quoting at both stages would quote the value twice and
// corrupt the path.
func (s *Service) wildcardReplacerCommand(input string, inputFile string, outputFile string, source dto.TaskSource, metadata *dto.MetadataMap) string {
	return s.replaceWildcards(input, inputFile, outputFile, source, metadata, true)
}

func (s *Service) replaceWildcards(input string, inputFile string, outputFile string, source dto.TaskSource, metadata *dto.MetadataMap, quote bool) string {
	// q quotes file-path wildcards only when the result is a command line that
	// gets re-parsed into argv; for a plain resolved value it returns v unchanged.
	q := func(v string) string {
		if quote {
			return quoteArg(v)
		}
		return v
	}

	input = strings.ReplaceAll(input, "${INPUT_FILE}", q(inputFile))
	input = strings.ReplaceAll(input, "${OUTPUT_FILE}", q(outputFile))

	input = strings.ReplaceAll(input, "${INPUT_FILE_BASE}", q(filepath.Base(inputFile)))
	input = strings.ReplaceAll(input, "${OUTPUT_FILE_BASE}", q(filepath.Base(outputFile)))
	input = strings.ReplaceAll(input, "${INPUT_FILE_EXTENSION}", q(filepath.Ext(filepath.Base(inputFile))))
	input = strings.ReplaceAll(input, "${OUTPUT_FILE_EXTENSION}", q(filepath.Ext(filepath.Base(outputFile))))
	input = strings.ReplaceAll(input, "${INPUT_FILE_BASENAME}", q(strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(filepath.Base(inputFile)))))
	input = strings.ReplaceAll(input, "${OUTPUT_FILE_BASENAME}", q(strings.TrimSuffix(filepath.Base(outputFile), filepath.Ext(filepath.Base(outputFile)))))
	input = strings.ReplaceAll(input, "${INPUT_FILE_DIR}", q(filepath.Dir(inputFile)))
	input = strings.ReplaceAll(input, "${OUTPUT_FILE_DIR}", q(filepath.Dir(outputFile)))

	input = strings.ReplaceAll(input, "${DATE_YEAR}", time.Now().Format("2006"))
	input = strings.ReplaceAll(input, "${DATE_SHORTYEAR}", time.Now().Format("06"))
	input = strings.ReplaceAll(input, "${DATE_MONTH}", time.Now().Format("01"))
	input = strings.ReplaceAll(input, "${DATE_DAY}", time.Now().Format("02"))

	_, week := time.Now().ISOWeek()
	input = strings.ReplaceAll(input, "${DATE_WEEK}", strconv.Itoa(week))

	input = strings.ReplaceAll(input, "${TIME_HOUR}", time.Now().Format("15"))
	input = strings.ReplaceAll(input, "${TIME_MINUTE}", time.Now().Format("04"))
	input = strings.ReplaceAll(input, "${TIME_SECOND}", time.Now().Format("05"))

	input = strings.ReplaceAll(input, "${TIMESTAMP_SECONDS}", strconv.FormatInt(time.Now().Unix(), 10))
	input = strings.ReplaceAll(input, "${TIMESTAMP_MILLISECONDS}", strconv.FormatInt(time.Now().UnixMilli(), 10))
	input = strings.ReplaceAll(input, "${TIMESTAMP_MICROSECONDS}", strconv.FormatInt(time.Now().UnixMicro(), 10))
	input = strings.ReplaceAll(input, "${TIMESTAMP_NANOSECONDS}", strconv.FormatInt(time.Now().UnixNano(), 10))

	input = strings.ReplaceAll(input, "${OS_NAME}", runtime.GOOS)
	input = strings.ReplaceAll(input, "${OS_ARCH}", runtime.GOARCH)

	input = strings.ReplaceAll(input, "${SOURCE}", string(source))

	input = strings.ReplaceAll(input, "${UUID}", uuid.NewString())

	input = strings.ReplaceAll(input, "${FFMPEG}", cfg.GetString("ffmate.ffmpeg"))

	// handle metadata wildcard
	if metadata != nil {
		metadataJSON, err := json.Marshal(metadata)

		if err == nil {
			jsonStr := string(metadataJSON)
			re := regexp.MustCompile(`\$\{METADATA_([^}]+)\}`)
			input = re.ReplaceAllStringFunc(input, func(match string) string {
				path := re.FindStringSubmatch(match)[1]
				val := gjson.Get(jsonStr, path)
				if val.Exists() {
					return val.String()
				}
				return ""
			})
		}
	}

	return input
}

// quoteArg renders v as a single token for the shell-style command parsers used
// to build the ffmpeg and script command lines (github.com/mattn/go-shellwords,
// and the Windows-path-safe parser in internal/service/ffmpeg). It wraps v in
// single quotes and renders any embedded single quote as the sequence '"'"', so
// that an attacker-influenced value is always parsed as exactly one argument and
// cannot break out of the quoting to inject extra arguments (CWE-88). Single
// quoting is used because its meaning is identical in both parsers and needs no
// backslash escaping, which the Windows parser treats as a literal path
// separator rather than an escape.
func quoteArg(v string) string {
	return "'" + strings.ReplaceAll(v, "'", `'"'"'`) + "'"
}
