package inertia

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"
	"maps"
	"net/http"
	"path/filepath"
)

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

func (i *Inertia) createScrollProps(r *http.Request, rt *runtime, page *Page) error {
	scrollProps, err := contextGet[map[string]ScrollPageProp](r.Context(), contextKeyScrollProp)
	if err != nil {
		return err
	}

	for key, prop := range scrollProps {
		if page.ScrollProps == nil {
			page.ScrollProps = make(map[string]ScrollPageProp)
		}

		_, ok := rt.reset[key]
		if ok {
			prop.Reset = true
		}

		page.ScrollProps[key] = prop
	}

	return nil
}

func (i *Inertia) createOnceProps(r *http.Request, rt *runtime, page *Page) error {
	return i.createMainProps(r, rt, page, contextKeyOnceProps)
}

func (i *Inertia) createOnceModifiers(r *http.Request, rt *runtime, page *Page) error {
	onceModifiers, err := contextGet[map[string]OncePageProp](r.Context(), contextKeyOnce)
	if err != nil {
		return err
	}

	for key, prop := range onceModifiers {
		if page.OnceProps == nil {
			page.OnceProps = make(map[string]OncePageProp)
		}

		page.OnceProps[key] = prop

		_, ok := rt.exceptOnce[key]
		if ok {
			delete(page.Props, key)
		}
	}

	return nil
}

func (i *Inertia) createFlashProps(r *http.Request, _ *runtime, page *Page) error {
	flash, ok := r.Context().Value(contextKeyFlashProp).(map[string]any)
	if ok {
		page.Flash = flash
	}

	return nil
}

func (i *Inertia) createMainProps(r *http.Request, rt *runtime, page *Page, key contextKey) error {
	props, err := contextGet[map[string]func() any](r.Context(), key)
	if err != nil {
		return err
	}

	for k, value := range props {
		_, ok := rt.except[k]
		if ok {
			continue
		}

		switch key {
		case contextKeyOptionalProps:
			if rt.isPartial {
				_, ok = rt.only[k]
				if ok {
					page.Props[k] = value()
				}
			}
		case contextKeyAlwaysProps:
			page.Props[k] = value()
		case contextKeyOnceProps:
			if page.OnceProps == nil {
				page.OnceProps = make(map[string]OncePageProp)
			}

			page.OnceProps[k] = OncePageProp{Prop: k}

			_, exceptOnce := rt.exceptOnce[k]
			if !exceptOnce {
				_, ok = rt.only[k]
				if len(rt.only) == 0 || ok {
					page.Props[k] = value()
				}
			}
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
		if ok {
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
		i.ssrURL,
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
