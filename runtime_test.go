package inertia

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRuntime(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	props := map[string]any{"title": "Test"}

	rt := newRuntime(r, "test/component", props)

	if rt.isPartial {
		t.Error("expected isPartial to be false")
	}

	if rt.props["title"] != "Test" {
		t.Errorf("expected: Test, got: %v", rt.props["title"])
	}

	if len(rt.only) != 0 {
		t.Errorf("expected empty only, got: %d", len(rt.only))
	}

	if len(rt.except) != 0 {
		t.Errorf("expected empty except, got: %d", len(rt.except))
	}

	if len(rt.exceptOnce) != 0 {
		t.Errorf("expected empty exceptOnce, got: %d", len(rt.exceptOnce))
	}

	if len(rt.reset) != 0 {
		t.Errorf("expected empty reset, got: %d", len(rt.reset))
	}
}

func TestNewRuntimeWithPartialOnly(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderPartialComponent, "test/component")
	r.Header.Set(HeaderPartialOnly, "title,user")

	rt := newRuntime(r, "test/component", nil)

	if !rt.isPartial {
		t.Error("expected isPartial to be true")
	}

	if len(rt.only) != 2 {
		t.Errorf("expected 2 only entries, got: %d", len(rt.only))
	}

	_, ok := rt.only["title"]
	if !ok {
		t.Error("expected title in only")
	}

	_, ok = rt.only["user"]
	if !ok {
		t.Error("expected user in only")
	}
}

func TestNewRuntimeWithPartialExcept(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderPartialComponent, "test/component")
	r.Header.Set(HeaderPartialExcept, "secret")

	rt := newRuntime(r, "test/component", nil)

	if !rt.isPartial {
		t.Error("expected isPartial to be true")
	}

	if len(rt.except) != 1 {
		t.Errorf("expected 1 except entry, got: %d", len(rt.except))
	}

	_, ok := rt.except["secret"]
	if !ok {
		t.Error("expected secret in except")
	}
}

func TestNewRuntimeWithExceptOnce(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderExceptOnceProps, "plans,flags")

	rt := newRuntime(r, "test/component", nil)

	if rt.isPartial {
		t.Error("expected isPartial to be false")
	}

	if len(rt.exceptOnce) != 2 {
		t.Errorf("expected 2 exceptOnce entries, got: %d", len(rt.exceptOnce))
	}

	_, ok := rt.exceptOnce["plans"]
	if !ok {
		t.Error("expected plans in exceptOnce")
	}

	_, ok = rt.exceptOnce["flags"]
	if !ok {
		t.Error("expected flags in exceptOnce")
	}
}

func TestNewRuntimeWithReset(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderReset, "results,items")

	rt := newRuntime(r, "test/component", nil)

	if rt.isPartial {
		t.Error("expected isPartial to be false")
	}

	if len(rt.reset) != 2 {
		t.Errorf("expected 2 reset entries, got: %d", len(rt.reset))
	}

	_, ok := rt.reset["results"]
	if !ok {
		t.Error("expected results in reset")
	}

	_, ok = rt.reset["items"]
	if !ok {
		t.Error("expected items in reset")
	}
}

func TestNewRuntimeWithDifferentComponent(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderPartialComponent, "other/component")
	r.Header.Set(HeaderPartialOnly, "title")

	rt := newRuntime(r, "test/component", nil)

	if rt.isPartial {
		t.Error("expected isPartial to be false when component does not match")
	}

	if len(rt.only) != 0 {
		t.Errorf("expected empty only, got: %d", len(rt.only))
	}
}
