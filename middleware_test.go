package inertia

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewareWithNormalRequest(t *testing.T) {
	url := "http://inertia-go.test"

	i := New(url, "", "abc123")
	ih := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "handler")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	m := i.Middleware(http.HandlerFunc(ih))
	m.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code: %d, got: %d", http.StatusOK, resp.StatusCode)
	}

	if string(body) != "handler" {
		t.Errorf("expected body: handler, got: %s", body)
	}

	req = httptest.NewRequest(http.MethodPost, "/", nil)
	w = httptest.NewRecorder()
	m.ServeHTTP(w, req)

	resp = w.Result()
	body, _ = io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code: %d, got: %d", http.StatusOK, resp.StatusCode)
	}

	if string(body) != "handler" {
		t.Errorf("expected body: handler, got: %s", body)
	}
}

func TestMiddlewareWithInertiaRequest(t *testing.T) {
	url := "http://inertia-go.test"

	i := New(url, "", "abc123")
	ih := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "handler")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(HeaderInertia, "true")
	req.Header.Set(HeaderVersion, "abc123")
	w := httptest.NewRecorder()

	m := i.Middleware(http.HandlerFunc(ih))
	m.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code: %d, got: %d", http.StatusOK, resp.StatusCode)
	}

	if string(body) != "handler" {
		t.Errorf("expected body: handler, got: %s", body)
	}

	req = httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(HeaderInertia, "true")
	req.Header.Set(HeaderVersion, "abc")

	w = httptest.NewRecorder()
	m.ServeHTTP(w, req)

	resp = w.Result()
	body, _ = io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code: %d, got: %d", http.StatusOK, resp.StatusCode)
	}

	if string(body) != "handler" {
		t.Errorf("expected body: handler, got: %s", body)
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(HeaderInertia, "true")
	req.Header.Set(HeaderVersion, "abc")

	w = httptest.NewRecorder()
	m.ServeHTTP(w, req)

	resp = w.Result()
	body, _ = io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusConflict {
		t.Errorf("expected status code: %d, got: %d", http.StatusConflict, resp.StatusCode)
	}

	loc := resp.Header.Get(HeaderLocation)

	if loc != url+"/" {
		t.Errorf("expected location: %s, got: %s", url+"/", loc)
	}

	if len(body) != 0 {
		t.Errorf("expected empty body, got: %s", body)
	}
}
