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

func TestEnableSsrWithClient(t *testing.T) {
	i := New("", "", "")
	client := &http.Client{}
	i.EnableSsr("ssr.test", client)

	if i.ssrURL != "ssr.test" {
		t.Errorf("expected: ssr.test, got: %v", i.ssrURL)
	}

	if i.ssrClient != client {
		t.Error("expected: custom *http.Client, got: different client")
	}
}

func TestEnableSsrConcurrent(t *testing.T) {
	i := New("http://inertia-go.test", "", "")

	done := make(chan struct{})

	go func() {
		defer close(done)

		for range 100 {
			i.EnableSsr("http://127.0.0.1:13714/render")
		}
	}()

	for range 100 {
		i.IsSsrEnabled()
	}

	<-done
}

func TestEnableSsrWithDefault(t *testing.T) {
	i := New("", "", "")
	i.EnableSsrWithDefault()

	if i.ssrURL != "http://127.0.0.1:13714/render" {
		t.Errorf("expected: http://127.0.0.1:13714/render, got: %v", i.ssrURL)
	}

	if i.ssrClient == nil {
		t.Error("expected: *http.Client, got: nil")
	}
}

func TestEnableSsrWithDefaultWithClient(t *testing.T) {
	i := New("", "", "")
	client := &http.Client{}
	i.EnableSsrWithDefault(client)

	if i.ssrURL != "http://127.0.0.1:13714/render" {
		t.Errorf("expected: http://127.0.0.1:13714/render, got: %v", i.ssrURL)
	}

	if i.ssrClient != client {
		t.Error("expected: custom *http.Client, got: different client")
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

func TestDisableSsrConcurrent(t *testing.T) {
	i := New("http://inertia-go.test", "", "")
	i.EnableSsrWithDefault()

	done := make(chan struct{})

	go func() {
		defer close(done)

		for range 100 {
			i.DisableSsr()
		}
	}()

	for range 100 {
		i.IsSsrEnabled()
	}

	<-done
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

func TestShareFuncConcurrent(t *testing.T) {
	i := New("http://inertia-go.test", "", "")

	done := make(chan struct{})

	go func() {
		defer close(done)

		for n := range 100 {
			i.ShareFunc("key", n)
		}
	}()

	for n := range 100 {
		i.ShareFunc("key", n)
	}

	<-done
}

func TestShareFuncAndRenderConcurrent(t *testing.T) {
	i := New("http://inertia-go.test", "", "")

	done := make(chan struct{})

	go func() {
		defer close(done)

		for n := range 100 {
			i.ShareFunc("key", n)
		}
	}()

	for range 100 {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set(HeaderInertia, "true")
		w := httptest.NewRecorder()

		i.Render(w, r, "test/component", nil)
	}

	<-done
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

func TestShareViewDataConcurrent(t *testing.T) {
	i := New("http://inertia-go.test", "", "")

	done := make(chan struct{})

	go func() {
		defer close(done)

		for n := range 100 {
			i.ShareViewData("key", n)
		}
	}()

	for n := range 100 {
		i.ShareViewData("key", n)
	}

	<-done
}

func TestShareViewDataAndRenderConcurrent(t *testing.T) {
	i := New("http://inertia-go.test", "", "")

	done := make(chan struct{})

	go func() {
		defer close(done)

		for n := range 100 {
			i.ShareViewData("key", n)
		}
	}()

	for range 100 {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set(HeaderInertia, "true")
		w := httptest.NewRecorder()

		i.Render(w, r, "test/component", nil)
	}

	<-done
}

func TestWithViewData(t *testing.T) {
	ctx := context.TODO()

	i := New("", "", "")
	ctx = i.WithViewData(ctx, "meta", "test-meta")

	contextViewData, ok := ctx.Value(contextKeyViewData).(map[string]any)
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

func TestShareConcurrent(t *testing.T) {
	i := New("http://inertia-go.test", "", "")

	done := make(chan struct{})

	go func() {
		defer close(done)

		for n := range 100 {
			i.Share("key", n)
		}
	}()

	for n := range 100 {
		i.Share("key", n)
	}

	<-done
}

func TestShareAndRenderConcurrent(t *testing.T) {
	i := New("http://inertia-go.test", "", "")

	done := make(chan struct{})

	go func() {
		defer close(done)

		for n := range 100 {
			i.Share("key", n)
		}
	}()

	for range 100 {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set(HeaderInertia, "true")
		w := httptest.NewRecorder()

		i.Render(w, r, "test/component", nil)
	}

	<-done
}

func TestWithProp(t *testing.T) {
	ctx := context.TODO()

	i := New("", "", "")
	ctx = i.WithProp(ctx, "user", "test-user")

	contextProps, ok := ctx.Value(contextKeyProps).(map[string]any)
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

func TestRender(t *testing.T) {
	url := "http://inertia-go.test"
	i := New(url, "", "")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderInertia, "true")
	w := httptest.NewRecorder()

	err := i.Render(w, r, "test/component", map[string]any{
		"userID": 1,
	})
	if err != nil {
		t.Error(err)
	}

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code: %d, got: %d", http.StatusOK, resp.StatusCode)
	}

	varyValues := resp.Header.Values("Vary")
	if len(varyValues) != 1 {
		t.Errorf("expected 1 Vary value, got: %d", len(varyValues))
	}

	if varyValues[0] != HeaderInertia {
		t.Errorf("expected: %s, got: %s", HeaderInertia, varyValues[0])
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

func TestRenderWithOptionalProp(t *testing.T) {
	i := New("http://inertia-go.test", "", "")

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderInertia, "true")
	ctx := i.WithOptionalProp(r.Context(), "extra", func() any { return "opt" })
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	err := i.Render(w, r, "test/component", map[string]any{
		"title": "Test",
	})
	if err != nil {
		t.Error(err)
	}

	var page Page

	err = json.NewDecoder(w.Result().Body).Decode(&page)
	if err != nil {
		t.Error(err)
	}

	if _, ok := page.Props["extra"]; ok {
		t.Error("expected optional prop to be excluded from full load")
	}

	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderInertia, "true")
	r.Header.Set(HeaderPartialComponent, "test/component")
	r.Header.Set(HeaderPartialOnly, "extra")
	ctx = i.WithOptionalProp(r.Context(), "extra", func() any { return "opt" })
	r = r.WithContext(ctx)
	w = httptest.NewRecorder()

	err = i.Render(w, r, "test/component", map[string]any{
		"title": "Test",
	})
	if err != nil {
		t.Error(err)
	}

	var page2 Page

	err = json.NewDecoder(w.Result().Body).Decode(&page2)
	if err != nil {
		t.Error(err)
	}

	if page2.Props["extra"] != "opt" {
		t.Errorf("expected: opt, got: %v", page2.Props["extra"])
	}
}

func TestRenderWithAlwaysProp(t *testing.T) {
	i := New("http://inertia-go.test", "", "")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderInertia, "true")
	r.Header.Set(HeaderPartialComponent, "test/component")
	r.Header.Set(HeaderPartialOnly, "title")
	ctx := i.WithAlwaysProp(r.Context(), "errors", func() any { return map[string]string{} })
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	err := i.Render(w, r, "test/component", map[string]any{
		"title": "Test",
	})
	if err != nil {
		t.Error(err)
	}

	var page Page

	err = json.NewDecoder(w.Result().Body).Decode(&page)
	if err != nil {
		t.Error(err)
	}

	if _, ok := page.Props["errors"]; !ok {
		t.Error("expected always prop to be included even when not requested")
	}
}

func TestRenderWithDeferredProp(t *testing.T) {
	i := New("http://inertia-go.test", "", "")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderInertia, "true")
	ctx := i.WithDeferredProp(r.Context(), "comments", func() any { return []string{"a", "b"} })
	ctx = i.WithDeferredProp(ctx, "sidebar", func() any { return "side" }, "sidebar")
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	err := i.Render(w, r, "test/component", map[string]any{
		"title": "Test",
	})
	if err != nil {
		t.Error(err)
	}

	var page Page

	err = json.NewDecoder(w.Result().Body).Decode(&page)
	if err != nil {
		t.Error(err)
	}

	if _, ok := page.Props["comments"]; ok {
		t.Error("expected deferred prop comments to be excluded from props")
	}

	if _, ok := page.Props["sidebar"]; ok {
		t.Error("expected deferred prop sidebar to be excluded from props")
	}

	if page.Props["title"] != "Test" {
		t.Errorf("expected: Test, got: %s", page.Props["title"])
	}

	if len(page.DeferredProps["default"]) != 1 || page.DeferredProps["default"][0] != "comments" {
		t.Errorf("expected deferred default group [comments], got: %v", page.DeferredProps["default"])
	}

	if len(page.DeferredProps["sidebar"]) != 1 || page.DeferredProps["sidebar"][0] != "sidebar" {
		t.Errorf("expected deferred sidebar group [sidebar], got: %v", page.DeferredProps["sidebar"])
	}
}

func TestRenderWithDeferredPropPartialReload(t *testing.T) {
	i := New("http://inertia-go.test", "", "")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderInertia, "true")
	r.Header.Set(HeaderPartialComponent, "test/component")
	r.Header.Set(HeaderPartialOnly, "comments")
	ctx := i.WithDeferredProp(r.Context(), "comments", func() any { return []string{"a", "b"} })
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	err := i.Render(w, r, "test/component", map[string]any{
		"title": "Test",
	})
	if err != nil {
		t.Error(err)
	}

	var page Page

	err = json.NewDecoder(w.Result().Body).Decode(&page)
	if err != nil {
		t.Error(err)
	}

	comments, ok := page.Props["comments"]
	if !ok {
		t.Error("expected comments to be resolved on partial reload")
	}

	items, ok := comments.([]any)
	if !ok || len(items) != 2 {
		t.Errorf("expected [a b], got: %v", comments)
	}

	if _, ok := page.Props["title"]; ok {
		t.Error("expected title to be excluded from partial reload")
	}
}

func TestRenderWithMergeProp(t *testing.T) {
	i := New("http://inertia-go.test", "", "")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderInertia, "true")
	ctx := i.WithMergeProp(r.Context(), "results", func() any { return []int{1, 2} })
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	err := i.Render(w, r, "test/component", nil)
	if err != nil {
		t.Error(err)
	}

	var page Page

	err = json.NewDecoder(w.Result().Body).Decode(&page)
	if err != nil {
		t.Error(err)
	}

	if _, ok := page.Props["results"]; !ok {
		t.Error("expected results prop to be resolved")
	}

	if len(page.MergeProps) != 1 || page.MergeProps[0] != "results" {
		t.Errorf("expected mergeProps [results], got: %v", page.MergeProps)
	}
}

