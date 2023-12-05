package web

import (
	"fmt"
	"net/http"
	"strings"
)

type HttpError struct {
	Code    int
	Message string
}

func (e HttpError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

func Error(code int, msg ...string) HttpError {
	var message = http.StatusText(code)
	if len(msg) > 0 {
		message = strings.Join(msg, ",")
	}
	return HttpError{Code: code, Message: message}
}
