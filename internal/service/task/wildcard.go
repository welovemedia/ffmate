package task

import (
	"encoding/json"
	"fmt"
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

func (s *Service) wildcardReplacer(input string, inputFile string, outputFile string, source dto.TaskSource, metadata *dto.MetadataMap) string {
	input = strings.ReplaceAll(input, "${INPUT_FILE}", fmt.Sprintf("\"%s\"", inputFile))
	input = strings.ReplaceAll(input, "${OUTPUT_FILE}", fmt.Sprintf("\"%s\"", outputFile))

	input = strings.ReplaceAll(input, "${INPUT_FILE_BASE}", filepath.Base(inputFile))
	input = strings.ReplaceAll(input, "${OUTPUT_FILE_BASE}", filepath.Base(inputFile))
	input = strings.ReplaceAll(input, "${INPUT_FILE_EXTENSION}", filepath.Ext(filepath.Base(inputFile)))
	input = strings.ReplaceAll(input, "${OUTPUT_FILE_EXTENSION", filepath.Ext(filepath.Base(outputFile)))
	input = strings.ReplaceAll(input, "${INPUT_FILE_BASENAME}", strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(filepath.Base(inputFile))))
	input = strings.ReplaceAll(input, "${OUTPUT_FILE_BASENAME}", strings.TrimSuffix(filepath.Base(outputFile), filepath.Ext(filepath.Base(outputFile))))
	input = strings.ReplaceAll(input, "${INPUT_FILE_DIR}", filepath.Dir(inputFile))
	input = strings.ReplaceAll(input, "${OUTPUT_FILE_DIR}", filepath.Dir(inputFile))

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
