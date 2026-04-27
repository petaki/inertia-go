package inertia

import (
	"context"
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"
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
func New(url, rootTemplate, version string, templateFS ...fs.FS) *Inertia {
	i := &Inertia{
		url:            url,
		rootTemplate:   rootTemplate,
		version:        version,
		sharedProps:    make(map[string]any),
		sharedFuncMap:  template.FuncMap{"marshal": marshal, "raw": raw},
		sharedViewData: make(map[string]any),
	}

	if len(templateFS) > 0 && templateFS[0] != nil {
		i.templateFS = templateFS[0]
	}

	return i
}

// IsSsrEnabled function.
func (i *Inertia) IsSsrEnabled() bool {
	i.mu.RLock()
	defer i.mu.RUnlock()

	return i.isSsrEnabled()
}

// EnableSsr function.
func (i *Inertia) EnableSsr(ssrURL string, client ...*http.Client) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.ssrURL = ssrURL

	if len(client) > 0 && client[0] != nil {
		i.ssrClient = client[0]
	} else {
		i.ssrClient = &http.Client{}
	}
}

// EnableSsrWithDefault function.
func (i *Inertia) EnableSsrWithDefault(client ...*http.Client) {
	i.EnableSsr("http://127.0.0.1:13714/render", client...)
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
	return contextSet(ctx, contextKeyViewData, key, value)
}

// Share function.
func (i *Inertia) Share(key string, value any) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.sharedProps[key] = value
}

// WithProp function.
func (i *Inertia) WithProp(ctx context.Context, key string, value any) context.Context {
	return contextSet(ctx, contextKeyProps, key, value)
}

// WithOptionalProp function.
func (i *Inertia) WithOptionalProp(ctx context.Context, key string, value func() any) context.Context {
	return contextSet(ctx, contextKeyOptionalProps, key, value)
}

// WithAlwaysProp function.
func (i *Inertia) WithAlwaysProp(ctx context.Context, key string, value func() any) context.Context {
	return contextSet(ctx, contextKeyAlwaysProps, key, value)
}

// WithDeferredProp function.
func (i *Inertia) WithDeferredProp(ctx context.Context, key string, value func() any, group ...string) context.Context {
	g := "default"
	if len(group) > 0 && group[0] != "" {
		g = group[0]
	}

	return contextSet(ctx, contextKeyDeferredProps, key, contextDeferredProp{Group: g, Value: value})
}

// WithMergeProp function.
func (i *Inertia) WithMergeProp(ctx context.Context, key string, value func() any, matchOn ...string) context.Context {
	return contextSet(ctx, contextKeyMergeProps, key, contextMergeableProp{MatchOn: matchOn, Value: value})
}

// WithDeepMergeProp function.
func (i *Inertia) WithDeepMergeProp(ctx context.Context, key string, value func() any, matchOn ...string) context.Context {
	return contextSet(ctx, contextKeyDeepMergeProps, key, contextMergeableProp{MatchOn: matchOn, Value: value})
}

// WithPrependProp function.
func (i *Inertia) WithPrependProp(ctx context.Context, key string, value func() any, matchOn ...string) context.Context {
	return contextSet(ctx, contextKeyPrependProps, key, contextMergeableProp{MatchOn: matchOn, Value: value})
}

// WithScrollProp function.
func (i *Inertia) WithScrollProp(ctx context.Context, key string, prop ScrollPageProp) context.Context {
	return contextSet(ctx, contextKeyScrollProp, key, prop)
}

// WithOnceProp function.
func (i *Inertia) WithOnceProp(ctx context.Context, key string, value func() any) context.Context {
	return contextSet(ctx, contextKeyOnceProps, key, value)
}

// WithOnce function.
func (i *Inertia) WithOnce(ctx context.Context, key string, prop ...OncePageProp) context.Context {
	p := OncePageProp{}
	if len(prop) > 0 {
		p = prop[0]
	}

	p.Prop = key

	return contextSet(ctx, contextKeyOnce, key, p)
}

// WithErrorProp function.
func (i *Inertia) WithErrorProp(ctx context.Context, key string, value any) context.Context {
	return contextSet(ctx, contextKeyErrors, key, value)
}

// WithFlashProp function.
func (i *Inertia) WithFlashProp(ctx context.Context, data map[string]any) context.Context {
	return context.WithValue(ctx, contextKeyFlash, data)
}

// WithClearHistory function.
func (i *Inertia) WithClearHistory(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKeyClearHistory, true)
}

// WithEncryptHistory function.
func (i *Inertia) WithEncryptHistory(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKeyEncryptHistory, true)
}

// Middleware function.
func (i *Inertia) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(HeaderInertia) == "" {
			next.ServeHTTP(w, r)

			return
		}

		i.mu.RLock()
		version := i.version
		url := i.url
		i.mu.RUnlock()

		if r.Method == http.MethodGet && r.Header.Get(HeaderVersion) != version {
			w.Header().Set(HeaderLocation, url+r.RequestURI)
			w.WriteHeader(http.StatusConflict)

			return
		}

		next.ServeHTTP(w, r)
	})
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
		i.createOptionalProps,
		i.createAlwaysProps,
		i.createDeferredProps,
		i.createMergeProps,
		i.createDeepMergeProps,
		i.createPrependProps,
		i.createScrollProps,
		i.createOnceProps,
		i.createOnceModifiers,
		i.createErrorProps,
		i.createFlashProp,
	} {
		err := create(r, rt, page)
		if err != nil {
			return err
		}
	}

	clearHistory, ok := r.Context().Value(contextKeyClearHistory).(bool)
	if ok {
		page.ClearHistory = clearHistory
	}

	encryptHistory, ok := r.Context().Value(contextKeyEncryptHistory).(bool)
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
