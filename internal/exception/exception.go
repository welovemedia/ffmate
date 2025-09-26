package exception

type HTTPError struct {
	Code     string `json:"code"`
	Error    string `json:"err"`
	Message  string `json:"message"`
	Docs     string `json:"docs,omitempty"`
	HTTPCode int    `json:"-"`
}

func InternalServerError(err error) *HTTPError {
	return &HTTPError{HTTPCode: 500, Code: "001.000.0000", Error: "internal.server.error", Message: err.Error()}
}

func HTTPInvalidRequest() *HTTPError {
	return &HTTPError{HTTPCode: 400, Code: "001.000.0002", Error: "invalid.request", Message: "invalid request"}
}

func HTTPBadRequest(err error, docs string) *HTTPError {
	return &HTTPError{HTTPCode: 400, Code: "001.000.0003", Error: "bad.request", Message: err.Error(), Docs: docs}
}

func HTTPInvalidBody(err error) *HTTPError {
	return &HTTPError{HTTPCode: 400, Code: "001.000.0005", Error: "invalid.request.body", Message: err.Error()}
}

func HTTPInvalidParam(name string) *HTTPError {
	return &HTTPError{HTTPCode: 400, Code: "001.000.0006", Error: "invalid.param", Message: "invalid parameter '" + name + "'"}
}

func HTTPInvalidQuery(name string) *HTTPError {
	return &HTTPError{HTTPCode: 400, Code: "001.000.0007", Error: "invalid.query", Message: "invalid query '" + name + "'"}
}

func HTTPNotFound(err error, docs string) *HTTPError {
	return &HTTPError{HTTPCode: 400, Code: "001.000.0008", Error: "not.found", Message: err.Error(), Docs: docs}
}
