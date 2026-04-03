package inertia

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"
	"io/fs"
	"maps"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
)

// Inertia type.
type Inertia struct {
	mu             sync.RWMutex
	url            string
	rootTemplate   string
	version        string
	sharedProps    map[string]any
	sharedFuncMap  template.FuncMap
	sharedViewData map[string]any
	parsedTemplate *template.Template
	templateFS     fs.FS
	ssrURL         string
	ssrClient      *http.Client
}

// New function.
func New(url, rootTemplate, version string) *Inertia {
	return &Inertia{
		url:            url,
		rootTemplate:   rootTemplate,
		version:        version,
		sharedProps:    make(map[string]any),
		sharedFuncMap:  template.FuncMap{"marshal": marshal, "raw": raw},
		sharedViewData: make(map[string]any),
	}
}

// NewWithFS function.
func NewWithFS(url, rootTemplate, version string, templateFS fs.FS) *Inertia {
	i := New(url, rootTemplate, version)
	i.templateFS = templateFS

	return i
}

// IsSsrEnabled function.
func (i *Inertia) IsSsrEnabled() bool {
	i.mu.RLock()
	defer i.mu.RUnlock()

	return i.isSsrEnabled()
}

// EnableSsr function.
func (i *Inertia) EnableSsr(ssrURL string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.ssrURL = ssrURL
	i.ssrClient = &http.Client{}
}

// EnableSsrWithDefault function.
func (i *Inertia) EnableSsrWithDefault() {
	i.EnableSsr("http://127.0.0.1:13714")
}

// DisableSsr function.
func (i *Inertia) DisableSsr() {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.ssrURL = ""
	i.ssrClient = nil
}

// ShareFunc function.
func (i *Inertia) ShareFunc(key string, value any) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.sharedFuncMap[key] = value
	i.parsedTemplate = nil
}

// ShareViewData function.
func (i *Inertia) ShareViewData(key string, value any) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.sharedViewData[key] = value
}

// WithViewData function.
func (i *Inertia) WithViewData(ctx context.Context, key string, value any) context.Context {
	contextViewData := ctx.Value(ContextKeyViewData)

	if contextViewData != nil {
		contextViewData, ok := contextViewData.(map[string]any)
		if ok {
			contextViewData[key] = value

			return context.WithValue(ctx, ContextKeyViewData, contextViewData)
		}
	}

	return context.WithValue(ctx, ContextKeyViewData, map[string]any{
		key: value,
	})
}

// Share function.
func (i *Inertia) Share(key string, value any) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.sharedProps[key] = value
}

// WithProp function.
func (i *Inertia) WithProp(ctx context.Context, key string, value any) context.Context {
	contextProps := ctx.Value(ContextKeyProps)

	if contextProps != nil {
		contextProps, ok := contextProps.(map[string]any)
		if ok {
			contextProps[key] = value

			return context.WithValue(ctx, ContextKeyProps, contextProps)
		}
	}

	return context.WithValue(ctx, ContextKeyProps, map[string]any{
		key: value,
	})
}

// WithDeferredProp function.
func (i *Inertia) WithDeferredProp(ctx context.Context, key string, value func() any) context.Context {
	return i.WithDeferredGroupProp(ctx, key, value, "default")
}

// WithDeferredGroupProp function.
func (i *Inertia) WithDeferredGroupProp(ctx context.Context, key string, value func() any, group string) context.Context {
	props := ctx.Value(ContextKeyDeferredProps)

	if props != nil {
		props, ok := props.(map[string]ContextValueDeferredProp)
		if ok {
			props[key] = ContextValueDeferredProp{Group: group, Value: value}

			return context.WithValue(ctx, ContextKeyDeferredProps, props)
		}
	}

	return context.WithValue(ctx, ContextKeyDeferredProps, map[string]ContextValueDeferredProp{
		key: {Group: group, Value: value},
	})
}

// WithMergeProp function.
func (i *Inertia) WithMergeProp(ctx context.Context, key string, value func() any) context.Context {
	return i.withLazyProp(ctx, ContextKeyMergeProps, key, value)
}

// WithDeepMergeProp function.
func (i *Inertia) WithDeepMergeProp(ctx context.Context, key string, value func() any) context.Context {
	return i.withLazyProp(ctx, ContextKeyDeepMergeProps, key, value)
}

// WithPrependProp function.
func (i *Inertia) WithPrependProp(ctx context.Context, key string, value func() any) context.Context {
	return i.withLazyProp(ctx, ContextKeyPrependProps, key, value)
}

// WithOptionalProp function.
func (i *Inertia) WithOptionalProp(ctx context.Context, key string, value func() any) context.Context {
	return i.withLazyProp(ctx, ContextKeyOptionalProps, key, value)
}

