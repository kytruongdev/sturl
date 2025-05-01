package httpserver

import (
	"encoding/json"
	"io"
	"net/http"
)

var (
	webErrInvalidRequestBody = &Error{Status: http.StatusBadRequest, Code: "invalid_request_body", Desc: "invalid request body"}
)

// ParseJSON parses request body to json
func ParseJSON(r io.ReadCloser, result interface{}) *Error {
	reqBytes, err := io.ReadAll(r)
	defer r.Close()
	if err != nil {
		return webErrInvalidRequestBody
	}

	if err = json.Unmarshal(reqBytes, &result); err != nil {
		return webErrInvalidRequestBody
	}

	return nil
}
