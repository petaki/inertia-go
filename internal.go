package inertia

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"maps"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
)

func (i *Inertia) createRuntime(r *http.Request, component string, props map[string]any) *runtime {
	i.mu.RLock()
	defer i.mu.RUnlock()

	rt := &runtime{
		isSsrEnabled:   i.isSsrEnabled(),
		version:        i.version,
		ssrURL:         i.ssrURL,
		ssrClient:      i.ssrClient,
		props:          props,
		sharedProps:    maps.Clone(i.sharedProps),
		sharedViewData: maps.Clone(i.sharedViewData),
		only:           make(map[string]struct{}),
		except:         make(map[string]struct{}),
		exceptOnce:     make(map[string]struct{}),
		once:           make(map[string]struct{}),
		reset:          make(map[string]struct{}),
	}

	if r.Header.Get(HeaderPartialComponent) == component {
		if partial := r.Header.Get(HeaderPartialOnly); partial != "" {
			rt.isPartial = true

			for value := range strings.SplitSeq(partial, ",") {
				rt.only[value] = struct{}{}
			}
		}

		if partialExcept := r.Header.Get(HeaderPartialExcept); partialExcept != "" {
			rt.isPartial = true

			for value := range strings.SplitSeq(partialExcept, ",") {
				rt.except[value] = struct{}{}
			}
		}
	}

	if exceptOnceHeader := r.Header.Get(HeaderExceptOnceProps); exceptOnceHeader != "" {
		for value := range strings.SplitSeq(exceptOnceHeader, ",") {
			rt.exceptOnce[value] = struct{}{}
		}
	}

	if resetHeader := r.Header.Get(HeaderReset); resetHeader != "" {
		for value := range strings.SplitSeq(resetHeader, ",") {
			rt.reset[value] = struct{}{}
		}
	}

	return rt
}

func (i *Inertia) createOnceMetadata(r *http.Request, rt *runtime, page *Page) error {
	onceProps, err := contextGet[map[string]func() any](r.Context(), contextKeyOnceProps)
	if err != nil {
		return err
	}

	onceModifiers, err := contextGet[map[string]OncePageProp](r.Context(), contextKeyOnce)
	if err != nil {
		return err
	}

	metadata := make(map[string]OncePageProp, len(onceProps)+len(onceModifiers))

	for key := range onceProps {
		metadata[key] = OncePageProp{Prop: key}
	}

	maps.Copy(metadata, onceModifiers)

	for key, prop := range metadata {
		rt.once[key] = struct{}{}

		_, ok := rt.except[key]
		if ok {
			continue
		}

		_, ok = rt.only[key]
		if len(rt.only) != 0 && !ok {
			continue
		}

		if page.OnceProps == nil {
			page.OnceProps = make(map[string]OncePageProp)
		}

		page.OnceProps[key] = prop
	}

	return nil
}

func (i *Inertia) createBaseProps(r *http.Request, rt *runtime, page *Page) error {
	contextProps, err := contextGet[map[string]any](r.Context(), contextKeyProps)
	if err != nil {
		return err
	}

	baseProps := make(map[string]any)
	maps.Copy(baseProps, rt.sharedProps)
	maps.Copy(baseProps, contextProps)
	maps.Copy(baseProps, rt.props)

	for key, value := range baseProps {
		if rt.isCachedOnce(key) {
			continue
		}

		if key != "errors" {
			_, ok := rt.except[key]
			if ok {
				continue
			}

			_, ok = rt.only[key]
			if len(rt.only) != 0 && !ok {
				continue
			}
		}

		page.Props[key] = value
	}

	return nil
}

func (i *Inertia) createOptionalProps(r *http.Request, rt *runtime, page *Page) error {
	return i.createMainProps(r, rt, page, contextKeyOptionalProps)
}

func (i *Inertia) createAlwaysProps(r *http.Request, rt *runtime, page *Page) error {
	return i.createMainProps(r, rt, page, contextKeyAlwaysProps)
}

