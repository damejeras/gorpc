package transport

import (
	"net/http"
)

var DefaultErrorHandler ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
	clientErr, ok := err.(ClientError)
	if ok {
		_ = Encode(w, r, clientErr.Code, clientErr)
		return
	}

	_ = Encode(w, r, http.StatusInternalServerError, map[string]string{"message": "internal server error"})
}

type ErrorHandler func(w http.ResponseWriter, r *http.Request, err error)

type ClientError struct {
	Code    int    `json:"-"`
	Message string `json:"error"`
}

func (e ClientError) Error() string { return e.Message }
