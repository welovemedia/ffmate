package exception

type HttpError struct {
	HttpCode int    `json:"-"`
	Code     string `json:"code"`
	Error    string `json:"err"`
	Message  string `json:"message"`
	Docs     string `json:"docs,omitempty"`
}

func InternalServerError(err error) *HttpError {
	return &HttpError{HttpCode: 500, Code: "001.000.0000", Error: "internal.server.error", Message: err.Error()}
}

func HttpInvalidRequest() *HttpError {
	return &HttpError{HttpCode: 400, Code: "001.000.0002", Error: "invalid.request", Message: "invalid request"}
}

func HttpBadRequest(err error, docs string) *HttpError {
	return &HttpError{HttpCode: 400, Code: "001.000.0003", Error: "bad.request", Message: err.Error(), Docs: docs}
}

func HttpInvalidBody(err error) *HttpError {
	return &HttpError{HttpCode: 400, Code: "001.000.0005", Error: "invalid.request.body", Message: err.Error()}
}

func HttpInvalidParam(name string) *HttpError {
	return &HttpError{HttpCode: 400, Code: "001.000.0006", Error: "invalid.param", Message: "invalid parameter '" + name + "'"}
}

func HttpInvalidQuery(name string) *HttpError {
	return &HttpError{HttpCode: 400, Code: "001.000.0007", Error: "invalid.query", Message: "invalid query '" + name + "'"}
}

func HttpNotFound(err error, docs string) *HttpError {
	return &HttpError{HttpCode: 400, Code: "001.000.0008", Error: "not.found", Message: err.Error(), Docs: docs}
}