func TestRenderWithMergePropMatchOn(t *testing.T) {
	i := New("http://inertia-go.test", "", "")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderInertia, "true")
	ctx := i.WithMergeProp(r.Context(), "results", func() any { return []int{1, 2} }, "id")
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	err := i.Render(w, r, "test/component", nil)
	if err != nil {
		t.Error(err)
	}

	var page Page

	err = json.NewDecoder(w.Result().Body).Decode(&page)
	if err != nil {
		t.Error(err)
	}

	if len(page.MergeProps) != 1 || page.MergeProps[0] != "results" {
		t.Errorf("expected mergeProps [results], got: %v", page.MergeProps)
	}

	if len(page.MatchPropsOn) != 1 || page.MatchPropsOn[0] != "results.id" {
		t.Errorf("expected matchPropsOn [results.id], got: %v", page.MatchPropsOn)
	}
}

func TestRenderWithResetMergeProp(t *testing.T) {
	i := New("http://inertia-go.test", "", "")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderInertia, "true")
	r.Header.Set(HeaderReset, "results")
	ctx := i.WithMergeProp(r.Context(), "results", func() any { return []int{1, 2} }, "id")
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	err := i.Render(w, r, "test/component", nil)
	if err != nil {
		t.Error(err)
	}

	var page Page

	err = json.NewDecoder(w.Result().Body).Decode(&page)
	if err != nil {
		t.Error(err)
	}

	if _, ok := page.Props["results"]; !ok {
		t.Error("expected results prop to be included")
	}

	if len(page.MergeProps) != 0 {
		t.Errorf("expected mergeProps to be empty, got: %v", page.MergeProps)
	}

	if len(page.MatchPropsOn) != 0 {
		t.Errorf("expected matchPropsOn to be empty, got: %v", page.MatchPropsOn)
	}
}