func (i *Inertia) createDeferredProps(r *http.Request, rt *runtime, page *Page) error {
	deferredProps, err := contextGet[map[string]contextDeferredProp](r.Context(), contextKeyDeferredProps)
	if err != nil {
		return err
	}

	for key, value := range deferredProps {
		_, ok := rt.except[key]
		if ok || rt.isCachedOnce(key) {
			continue
		}

		if rt.isPartial {
			_, ok = rt.only[key]
			if len(rt.only) == 0 || ok {
				v, err := value.Value()
				if err != nil {
					page.RescuedProps = append(page.RescuedProps, key)

					continue
				}

				page.Props[key] = v
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

func (i *Inertia) createScrollProps(r *http.Request, rt *runtime, page *Page) error {
	scrollProps, err := contextGet[map[string]ScrollPageProp](r.Context(), contextKeyScrollProp)
	if err != nil {
		return err
	}

	for key, prop := range scrollProps {
		_, ok := rt.except[key]
		if ok {
			continue
		}

		_, ok = rt.only[key]
		if len(rt.only) == 0 || ok {
			if page.ScrollProps == nil {
				page.ScrollProps = make(map[string]ScrollPageProp)
			}

			_, ok = rt.reset[key]
			if ok {
				prop.Reset = true
			}

			page.ScrollProps[key] = prop
		}
	}

	return nil
}

func (i *Inertia) createOnceProps(r *http.Request, rt *runtime, page *Page) error {
	return i.createMainProps(r, rt, page, contextKeyOnceProps)
}

func (i *Inertia) createErrorProps(r *http.Request, rt *runtime, page *Page) error {
	inlineErrors, ok := rt.props["errors"]
	if ok {
		if inlineErrors == nil {
			page.Props["errors"] = make(map[string]any)
		}

		return nil
	}

	contextErrors, err := contextGet[map[string]any](r.Context(), contextKeyErrors)
	if err != nil {
		return err
	}

	if contextErrors != nil {
		page.Props["errors"] = contextErrors
	} else if page.Props["errors"] == nil {
		page.Props["errors"] = make(map[string]any)
	}

	return nil
}

func (i *Inertia) createSharedProps(_ *http.Request, rt *runtime, page *Page) error {
	if len(rt.sharedProps) > 0 {
		page.SharedProps = slices.Sorted(maps.Keys(rt.sharedProps))
	}

	return nil
}

func (i *Inertia) createPageSettings(r *http.Request, _ *runtime, page *Page) error {
	flash, ok := r.Context().Value(contextKeyFlash).(map[string]any)
	if ok {
		page.Flash = flash
	}

	clearHistory, ok := r.Context().Value(contextKeyClearHistory).(bool)
	if ok {
		page.ClearHistory = clearHistory
	}

	encryptHistory, ok := r.Context().Value(contextKeyEncryptHistory).(bool)
	if ok {
		page.EncryptHistory = encryptHistory
	}

	preserveFragment, ok := r.Context().Value(contextKeyPreserveFragment).(bool)
	if ok {
		page.PreserveFragment = preserveFragment
	}

	return nil
}

func (i *Inertia) createVaryHeader(w http.ResponseWriter) {
	if w.Header().Get("Vary") == "" {
		w.Header().Set("Vary", HeaderInertia)
	} else {
		w.Header().Add("Vary", HeaderInertia)
	}
}

func (i *Inertia) createRootTemplate() (*template.Template, error) {
	i.tplMu.Lock()
	defer i.tplMu.Unlock()

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

func (i *Inertia) createViewData(r *http.Request, rt *runtime) (map[string]any, error) {
	contextViewData, err := contextGet[map[string]any](r.Context(), contextKeyViewData)
	if err != nil {
		return nil, err
	}

	viewData := rt.sharedViewData
	maps.Copy(viewData, contextViewData)

	return viewData, nil
}

func (i *Inertia) isSsrEnabled() bool {
	return i.ssrURL != "" && i.ssrClient != nil
}

func (i *Inertia) ssr(ctx context.Context, rt *runtime, page *Page) (*Ssr, error) {
	body, err := json.Marshal(page)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		rt.ssrURL,
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := rt.ssrClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("%w: %d", ErrBadSsrStatusCode, resp.StatusCode)
	}

	var ssr Ssr

	err = json.NewDecoder(resp.Body).Decode(&ssr)
	if err != nil {
		return nil, err
	}

	return &ssr, nil
}

func (i *Inertia) createMainProps(r *http.Request, rt *runtime, page *Page, key contextKey) error {
	props, err := contextGet[map[string]func() any](r.Context(), key)
	if err != nil {
		return err
	}

	for k, value := range props {
		if rt.isCachedOnce(k) {
			continue
		}

		if key == contextKeyAlwaysProps {
			page.Props[k] = value()

			continue
		}

		_, ok := rt.except[k]
		if ok {
			continue
		}

		_, ok = rt.only[k]
		if len(rt.only) != 0 && !ok {
			continue
		}

		switch key {
		case contextKeyOptionalProps:
			if rt.isPartial {
				page.Props[k] = value()
			}
		case contextKeyOnceProps:
			page.Props[k] = value()
		}
	}

	return nil
}

func (i *Inertia) createMergeableProps(r *http.Request, rt *runtime, page *Page, key contextKey) error {
	props, err := contextGet[map[string]contextMergeableProp](r.Context(), key)
	if err != nil {
		return err
	}

	for k, prop := range props {
		_, ok := rt.except[k]
		if ok || rt.isCachedOnce(k) {
			continue
		}

		_, ok = rt.only[k]
		if len(rt.only) == 0 || ok {
			page.Props[k] = prop.Value()

			_, resetting := rt.reset[k]
			if !resetting {
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
	}

	return nil
}
