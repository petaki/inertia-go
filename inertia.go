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
	return i.ssrURL != "" && i.ssrClient != nil
}

// EnableSsr function.
func (i *Inertia) EnableSsr(ssrURL string) {
	i.ssrURL = ssrURL
	i.ssrClient = &http.Client{}
}

// EnableSsrWithDefault function.
func (i *Inertia) EnableSsrWithDefault() {
	i.EnableSsr("http://127.0.0.1:13714")
}

// DisableSsr function.
func (i *Inertia) DisableSsr() {
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
func (i *Inertia) WithDeferredProp(ctx context.Context, key string, value func() any, group string) context.Context {
	if group == "" {
		group = "default"
	}

	props := ctx.Value(ContextKeyDeferredProps)

	if props != nil {
		props, ok := props.(map[string]ContextEntryDeferredProp)
		if ok {
			props[key] = ContextEntryDeferredProp{Group: group, Value: value}

			return context.WithValue(ctx, ContextKeyDeferredProps, props)
		}
	}

	return context.WithValue(ctx, ContextKeyDeferredProps, map[string]ContextEntryDeferredProp{
		key: {Group: group, Value: value},
	})
}

// WithMergeProp function.
func (i *Inertia) WithMergeProp(ctx context.Context, key string, value func() any) context.Context {
	return i.withLazyProp(ctx, ContextKeyMergeProps, key, ContextEntryLazyProp{Value: value})
}

// WithDeepMergeProp function.
func (i *Inertia) WithDeepMergeProp(ctx context.Context, key string, value func() any) context.Context {
	return i.withLazyProp(ctx, ContextKeyDeepMergeProps, key, ContextEntryLazyProp{Value: value})
}

// WithPrependProp function.
func (i *Inertia) WithPrependProp(ctx context.Context, key string, value func() any) context.Context {
	return i.withLazyProp(ctx, ContextKeyPrependProps, key, ContextEntryLazyProp{Value: value})
}

// WithOptionalProp function.
func (i *Inertia) WithOptionalProp(ctx context.Context, key string, value func() any) context.Context {
	return i.withLazyProp(ctx, ContextKeyOptionalProps, key, ContextEntryLazyProp{Value: value})
}

// WithAlwaysProp function.
func (i *Inertia) WithAlwaysProp(ctx context.Context, key string, value func() any) context.Context {
	return i.withLazyProp(ctx, ContextKeyAlwaysProps, key, ContextEntryLazyProp{Value: value})
}

// WithOnceProp function.
func (i *Inertia) WithOnceProp(ctx context.Context, key string, value func() any) context.Context {
	return i.withLazyProp(ctx, ContextKeyOnceProps, key, ContextEntryLazyProp{Value: value})
}

func (i *Inertia) withLazyProp(ctx context.Context, ctxKey contextKey, key string, entry ContextEntryLazyProp) context.Context {
	props := ctx.Value(ctxKey)

	if props != nil {
		props, ok := props.(map[string]ContextEntryLazyProp)
		if ok {
			props[key] = entry

			return context.WithValue(ctx, ctxKey, props)
		}
	}

	return context.WithValue(ctx, ctxKey, map[string]ContextEntryLazyProp{
		key: entry,
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

	only := make(map[string]struct{})
	except := make(map[string]struct{})
	exceptOnce := make(map[string]struct{})
	isPartial := false

	if r.Header.Get(HeaderPartialComponent) == component {
		if partial := r.Header.Get(HeaderPartialOnly); partial != "" {
			isPartial = true

			for value := range strings.SplitSeq(partial, ",") {
				only[value] = struct{}{}
			}
		}

		if partialExcept := r.Header.Get(HeaderPartialExcept); partialExcept != "" {
			isPartial = true

			for value := range strings.SplitSeq(partialExcept, ",") {
				except[value] = struct{}{}
			}
		}
	}

	if exceptOnceHeader := r.Header.Get(HeaderExceptOnceProps); exceptOnceHeader != "" {
		for value := range strings.SplitSeq(exceptOnceHeader, ",") {
			exceptOnce[value] = struct{}{}
		}
	}

	page := &Page{
		Component: component,
		Props:     make(map[string]any),
		URL:       r.RequestURI,
		Version:   i.version,
	}

	if encryptHistory, ok := r.Context().Value(ContextKeyEncryptHistory).(bool); ok {
		page.EncryptHistory = encryptHistory
	}

	if clearHistory, ok := r.Context().Value(ContextKeyClearHistory).(bool); ok {
		page.ClearHistory = clearHistory
	}

	allProps := make(map[string]any)

	maps.Copy(allProps, i.sharedProps)

	contextProps := r.Context().Value(ContextKeyProps)

	if contextProps != nil {
		contextProps, ok := contextProps.(map[string]any)
		if !ok {
			return ErrInvalidContextProps
		}

		maps.Copy(allProps, contextProps)
	}

	maps.Copy(allProps, props)

	deferredProps, _ := r.Context().Value(ContextKeyDeferredProps).(map[string]ContextEntryDeferredProp)
	mergeProps, _ := r.Context().Value(ContextKeyMergeProps).(map[string]ContextEntryLazyProp)
	prependProps, _ := r.Context().Value(ContextKeyPrependProps).(map[string]ContextEntryLazyProp)
	deepMergeProps, _ := r.Context().Value(ContextKeyDeepMergeProps).(map[string]ContextEntryLazyProp)
	optionalProps, _ := r.Context().Value(ContextKeyOptionalProps).(map[string]ContextEntryLazyProp)
	alwaysProps, _ := r.Context().Value(ContextKeyAlwaysProps).(map[string]ContextEntryLazyProp)
	onceProps, _ := r.Context().Value(ContextKeyOnceProps).(map[string]ContextEntryLazyProp)

	for key, value := range allProps {
		if _, ok := except[key]; ok {
			continue
		}

		if _, ok := only[key]; len(only) == 0 || ok {
			page.Props[key] = value
		}
	}

	for key, entry := range deferredProps {
		if _, ok := except[key]; ok {
			continue
		}

		if isPartial {
			if _, ok := only[key]; len(only) == 0 || ok {
				page.Props[key] = entry.Value()
			}
		} else {
			if page.DeferredProps == nil {
				page.DeferredProps = make(map[string][]string)
			}

			page.DeferredProps[entry.Group] = append(page.DeferredProps[entry.Group], key)
		}
	}

	for key, value := range mergeProps {
		if _, ok := except[key]; ok {
			continue
		}

		if _, ok := only[key]; len(only) == 0 || ok {
			page.Props[key] = value.Value()
			page.MergeProps = append(page.MergeProps, key)
		}
	}

	for key, value := range prependProps {
		if _, ok := except[key]; ok {
			continue
		}

		if _, ok := only[key]; len(only) == 0 || ok {
			page.Props[key] = value.Value()
			page.PrependProps = append(page.PrependProps, key)
		}
	}

	for key, value := range deepMergeProps {
		if _, ok := except[key]; ok {
			continue
		}

		if _, ok := only[key]; len(only) == 0 || ok {
			page.Props[key] = value.Value()
			page.DeepMergeProps = append(page.DeepMergeProps, key)
		}
	}

	for key, value := range optionalProps {
		if _, ok := except[key]; ok {
			continue
		}

		if isPartial {
			if _, ok := only[key]; ok {
				page.Props[key] = value.Value()
			}
		}
	}

	for key, value := range alwaysProps {
		if _, ok := except[key]; ok {
			continue
		}

		page.Props[key] = value.Value()
	}

	for key, value := range onceProps {
		if _, ok := except[key]; ok {
			continue
		}

		if _, ok := exceptOnce[key]; !ok {
			if _, ok := only[key]; len(only) == 0 || ok {
				page.Props[key] = value.Value()
			}
		}
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

	if i.IsSsrEnabled() {
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

func (i *Inertia) createViewData(r *http.Request) (map[string]any, error) {
	viewData := make(map[string]any, len(i.sharedViewData))
	maps.Copy(viewData, i.sharedViewData)

	contextViewData := r.Context().Value(ContextKeyViewData)

	if contextViewData != nil {
		contextViewData, ok := contextViewData.(map[string]any)
		if !ok {
			return nil, ErrInvalidContextViewData
		}

		maps.Copy(viewData, contextViewData)
	}

	return viewData, nil
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