func TestRenderWithDeepMergeProp(t *testing.T) {
	i := New("http://inertia-go.test", "", "")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderInertia, "true")
	ctx := i.WithDeepMergeProp(r.Context(), "settings", func() any { return map[string]string{"theme": "dark"} })
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	err := i.Render(w, r, "test/component", nil)
	if err != nil {
		t.Error(err)
	}

	var page Page

	err = json.NewDecoder(w.Result().Body).Decode(&page)
	if err != nil {
		t.Error(err)
	}

	if _, ok := page.Props["settings"]; !ok {
		t.Error("expected settings prop to be resolved")
	}

	if len(page.DeepMergeProps) != 1 || page.DeepMergeProps[0] != "settings" {
		t.Errorf("expected deepMergeProps [settings], got: %v", page.DeepMergeProps)
	}
}

func TestRenderWithDeepMergePropMatchOn(t *testing.T) {
	i := New("http://inertia-go.test", "", "")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderInertia, "true")
	ctx := i.WithDeepMergeProp(r.Context(), "settings", func() any { return map[string]string{"theme": "dark"} }, "key")
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	err := i.Render(w, r, "test/component", nil)
	if err != nil {
		t.Error(err)
	}

	var page Page

	err = json.NewDecoder(w.Result().Body).Decode(&page)
	if err != nil {
		t.Error(err)
	}

	if len(page.DeepMergeProps) != 1 || page.DeepMergeProps[0] != "settings" {
		t.Errorf("expected deepMergeProps [settings], got: %v", page.DeepMergeProps)
	}

	if len(page.MatchPropsOn) != 1 || page.MatchPropsOn[0] != "settings.key" {
		t.Errorf("expected matchPropsOn [settings.key], got: %v", page.MatchPropsOn)
	}
}