// WithAlwaysProp function.
func (i *Inertia) WithAlwaysProp(ctx context.Context, key string, value func() any) context.Context {
	return i.withLazyProp(ctx, ContextKeyAlwaysProps, key, value)
}

// WithOnceProp function.
func (i *Inertia) WithOnceProp(ctx context.Context, key string, value func() any) context.Context {
	return i.withLazyProp(ctx, ContextKeyOnceProps, key, value)
}

func (i *Inertia) withLazyProp(ctx context.Context, ctxKey contextKey, key string, value func() any) context.Context {
	props := ctx.Value(ctxKey)

	if props != nil {
		props, ok := props.(map[string]func() any)
		if ok {
			props[key] = value

			return context.WithValue(ctx, ctxKey, props)
		}
	}

	return context.WithValue(ctx, ctxKey, map[string]func() any{
		key: value,
	})
}

// WithClearHistory function.
func (i *Inertia) WithClearHistory(ctx context.Context) context.Context {
	return context.WithValue(ctx, ContextKeyClearHistory, true)
}

// WithEncryptHistory function.
func (i *Inertia) WithEncryptHistory(ctx context.Context) context.Context {
	return context.WithValue(ctx, ContextKeyEncryptHistory, true)
}

// Render function.
func (i *Inertia) Render(w http.ResponseWriter, r *http.Request, component string, props map[string]any) error {
	i.mu.RLock()
	defer i.mu.RUnlock()

	rt := newRuntime(r, component, props)

	page := &Page{
		Component: component,
		Props:     make(map[string]any),
		URL:       r.RequestURI,
		Version:   i.version,
	}

	for _, create := range []func(*http.Request, *runtime, *Page) error{
		i.createBaseProps,
		i.createDeferredProps,
		i.createMergeProps,
		i.createDeepMergeProps,
		i.createPrependProps,
		i.createOptionalProps,
		i.createAlwaysProps,
		i.createOnceProps,
	} {
		err := create(r, rt, page)
		if err != nil {
			return err
		}
	}

	clearHistory, ok := r.Context().Value(ContextKeyClearHistory).(bool)
	if ok {
		page.ClearHistory = clearHistory
	}

	encryptHistory, ok := r.Context().Value(ContextKeyEncryptHistory).(bool)
	if ok {
		page.EncryptHistory = encryptHistory
	}

	if r.Header.Get(HeaderInertia) != "" {
		js, err := json.Marshal(page)
		if err != nil {
			return err
		}

		if w.Header().Get("Vary") == "" {
			w.Header().Set("Vary", HeaderInertia)
		} else {
			w.Header().Add("Vary", HeaderInertia)
		}

		w.Header().Set(HeaderInertia, "true")
		w.Header().Set("Content-Type", "application/json")

		_, err = w.Write(js)

		return err
	}

	rootTemplate, err := i.createRootTemplate()
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/html")

	viewData, err := i.createViewData(r)
	if err != nil {
		return err
	}

	viewData["page"] = page

	if i.isSsrEnabled() {
		ssr, err := i.ssr(r.Context(), page)
		if err != nil {
			return err
		}

		viewData["ssr"] = ssr
	} else {
		viewData["ssr"] = nil
	}

	return rootTemplate.Execute(w, viewData)
}

// Location function.
func (i *Inertia) Location(w http.ResponseWriter, r *http.Request, url string) {
	if r.Header.Get(HeaderInertia) != "" {
		w.Header().Set(HeaderLocation, url)
		w.WriteHeader(http.StatusConflict)
	} else {
		http.Redirect(w, r, url, http.StatusFound)
	}
}

func (i *Inertia) createRootTemplate() (*template.Template, error) {
	if i.parsedTemplate != nil {
		return i.parsedTemplate, nil
	}

	ts := template.New(filepath.Base(i.rootTemplate)).Funcs(i.sharedFuncMap)

	var tpl *template.Template
	var err error

	if i.templateFS != nil {
		tpl, err = ts.ParseFS(i.templateFS, i.rootTemplate)
	} else {
		tpl, err = ts.ParseFiles(i.rootTemplate)
	}

	if err != nil {
		return nil, err
	}

	i.parsedTemplate = tpl

	return i.parsedTemplate, nil
}

func (i *Inertia) createBaseProps(r *http.Request, rt *runtime, page *Page) error {
	contextProps, err := contextValue[map[string]any](r.Context(), ContextKeyProps)
	if err != nil {
		return err
	}

	baseProps := make(map[string]any)
	maps.Copy(baseProps, i.sharedProps)
	maps.Copy(baseProps, contextProps)
	maps.Copy(baseProps, rt.props)

	for key, value := range baseProps {
		_, ok := rt.except[key]
		if ok {
			continue
		}

		_, ok = rt.only[key]
		if len(rt.only) == 0 || ok {
			page.Props[key] = value
		}
	}

	return nil
}

