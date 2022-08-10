package transport

import (
	"strings"
)

type Option func(*server)

func WithErrorHandler(handler ErrorHandler) Option {
	return func(s *server) {
		s.errHandler = handler
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

func WithMiddleware(mw Middleware) Option {
	return func(s *server) {
		s.mw = append(s.mw, mw)
	}
}