func TestRenderWithPrependProp(t *testing.T) {
	i := New("http://inertia-go.test", "", "")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderInertia, "true")
	ctx := i.WithPrependProp(r.Context(), "notifications", func() any { return []string{"new"} })
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	err := i.Render(w, r, "test/component", nil)
	if err != nil {
		t.Error(err)
	}

	var page Page

	err = json.NewDecoder(w.Result().Body).Decode(&page)
	if err != nil {
		t.Error(err)
	}

	if _, ok := page.Props["notifications"]; !ok {
		t.Error("expected notifications prop to be resolved")
	}

	if len(page.PrependProps) != 1 || page.PrependProps[0] != "notifications" {
		t.Errorf("expected prependProps [notifications], got: %v", page.PrependProps)
	}
}

func TestRenderWithPrependPropMatchOn(t *testing.T) {
	i := New("http://inertia-go.test", "", "")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderInertia, "true")
	ctx := i.WithPrependProp(r.Context(), "notifications", func() any { return []string{"new"} }, "uuid")
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	err := i.Render(w, r, "test/component", nil)
	if err != nil {
		t.Error(err)
	}

	var page Page

	err = json.NewDecoder(w.Result().Body).Decode(&page)
	if err != nil {
		t.Error(err)
	}

	if len(page.PrependProps) != 1 || page.PrependProps[0] != "notifications" {
		t.Errorf("expected prependProps [notifications], got: %v", page.PrependProps)
	}

	if len(page.MatchPropsOn) != 1 || page.MatchPropsOn[0] != "notifications.uuid" {
		t.Errorf("expected matchPropsOn [notifications.uuid], got: %v", page.MatchPropsOn)
	}
}