func (i *Inertia) createDeferredProps(r *http.Request, rt *runtime, page *Page) error {
	deferredProps, err := contextValue[map[string]ContextValueDeferredProp](r.Context(), ContextKeyDeferredProps)
	if err != nil {
		return err
	}

	for key, value := range deferredProps {
		_, ok := rt.except[key]
		if ok {
			continue
		}

		if rt.isPartial {
			_, ok = rt.only[key]
			if len(rt.only) == 0 || ok {
				page.Props[key] = value.Value()
			}
		} else {
			if page.DeferredProps == nil {
				page.DeferredProps = make(map[string][]string)
			}

			page.DeferredProps[value.Group] = append(page.DeferredProps[value.Group], key)
		}
	}

	return nil
}

func (i *Inertia) createMergeProps(r *http.Request, rt *runtime, page *Page) error {
	mergeProps, err := contextValue[map[string]func() any](r.Context(), ContextKeyMergeProps)
	if err != nil {
		return err
	}

	for key, value := range mergeProps {
		_, ok := rt.except[key]
		if ok {
			continue
		}

		_, ok = rt.only[key]
		if len(rt.only) == 0 || ok {
			page.Props[key] = value()
			page.MergeProps = append(page.MergeProps, key)
		}
	}

	return nil
}

func (i *Inertia) createDeepMergeProps(r *http.Request, rt *runtime, page *Page) error {
	deepMergeProps, err := contextValue[map[string]func() any](r.Context(), ContextKeyDeepMergeProps)
	if err != nil {
		return err
	}

	for key, value := range deepMergeProps {
		_, ok := rt.except[key]
		if ok {
			continue
		}

		_, ok = rt.only[key]
		if len(rt.only) == 0 || ok {
			page.Props[key] = value()
			page.DeepMergeProps = append(page.DeepMergeProps, key)
		}
	}

	return nil
}

func (i *Inertia) createPrependProps(r *http.Request, rt *runtime, page *Page) error {
	prependProps, err := contextValue[map[string]func() any](r.Context(), ContextKeyPrependProps)
	if err != nil {
		return err
	}

	for key, value := range prependProps {
		_, ok := rt.except[key]
		if ok {
			continue
		}

		_, ok = rt.only[key]
		if len(rt.only) == 0 || ok {
			page.Props[key] = value()
			page.PrependProps = append(page.PrependProps, key)
		}
	}

	return nil
}

func (i *Inertia) createOptionalProps(r *http.Request, rt *runtime, page *Page) error {
	optionalProps, err := contextValue[map[string]func() any](r.Context(), ContextKeyOptionalProps)
	if err != nil {
		return err
	}

	for key, value := range optionalProps {
		_, ok := rt.except[key]
		if ok {
			continue
		}

		if rt.isPartial {
			_, ok = rt.only[key]
			if ok {
				page.Props[key] = value()
			}
		}
	}

	return nil
}

func (i *Inertia) createAlwaysProps(r *http.Request, rt *runtime, page *Page) error {
	alwaysProps, err := contextValue[map[string]func() any](r.Context(), ContextKeyAlwaysProps)
	if err != nil {
		return err
	}

	for key, value := range alwaysProps {
		_, ok := rt.except[key]
		if ok {
			continue
		}

		page.Props[key] = value()
	}

	return nil
}

func (i *Inertia) createOnceProps(r *http.Request, rt *runtime, page *Page) error {
	onceProps, err := contextValue[map[string]func() any](r.Context(), ContextKeyOnceProps)
	if err != nil {
		return err
	}

	for key, value := range onceProps {
		_, ok := rt.except[key]
		if ok {
			continue
		}

		_, ok = rt.exceptOnce[key]
		if !ok {
			_, ok = rt.only[key]
			if len(rt.only) == 0 || ok {
				page.Props[key] = value()
			}
		}
	}

	return nil
}

func (i *Inertia) createViewData(r *http.Request) (map[string]any, error) {
	contextViewData, err := contextValue[map[string]any](r.Context(), ContextKeyViewData)
	if err != nil {
		return nil, err
	}

	viewData := make(map[string]any, len(i.sharedViewData))
	maps.Copy(viewData, i.sharedViewData)
	maps.Copy(viewData, contextViewData)

	return viewData, nil
}

func (i *Inertia) isSsrEnabled() bool {
	return i.ssrURL != "" && i.ssrClient != nil
}

func (i *Inertia) ssr(ctx context.Context, page *Page) (*Ssr, error) {
	body, err := json.Marshal(page)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		strings.ReplaceAll(i.ssrURL, "/render", "")+"/render",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := i.ssrClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, ErrBadSsrStatusCode
	}

	var ssr Ssr

	err = json.NewDecoder(resp.Body).Decode(&ssr)
	if err != nil {
		return nil, err
	}

	return &ssr, nil
}
