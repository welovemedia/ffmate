package ffmpeg

import (
	"strconv"
	"strings"
	"time"
)

func evaluateWildcards(request *ExecutionRequest) {
	request.Command = strings.ReplaceAll(request.Command, "${INPUT_FILE}", request.InputFile)
	request.Command = strings.ReplaceAll(request.Command, "${OUTPUT_FILE}", request.OutputFile)

	request.Command = strings.ReplaceAll(request.Command, "${DATE_YEAR}", time.Now().Format("2006"))
	request.Command = strings.ReplaceAll(request.Command, "${DATE_SHORTYEAR}", time.Now().Format("06"))
	request.Command = strings.ReplaceAll(request.Command, "${DATE_MONTH}", time.Now().Format("01"))
	request.Command = strings.ReplaceAll(request.Command, "${DATE_DAY}", time.Now().Format("02"))

	request.Command = strings.ReplaceAll(request.Command, "${TIME_HOUR}", time.Now().Format("15"))
	request.Command = strings.ReplaceAll(request.Command, "${TIME_MINUTE}", time.Now().Format("04"))
	request.Command = strings.ReplaceAll(request.Command, "${TIME_SECOND}", time.Now().Format("05"))

	request.Command = strings.ReplaceAll(request.Command, "${TIMESTAMP_SECONDS}", strconv.FormatInt(time.Now().Unix(), 10))
	request.Command = strings.ReplaceAll(request.Command, "${TIMESTAMP_MILLISECONDS}", strconv.FormatInt(time.Now().UnixMilli(), 10))
	request.Command = strings.ReplaceAll(request.Command, "${TIMESTAMP_MICROSECONDS}", strconv.FormatInt(time.Now().UnixMicro(), 10))
	request.Command = strings.ReplaceAll(request.Command, "${TIMESTAMP_NANOSECONDS}", strconv.FormatInt(time.Now().UnixNano(), 10))
}