func TestRenderWithScrollProp(t *testing.T) {
	i := New("http://inertia-go.test", "", "")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderInertia, "true")
	ctx := i.WithScrollProp(r.Context(), "items", ScrollPageProp{
		PageName:    "page",
		CurrentPage: 1,
		NextPage:    2,
	})
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	err := i.Render(w, r, "test/component", nil)
	if err != nil {
		t.Error(err)
	}

	var page Page

	err = json.NewDecoder(w.Result().Body).Decode(&page)
	if err != nil {
		t.Error(err)
	}

	if page.ScrollProps == nil {
		t.Error("expected scrollProps to be set")
	}

	scroll, ok := page.ScrollProps["items"]
	if !ok {
		t.Error("expected scrollProps[items] to exist")
	}

	if scroll.PageName != "page" {
		t.Errorf("expected pageName: page, got: %s", scroll.PageName)
	}

	if scroll.CurrentPage != float64(1) {
		t.Errorf("expected currentPage: 1, got: %v", scroll.CurrentPage)
	}

	if scroll.NextPage != float64(2) {
		t.Errorf("expected nextPage: 2, got: %v", scroll.NextPage)
	}
}

func TestRenderWithResetScrollProp(t *testing.T) {
	i := New("http://inertia-go.test", "", "")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderInertia, "true")
	r.Header.Set(HeaderReset, "items")
	ctx := i.WithScrollProp(r.Context(), "items", ScrollPageProp{
		PageName:    "page",
		CurrentPage: 1,
	})
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	err := i.Render(w, r, "test/component", nil)
	if err != nil {
		t.Error(err)
	}

	var page Page

	err = json.NewDecoder(w.Result().Body).Decode(&page)
	if err != nil {
		t.Error(err)
	}

	scroll, ok := page.ScrollProps["items"]
	if !ok {
		t.Error("expected scrollProps[items] to exist")
	}

	if !scroll.Reset {
		t.Error("expected scrollProps[items].reset to be true")
	}
}

func TestRenderWithOnceProp(t *testing.T) {
	i := New("http://inertia-go.test", "", "")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderInertia, "true")
	r.Header.Set(HeaderExceptOnceProps, "plans")
	ctx := i.WithOnceProp(r.Context(), "plans", func() any { return []string{"free", "pro"} })
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	err := i.Render(w, r, "test/component", map[string]any{
		"title": "Test",
	})
	if err != nil {
		t.Error(err)
	}

	var page Page

	err = json.NewDecoder(w.Result().Body).Decode(&page)
	if err != nil {
		t.Error(err)
	}

	if _, ok := page.Props["plans"]; ok {
		t.Error("expected once prop to be excluded when listed in except-once header")
	}

	if page.Props["title"] != "Test" {
		t.Errorf("expected: Test, got: %v", page.Props["title"])
	}

	if page.OnceProps == nil {
		t.Error("expected onceProps to be set")
	}

	if prop, ok := page.OnceProps["plans"]; !ok || prop.Prop != "plans" {
		t.Errorf("expected onceProps[plans].prop = plans, got: %v", page.OnceProps)
	}
}

func TestRenderWithOnceModifier(t *testing.T) {
	i := New("http://inertia-go.test", "", "")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderInertia, "true")
	ctx := i.WithMergeProp(r.Context(), "activity", func() any { return []string{"a"} })
	ctx = i.WithOnce(ctx, "activity")
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	err := i.Render(w, r, "test/component", nil)
	if err != nil {
		t.Error(err)
	}

	var page Page

	err = json.NewDecoder(w.Result().Body).Decode(&page)
	if err != nil {
		t.Error(err)
	}

	if _, ok := page.Props["activity"]; !ok {
		t.Error("expected activity prop to be resolved")
	}

	if len(page.MergeProps) != 1 || page.MergeProps[0] != "activity" {
		t.Errorf("expected mergeProps [activity], got: %v", page.MergeProps)
	}

	if prop, ok := page.OnceProps["activity"]; !ok || prop.Prop != "activity" {
		t.Errorf("expected onceProps[activity].prop = activity, got: %v", page.OnceProps)
	}
}

