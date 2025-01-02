package exceptions

import "github.com/welovemedia/ffmate/sev/exceptions"

func TaskNotCancelable(err error) *exceptions.HttpError {
	return &exceptions.HttpError{HttpCode: 400, Code: "001.001.0001", Error: "task.not.cancelable", Message: err.Error()}
}
