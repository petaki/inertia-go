package inertia

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEnableSsr(t *testing.T) {
	i := New("", "", "")
	i.EnableSsr("ssr.test")

	if i.ssrURL != "ssr.test" {
		t.Errorf("expected: ssr.test, got: %v", i.ssrURL)
	}

	if i.ssrClient == nil {
		t.Error("expected: *http.Client, got: nil")
	}
}

func TestEnableSsrWithDefault(t *testing.T) {
	i := New("", "", "")
	i.EnableSsrWithDefault()

	if i.ssrURL != "http://127.0.0.1:13714" {
		t.Errorf("expected: http://127.0.0.1:13714, got: %v", i.ssrURL)
	}

	if i.ssrClient == nil {
		t.Error("expected: *http.Client, got: nil")
	}
}

func TestIsSsrEnabled(t *testing.T) {
	i := New("", "", "")

	if i.IsSsrEnabled() {
		t.Error("expected: false, got: true")
	}

	i.EnableSsrWithDefault()

	if !i.IsSsrEnabled() {
		t.Error("expected: true, got: false")
	}
}

func TestDisableSsr(t *testing.T) {
	i := New("", "", "")
	i.EnableSsrWithDefault()
	i.DisableSsr()

	if i.IsSsrEnabled() {
		t.Error("expected: false, got: true")
	}

	if i.ssrClient != nil {
		t.Error("expected: nil, got: *http.Client")
	}
}

func TestShare(t *testing.T) {
	i := New("", "", "")
	i.Share("title", "Inertia.js Go")

	title, ok := i.sharedProps["title"].(string)
	if !ok {
		t.Error("expected: title, got: empty value")
	}

	if title != "Inertia.js Go" {
		t.Errorf("expected: Inertia.js Go, got: %s", title)
	}
}

func TestShareFunc(t *testing.T) {
	i := New("", "", "")
	i.ShareFunc("asset", func(path string) (string, error) {
		return "/" + path, nil
	})

	_, ok := i.sharedFuncMap["asset"].(func(string) (string, error))
	if !ok {
		t.Error("expected: asset func, got: empty value")
	}
}

func TestShareViewData(t *testing.T) {
	i := New("", "", "")
	i.ShareViewData("env", "production")

	env, ok := i.sharedViewData["env"].(string)
	if !ok {
		t.Error("expected: env, got: empty value")
	}

	if env != "production" {
		t.Errorf("expected: production, got: %s", env)
	}
}

func TestWithProp(t *testing.T) {
	ctx := context.TODO()

	i := New("", "", "")
	ctx = i.WithProp(ctx, "user", "test-user")

	contextProps, ok := ctx.Value(ContextKeyProps).(map[string]interface{})
	if !ok {
		t.Error("expected: context props, got: empty value")
	}

	user, ok := contextProps["user"].(string)
	if !ok {
		t.Error("expected: user, got: empty value")
	}

	if user != "test-user" {
		t.Errorf("expected: test-user, got: %s", user)
	}
}

func TestWithViewData(t *testing.T) {
	ctx := context.TODO()

	i := New("", "", "")
	ctx = i.WithViewData(ctx, "meta", "test-meta")

	contextViewData, ok := ctx.Value(ContextKeyViewData).(map[string]interface{})
	if !ok {
		t.Error("expected: context view data, got: empty value")
	}

	meta, ok := contextViewData["meta"].(string)
	if !ok {
		t.Error("expected: meta, got: empty value")
	}

	if meta != "test-meta" {
		t.Errorf("expected: test-meta, got: %s", meta)
	}
}

func TestRender(t *testing.T) {
	url := "http://inertia-go.test"
	i := New(url, "", "")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderInertia, "true")
	w := httptest.NewRecorder()

	err := i.Render(w, r, "test/component", map[string]interface{}{
		"userID": 1,
	})
	if err != nil {
		t.Error(err)
	}

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code: %d, got: %d", http.StatusOK, resp.StatusCode)
	}

	if resp.Header.Get("Vary") != "Accept" {
		t.Errorf("expected: Accept, got: %s", resp.Header.Get("Vary"))
	}

	if resp.Header.Get(HeaderInertia) != "true" {
		t.Errorf("expected: true, got: %s", resp.Header.Get(HeaderInertia))
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("expected: application/json, got: %s", resp.Header.Get("Content-Type"))
	}

	var page Page

	err = json.NewDecoder(resp.Body).Decode(&page)
	if err != nil {
		t.Error(err)
	}

	if page.Component != "test/component" {
		t.Errorf("expected: test/component, got: %s", page.Component)
	}

	if page.URL != "/" {
		t.Errorf("expected: /, got: %s", page.URL)
	}

	userID, ok := page.Props["userID"].(float64)
	if !ok {
		t.Error("expected: userID, got: empty value")
	}

	if userID != 1 {
		t.Errorf("expected: 1, got: %.2f", userID)
	}
}

func TestLocation(t *testing.T) {
	url := "http://inertia-go.test"
	externalUrl := "http://dashboard.inertia-go.test"

	i := New(url, "", "")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	i.Location(w, r, externalUrl)

	resp := w.Result()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected status code: %d, got: %d", http.StatusFound, resp.StatusCode)
	}

	loc := resp.Header.Get("Location")

	if loc != externalUrl {
		t.Errorf("expected: %s, got: %s", externalUrl, loc)
	}

	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderInertia, "true")
	w = httptest.NewRecorder()

	i.Location(w, r, externalUrl)

	resp = w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusConflict {
		t.Errorf("expected status code: %d, got: %d", http.StatusConflict, resp.StatusCode)
	}

	loc = resp.Header.Get(HeaderLocation)

	if loc != externalUrl {
		t.Errorf("expected location: %s, got: %s", externalUrl, loc)
	}

	if len(body) != 0 {
		t.Errorf("expected empty body, got: %s", body)
	}
}