func TestRenderWithOnceModifierExceptOnce(t *testing.T) {
	i := New("http://inertia-go.test", "", "")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderInertia, "true")
	r.Header.Set(HeaderExceptOnceProps, "activity")
	ctx := i.WithMergeProp(r.Context(), "activity", func() any { return []string{"a"} })
	ctx = i.WithOnce(ctx, "activity")
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	err := i.Render(w, r, "test/component", nil)
	if err != nil {
		t.Error(err)
	}

	var page Page

	err = json.NewDecoder(w.Result().Body).Decode(&page)
	if err != nil {
		t.Error(err)
	}

	if _, ok := page.Props["activity"]; ok {
		t.Error("expected activity prop to be excluded when in except-once")
	}

	if prop, ok := page.OnceProps["activity"]; !ok || prop.Prop != "activity" {
		t.Errorf("expected onceProps metadata to remain, got: %v", page.OnceProps)
	}
}

func TestRenderWithFlashProp(t *testing.T) {
	i := New("http://inertia-go.test", "", "")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderInertia, "true")
	ctx := i.WithFlashProp(r.Context(), map[string]any{"success": "created"})
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	err := i.Render(w, r, "test/component", nil)
	if err != nil {
		t.Error(err)
	}

	var page Page

	err = json.NewDecoder(w.Result().Body).Decode(&page)
	if err != nil {
		t.Error(err)
	}

	if page.Flash == nil {
		t.Error("expected flash to be set")
	}

	if page.Flash["success"] != "created" {
		t.Errorf("expected: created, got: %v", page.Flash["success"])
	}
}

func TestRenderWithPartialExcept(t *testing.T) {
	i := New("http://inertia-go.test", "", "")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderInertia, "true")
	r.Header.Set(HeaderPartialComponent, "test/component")
	r.Header.Set(HeaderPartialExcept, "secret")
	w := httptest.NewRecorder()

	err := i.Render(w, r, "test/component", map[string]any{
		"title":  "Test",
		"secret": "hidden",
	})
	if err != nil {
		t.Error(err)
	}

	var page Page

	err = json.NewDecoder(w.Result().Body).Decode(&page)
	if err != nil {
		t.Error(err)
	}

	if _, ok := page.Props["secret"]; ok {
		t.Error("expected excluded prop to be absent")
	}

	if page.Props["title"] != "Test" {
		t.Errorf("expected: Test, got: %v", page.Props["title"])
	}
}

func TestRenderWithClearHistory(t *testing.T) {
	i := New("http://inertia-go.test", "", "")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderInertia, "true")
	ctx := i.WithClearHistory(r.Context())
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	err := i.Render(w, r, "test/component", nil)
	if err != nil {
		t.Error(err)
	}

	var page Page

	err = json.NewDecoder(w.Result().Body).Decode(&page)
	if err != nil {
		t.Error(err)
	}

	if !page.ClearHistory {
		t.Error("expected clearHistory to be true")
	}
}

func TestRenderWithEncryptHistory(t *testing.T) {
	i := New("http://inertia-go.test", "", "")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderInertia, "true")
	ctx := i.WithEncryptHistory(r.Context())
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	err := i.Render(w, r, "test/component", nil)
	if err != nil {
		t.Error(err)
	}

	var page Page

	err = json.NewDecoder(w.Result().Body).Decode(&page)
	if err != nil {
		t.Error(err)
	}

	if !page.EncryptHistory {
		t.Error("expected encryptHistory to be true")
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
