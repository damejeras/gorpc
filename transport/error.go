package transport

import (
	"log"
	"net/http"
)

type ErrorHandler func(w http.ResponseWriter, r *http.Request, err error)

var defaultErrorHandler ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
	errObj := struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	}
	if err := Encode(w, r, http.StatusInternalServerError, errObj); err != nil {
		log.Printf("failed to encode error: %s\n", err)
	}
}
