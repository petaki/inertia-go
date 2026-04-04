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
	i.EnableSsr("http://127.0.0.1:13714", client...)
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

// WithOptionalProp function.
func (i *Inertia) WithOptionalProp(ctx context.Context, key string, value func() any) context.Context {
	return contextSet(ctx, contextKeyOptionalProps, key, value)
}

// WithAlwaysProp function.
func (i *Inertia) WithAlwaysProp(ctx context.Context, key string, value func() any) context.Context {
	return contextSet(ctx, contextKeyAlwaysProps, key, value)
}

// WithOnceProp function.
func (i *Inertia) WithOnceProp(ctx context.Context, key string, value func() any) context.Context {
	return contextSet(ctx, contextKeyOnceProps, key, value)
}

// WithClearHistory function.
func (i *Inertia) WithClearHistory(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKeyClearHistory, true)
}

// WithEncryptHistory function.
func (i *Inertia) WithEncryptHistory(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKeyEncryptHistory, true)
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
	contextProps, err := contextGet[map[string]any](r.Context(), contextKeyProps)
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
	deferredProps, err := contextGet[map[string]contextDeferredProp](r.Context(), contextKeyDeferredProps)
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
	return i.createMergeableProps(r, rt, page, contextKeyMergeProps)
}

func (i *Inertia) createDeepMergeProps(r *http.Request, rt *runtime, page *Page) error {
	return i.createMergeableProps(r, rt, page, contextKeyDeepMergeProps)
}

func (i *Inertia) createPrependProps(r *http.Request, rt *runtime, page *Page) error {
	return i.createMergeableProps(r, rt, page, contextKeyPrependProps)
}

func (i *Inertia) createMergeableProps(r *http.Request, rt *runtime, page *Page, key contextKey) error {
	props, err := contextGet[map[string]contextMergeableProp](r.Context(), key)
	if err != nil {
		return err
	}

	for k, prop := range props {
		_, ok := rt.except[k]
		if ok {
			continue
		}

		_, ok = rt.only[k]
		if len(rt.only) == 0 || ok {
			page.Props[k] = prop.Value()

			switch key {
			case contextKeyMergeProps:
				page.MergeProps = append(page.MergeProps, k)
			case contextKeyDeepMergeProps:
				page.DeepMergeProps = append(page.DeepMergeProps, k)
			case contextKeyPrependProps:
				page.PrependProps = append(page.PrependProps, k)
			}

			for _, m := range prop.MatchOn {
				page.MatchPropsOn = append(page.MatchPropsOn, k+"."+m)
			}
		}
	}

	return nil
}

func (i *Inertia) createOptionalProps(r *http.Request, rt *runtime, page *Page) error {
	optionalProps, err := contextGet[map[string]func() any](r.Context(), contextKeyOptionalProps)
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
	alwaysProps, err := contextGet[map[string]func() any](r.Context(), contextKeyAlwaysProps)
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
	onceProps, err := contextGet[map[string]func() any](r.Context(), contextKeyOnceProps)
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
	contextViewData, err := contextGet[map[string]any](r.Context(), contextKeyViewData)
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
