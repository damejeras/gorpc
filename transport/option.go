package transport

import (
	"net/http"
	"strings"
)

type Option func(*server)

func WithNotFoundHandler(notFoundHandler http.Handler) Option {
	return func(s *server) {
		s.notFoundHandler = notFoundHandler
	}
}

func WithErrorHandler(errHandler ErrorHandler) Option {
	return func(s *server) {
		s.errHandler = errHandler
	}
}

func WithPathPrefix(prefix string) Option {
	trimmedPrefix := strings.Trim(prefix, "/")

	return func(s *server) {
		s.pathFn = func(service, method string) string {
			return "/" + trimmedPrefix + "/" + service + "." + method
		}
	}
}
