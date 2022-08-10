package transport

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Server interface {
	http.Handler

	OnErr(w http.ResponseWriter, r *http.Request, err error)
	Register(service, method string, h http.HandlerFunc)
}

type Middleware func(http.Handler) http.Handler

type server struct {
	routes          map[string]http.Handler
	notFoundHandler http.Handler
	errHandler      ErrorHandler
	pathFn          func(service, method string) string
	mw              []Middleware
}

func NewServer(options ...Option) Server {
	srv := &server{
		routes:          make(map[string]http.Handler),
		notFoundHandler: http.NotFoundHandler(),
		errHandler:      DefaultErrorHandler,
		pathFn: func(service, method string) string {
			return "/" + service + "." + method
		},
		mw: make([]Middleware, 0),
	}

	for i := range options {
		options[i](srv)
	}

	return srv
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.notFoundHandler.ServeHTTP(w, r)

		return
	}

	handler, ok := s.routes[r.URL.Path]
	if !ok {
		s.notFoundHandler.ServeHTTP(w, r)

		return
	}

	handler.ServeHTTP(w, r)
}

func (s *server) OnErr(w http.ResponseWriter, r *http.Request, err error) {
	s.errHandler(w, r, err)
}

func (s *server) Register(service, method string, handler http.HandlerFunc) {
	if len(s.mw) > 0 {
		s.routes[s.pathFn(service, method)] = chainMiddleware(handler, s.mw...)
	} else {
		s.routes[s.pathFn(service, method)] = handler
	}
}

func Encode(w http.ResponseWriter, r *http.Request, status int, payload interface{}) error {
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	var out io.Writer = w
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		gzw := gzip.NewWriter(w)
		out = gzw
		defer gzw.Close()
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if _, err := out.Write(bodyBytes); err != nil {
		return fmt.Errorf("write body: %w", err)
	}

	return nil
}

func Decode(r *http.Request, v interface{}) error {
	if r.Header.Get("Content-type") != "application/json" {
		return ClientError{
			Code:    http.StatusUnsupportedMediaType,
			Message: "application/json content-type expected",
		}
	}

	if err := json.NewDecoder(io.LimitReader(r.Body, 1024*1024)).Decode(v); err != nil {
		return fmt.Errorf("decode request body: %w", err)
	}

	return r.Body.Close()
}

func chainMiddleware(handle http.Handler, mw ...Middleware) http.Handler {
	for i := range mw {
		handle = mw[len(mw)-1-i](handle)
	}

	return handle
}
