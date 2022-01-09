package otohttp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServer(t *testing.T) {
	srv := NewServer()
	srv.Register("Service", "Method", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/Service.Method", nil)
	srv.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("expected %d status code, got %d", http.StatusOK, w.Code)
	}

	expected := `{"status":"ok"}`
	if w.Body.String() != expected {
		t.Errorf("expected %q response body, got %q", expected, w.Body.String())
	}
}

func TestWithPathPrefix(t *testing.T) {
	srv := NewServer(WithPathPrefix("/gorpc/"))
	srv.Register("Service", "Method", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/gorpc/Service.Method", nil)
	srv.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("expected %d status code, got %d", http.StatusOK, w.Code)
	}

	expected := `{"status":"ok"}`
	if w.Body.String() != expected {
		t.Errorf("expected %q response body, got %q", expected, w.Body.String())
	}
}

func TestEncode(t *testing.T) {
	data := struct {
		Greeting string `json:"greeting"`
	}{
		Greeting: "Hi there",
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/gorpc/Service.Method", nil)
	err := Encode(w, r, http.StatusOK, data)
	if err != nil {
		t.Error(err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected %d status code, got %d", http.StatusOK, w.Code)
	}

	expected := `{"greeting":"Hi there"}`
	if w.Body.String() != expected {
		t.Errorf("expected %q response body, got %q", expected, w.Body.String())
	}

	expected = "application/json; charset=utf-8"
	if w.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("expected content type to be %q, got %q", expected, w.Header().Get("Content-Type"))
	}
}

func TestDecode(t *testing.T) {
	//is := is.New(t)
	type r struct {
		Name string
	}
	j := `[
		{"name": "Mat"},
		{"name": "David"},
		{"name": "Aaron"}
	]`
	req, err := http.NewRequest(http.MethodPost, "/service/method", strings.NewReader(j))
	//is.NoErr(err)
	if err != nil {
		t.Error(err)
	}

	req.Header.Set("Content-Type", "application/json")
	var requestObjects []r
	err = Decode(req, &requestObjects)
	if err != nil {
		t.Error(err)
	}

	if len(requestObjects) != 3 {
		t.Errorf("expected %d request objects, got %d", 3, len(requestObjects))
	}

	if requestObjects[0].Name != "Mat" {
		t.Errorf("first request object's name had to be Mat")
	}

	if requestObjects[1].Name != "David" {
		t.Errorf("second request object's name had to be David")
	}

	if requestObjects[2].Name != "Aaron" {
		t.Errorf("first request object's name had to be Aaron")
	}
}
