package testsuite

import (
	"encoding/json"
	"io"

	"goyave.dev/goyave/v5/util/typeutil"
)

func ParseJSONBody[T any](b io.ReadCloser) (T, error) {
	var zero T

	bodyBytes, err := io.ReadAll(b)
	if err != nil {
		return zero, err
	}
	defer b.Close() // nolint:errcheck

	if len(bodyBytes) == 0 {
		return zero, nil
	}

	var generic any
	if err := json.Unmarshal(bodyBytes, &generic); err != nil {
		return zero, err
	}

	result := typeutil.MustConvert[T](generic)
	return result, nil
}

func ParseBody(b io.ReadCloser) ([]byte, error) {
	bodyBytes, err := io.ReadAll(b)
	if err != nil {
		return nil, err
	}
	defer b.Close() // nolint:errcheck

	return bodyBytes, nil
}
