package inertia

import (
	"context"
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
